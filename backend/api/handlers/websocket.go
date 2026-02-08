package handlers

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"log"
	"net/http"
	"sunspear/services"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS handled by middleware
	},
}

type WSHandler struct {
	dockerService  *services.DockerService
	monitorService *services.MonitoringService
}

func NewWSHandler(dockerService *services.DockerService, monitorService *services.MonitoringService) *WSHandler {
	return &WSHandler{
		dockerService:  dockerService,
		monitorService: monitorService,
	}
}

// StreamEvents streams Docker container events over WebSocket
func (h *WSHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
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
				"type":   "container_event",
				"action": event.Action,
				"actor": map[string]interface{}{
					"id":         event.Actor.ID,
					"attributes": event.Actor.Attributes,
				},
				"time": event.Time,
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

	conn, err := upgrader.Upgrade(w, r, nil)
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

	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := reader.Read(buf)
			if n > 0 {
				// Docker log stream has 8-byte header per frame, strip it
				content := buf[:n]
				if n >= 8 {
					// Parse Docker multiplexing header: 1 byte stream type, 3 padding, 4 byte big-endian size
					frameSize := int(binary.BigEndian.Uint32(content[4:8]))
					if frameSize > 0 && 8+frameSize <= n {
						content = content[8 : 8+frameSize]
					} else if n > 8 {
						content = content[8:]
					}
				}
				msg := map[string]interface{}{
					"type": "log",
					"data": string(content),
				}
				data, _ := json.Marshal(msg)
				if writeErr := conn.WriteMessage(websocket.TextMessage, data); writeErr != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}
}

// StreamMetrics pushes system metrics over WebSocket every 3 seconds
func (h *WSHandler) StreamMetrics(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
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
		"metrics": metrics,
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
				"metrics": metrics,
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
