package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sunspear/services"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/docker/docker/pkg/stdcopy"
)

type WSHandler struct {
	dockerService  *services.DockerService
	monitorService *services.MonitoringService
	upgrader       websocket.Upgrader
}

func NewWSHandler(dockerService *services.DockerService, monitorService *services.MonitoringService, allowedOrigins []string) *WSHandler {
	return &WSHandler{
		dockerService:  dockerService,
		monitorService: monitorService,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true // Allow non-browser clients
				}
				parsed, err := url.Parse(origin)
				if err != nil {
					return false
				}
				originHost := strings.ToLower(parsed.Scheme + "://" + parsed.Host)
				for _, allowed := range allowedOrigins {
					if strings.ToLower(strings.TrimSpace(allowed)) == originHost {
						return true
					}
				}
				return false
			},
		},
	}
}

// StreamEvents streams Docker container events over WebSocket
func (h *WSHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Close on client disconnect
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	events, errs := h.dockerService.GetEvents(ctx)

	for {
		select {
		case event := <-events:
			msg := map[string]interface{}{
				"type": "container",
				"data": map[string]interface{}{
					"action":       event.Action,
					"containerId":  event.Actor.ID,
					"containerName": event.Actor.Attributes["name"],
					"attributes":   event.Actor.Attributes,
					"time":         event.Time,
				},
			}
			data, _ := json.Marshal(msg)
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case err := <-errs:
			if err != nil {
				log.Printf("Docker events error: %v", err)
				errMsg, _ := json.Marshal(map[string]string{"type": "error", "message": err.Error()})
				conn.WriteMessage(websocket.TextMessage, errMsg)
			}
			return
		case <-ctx.Done():
			return
		}
	}
}

// StreamLogs streams container logs in real-time over WebSocket
func (h *WSHandler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Close on client disconnect
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	// Get streaming logs (follow=true)
	reader, err := h.dockerService.StreamContainerLogs(ctx, containerID)
	if err != nil {
		errMsg, _ := json.Marshal(map[string]string{"type": "error", "message": err.Error()})
		conn.WriteMessage(websocket.TextMessage, errMsg)
		return
	}
	defer reader.Close()

	containerInfo, err := h.dockerService.GetContainer(ctx, containerID)
	if err != nil {
		errMsg, _ := json.Marshal(map[string]string{"type": "error", "message": err.Error()})
		conn.WriteMessage(websocket.TextMessage, errMsg)
		return
	}

	go func() {
		<-ctx.Done()
		reader.Close()
	}()

	sendLogChunk := func(content []byte) error {
		if len(content) == 0 {
			return nil
		}
		msg := map[string]interface{}{
			"type": "log",
			"data": string(content),
		}
		data, _ := json.Marshal(msg)
		return conn.WriteMessage(websocket.TextMessage, data)
	}

	if containerInfo.Config != nil && containerInfo.Config.Tty {
		buf := make([]byte, 4096)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := reader.Read(buf)
				if n > 0 {
					if writeErr := sendLogChunk(buf[:n]); writeErr != nil {
						return
					}
				}
				if err != nil {
					return
				}
			}
		}
	}

	var mu sync.Mutex
	writer := &wsLogWriter{conn: conn, mu: &mu}
	_, _ = stdcopy.StdCopy(writer, writer, reader)
}

// StreamMetrics pushes system metrics over WebSocket every 3 seconds
func (h *WSHandler) StreamMetrics(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Close on client disconnect
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// Send initial metrics immediately
	metrics := h.monitorService.GetMetrics()
	msg := map[string]interface{}{
		"type":    "metrics",
		"data":    metrics,
	}
	data, _ := json.Marshal(msg)
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return
	}

	for {
		select {
		case <-ticker.C:
			metrics := h.monitorService.GetMetrics()
			msg := map[string]interface{}{
				"type":    "metrics",
				"data":    metrics,
			}
			data, _ := json.Marshal(msg)
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

type wsLogWriter struct {
	conn *websocket.Conn
	mu   *sync.Mutex
}

func (w *wsLogWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	msg := map[string]interface{}{
		"type": "log",
		"data": string(p),
	}
	data, _ := json.Marshal(msg)
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return 0, err
	}
	return len(p), nil
}
