package main

import (
    "bufio"
    "bytes"
    "context"
    "crypto/subtle"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "strings"
    "time"
)

type Config struct {
    BaseURL string
    APIKey  string
    Model   string
    BridgeToken string
}

type OllamaMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type OllamaChatRequest struct {
    Model   string          `json:"model"`
    Messages []OllamaMessage `json:"messages"`
    Stream  bool            `json:"stream"`
    Options map[string]any  `json:"options"`
}

type OllamaGenerateRequest struct {
    Model  string `json:"model"`
    Prompt string `json:"prompt"`
    Stream bool   `json:"stream"`
    Options map[string]any `json:"options"`
}

type OpenAIChatRequest struct {
    Model    string           `json:"model"`
    Messages []OllamaMessage  `json:"messages"`
    Stream   bool             `json:"stream"`
    Temperature *float64      `json:"temperature,omitempty"`
}

type OpenAIChatResponse struct {
    Choices []struct {
        Message struct {
            Role    string `json:"role"`
            Content string `json:"content"`
        } `json:"message"`
    } `json:"choices"`
}

type OpenAIStreamChunk struct {
    Choices []struct {
        Delta struct {
            Role    string `json:"role"`
            Content string `json:"content"`
        } `json:"delta"`
    } `json:"choices"`
}

func main() {
    cfg := Config{
        BaseURL: strings.TrimRight(os.Getenv("RUNPOD_OPENAI_BASE_URL"), "/"),
        APIKey:  os.Getenv("RUNPOD_OPENAI_API_KEY"),
        Model:   os.Getenv("OLLAMA_MODEL"),
        BridgeToken: os.Getenv("RUNPOD_BRIDGE_TOKEN"),
    }
    if cfg.Model == "" {
        cfg.Model = "llama3.2:1b"
    }
    if cfg.BaseURL == "" {
        log.Fatal("RUNPOD_OPENAI_BASE_URL is required")
    }

    mux := http.NewServeMux()
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    })
    mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
        writeJSON(w, map[string]any{"version": "0.1.0"})
    })
    mux.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) {
        now := time.Now().UTC().Format(time.RFC3339Nano)
        writeJSON(w, map[string]any{
            "models": []map[string]any{
                {
                    "name":        cfg.Model,
                    "model":       cfg.Model,
                    "modified_at": now,
                    "size":        0,
                    "digest":      "",
                    "details": map[string]any{
                        "format":             "gguf",
                        "family":             "unknown",
                        "families":           []string{"unknown"},
                        "parameter_size":     "",
                        "quantization_level": "",
                    },
                },
            },
        })
    })
    mux.HandleFunc("/api/show", func(w http.ResponseWriter, r *http.Request) {
        writeJSON(w, map[string]any{
            "name":       cfg.Model,
            "modelfile":  "",
            "parameters": "",
            "template":   "",
            "details": map[string]any{
                "format":             "gguf",
                "family":             "unknown",
                "families":           []string{"unknown"},
                "parameter_size":     "",
                "quantization_level": "",
            },
        })
    })
    mux.HandleFunc("/api/ps", func(w http.ResponseWriter, r *http.Request) {
        writeJSON(w, map[string]any{"models": []any{}})
    })
    mux.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
        if !authorizeBridgeRequest(r, cfg.BridgeToken) {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        if r.Method != http.MethodPost {
            w.WriteHeader(http.StatusMethodNotAllowed)
            return
        }
        var req OllamaChatRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "invalid json", http.StatusBadRequest)
            return
        }
        model := cfg.Model
        if req.Model != "" {
            model = req.Model
        }
        temp := extractTemperature(req.Options)
        openaiReq := OpenAIChatRequest{
            Model:       model,
            Messages:    req.Messages,
            Stream:      req.Stream,
            Temperature: temp,
        }
        if req.Stream {
            streamChat(w, r, cfg, openaiReq, model)
            return
        }
        content, err := doChatOnce(r.Context(), cfg, openaiReq)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadGateway)
            return
        }
        writeJSON(w, map[string]any{
            "model":      model,
            "created_at": time.Now().UTC().Format(time.RFC3339Nano),
            "message": map[string]any{
                "role":    "assistant",
                "content": content,
            },
            "done": true,
        })
    })
    mux.HandleFunc("/api/generate", func(w http.ResponseWriter, r *http.Request) {
        if !authorizeBridgeRequest(r, cfg.BridgeToken) {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        if r.Method != http.MethodPost {
            w.WriteHeader(http.StatusMethodNotAllowed)
            return
        }
        var req OllamaGenerateRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "invalid json", http.StatusBadRequest)
            return
        }
        model := cfg.Model
        if req.Model != "" {
            model = req.Model
        }
        temp := extractTemperature(req.Options)
        openaiReq := OpenAIChatRequest{
            Model:    model,
            Messages: []OllamaMessage{{Role: "user", Content: req.Prompt}},
            Stream:   req.Stream,
            Temperature: temp,
        }
        if req.Stream {
            streamChat(w, r, cfg, openaiReq, model)
            return
        }
        content, err := doChatOnce(r.Context(), cfg, openaiReq)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadGateway)
            return
        }
        writeJSON(w, map[string]any{
            "model":      model,
            "created_at": time.Now().UTC().Format(time.RFC3339Nano),
            "response":   content,
            "done":       true,
        })
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "8081"
    }
    log.Printf("runpod-ollama-bridge listening on :%s", port)
    if err := http.ListenAndServe(":"+port, mux); err != nil {
        log.Fatal(err)
    }
}

func authorizeBridgeRequest(r *http.Request, bridgeToken string) bool {
    if bridgeToken == "" {
        return false
    }

    token := strings.TrimSpace(r.Header.Get("X-Bridge-Token"))
    if token == "" {
        auth := strings.TrimSpace(r.Header.Get("Authorization"))
        if strings.HasPrefix(auth, "Bearer ") {
            token = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
        }
    }

    if token == "" {
        return false
    }

    return subtle.ConstantTimeCompare([]byte(token), []byte(bridgeToken)) == 1
}

func extractTemperature(opts map[string]any) *float64 {
    if opts == nil {
        return nil
    }
    if v, ok := opts["temperature"]; ok {
        switch t := v.(type) {
        case float64:
            return &t
        case int:
            f := float64(t)
            return &f
        }
    }
    return nil
}

func doChatOnce(ctx context.Context, cfg Config, req OpenAIChatRequest) (string, error) {
    body, _ := json.Marshal(req)
    httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIChatURL(cfg.BaseURL), bytes.NewReader(body))
    if err != nil {
        return "", err
    }
    httpReq.Header.Set("Content-Type", "application/json")
    if cfg.APIKey != "" {
        httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
    }
    client := &http.Client{Timeout: 60 * time.Second}
    resp, err := client.Do(httpReq)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 300 {
        b, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("openai error %d: %s", resp.StatusCode, string(b))
    }
    var decoded OpenAIChatResponse
    if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
        return "", err
    }
    if len(decoded.Choices) == 0 {
        return "", fmt.Errorf("openai response missing choices")
    }
    return decoded.Choices[0].Message.Content, nil
}

func streamChat(w http.ResponseWriter, r *http.Request, cfg Config, req OpenAIChatRequest, model string) {
    body, _ := json.Marshal(req)
    httpReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, openAIChatURL(cfg.BaseURL), bytes.NewReader(body))
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Accept", "text/event-stream")
    if cfg.APIKey != "" {
        httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
    }
    client := &http.Client{Timeout: 0}
    resp, err := client.Do(httpReq)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 300 {
        b, _ := io.ReadAll(resp.Body)
        http.Error(w, fmt.Sprintf("openai error %d: %s", resp.StatusCode, string(b)), http.StatusBadGateway)
        return
    }

    w.Header().Set("Content-Type", "application/x-ndjson")
    w.WriteHeader(http.StatusOK)
    flusher, _ := w.(http.Flusher)

    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || !strings.HasPrefix(line, "data:") {
            continue
        }
        data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
        if data == "[DONE]" {
            break
        }
        var chunk OpenAIStreamChunk
        if err := json.Unmarshal([]byte(data), &chunk); err != nil {
            continue
        }
        if len(chunk.Choices) == 0 {
            continue
        }
        delta := chunk.Choices[0].Delta
        if delta.Content == "" {
            continue
        }
        out := map[string]any{
            "model":      model,
            "created_at": time.Now().UTC().Format(time.RFC3339Nano),
            "message": map[string]any{
                "role":    "assistant",
                "content": delta.Content,
            },
            "done": false,
        }
        writeJSONLine(w, out)
        if flusher != nil {
            flusher.Flush()
        }
    }

    writeJSONLine(w, map[string]any{
        "model":      model,
        "created_at": time.Now().UTC().Format(time.RFC3339Nano),
        "message": map[string]any{
            "role":    "assistant",
            "content": "",
        },
        "done": true,
    })
    if flusher != nil {
        flusher.Flush()
    }
}

func writeJSON(w http.ResponseWriter, v any) {
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(v)
}

func writeJSONLine(w http.ResponseWriter, v any) {
    b, _ := json.Marshal(v)
    _, _ = w.Write(append(b, '\n'))
}

func openAIChatURL(base string) string {
    if strings.HasSuffix(base, "/v1") {
        return base + "/chat/completions"
    }
    return base + "/v1/chat/completions"
}
