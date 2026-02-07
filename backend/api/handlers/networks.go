package handlers

import (
	"encoding/json"
	"net/http"
	"sunspear/services"

	"github.com/gorilla/mux"
)

type NetworkHandler struct {
	dockerService *services.DockerService
}

func NewNetworkHandler(dockerService *services.DockerService) *NetworkHandler {
	return &NetworkHandler{dockerService: dockerService}
}

func (h *NetworkHandler) ListNetworks(w http.ResponseWriter, r *http.Request) {
	networks, err := h.dockerService.ListNetworks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, networks)
}

func (h *NetworkHandler) CreateNetwork(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Driver   string `json:"driver"`
		Internal bool   `json:"internal"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.Driver == "" {
		req.Driver = "bridge"
	}

	network, err := h.dockerService.CreateNetwork(r.Context(), req.Name, req.Driver, req.Internal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, network)
}

func (h *NetworkHandler) InspectNetwork(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	network, err := h.dockerService.InspectNetwork(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, network)
}

func (h *NetworkHandler) RemoveNetwork(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.dockerService.RemoveNetwork(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (h *NetworkHandler) ConnectContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	networkID := vars["id"]

	var req struct {
		ContainerID string `json:"containerId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.ContainerID == "" {
		http.Error(w, "containerId is required", http.StatusBadRequest)
		return
	}

	if err := h.dockerService.ConnectNetwork(r.Context(), networkID, req.ContainerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "connected"})
}

func (h *NetworkHandler) DisconnectContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	networkID := vars["id"]

	var req struct {
		ContainerID string `json:"containerId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.ContainerID == "" {
		http.Error(w, "containerId is required", http.StatusBadRequest)
		return
	}

	if err := h.dockerService.DisconnectNetwork(r.Context(), networkID, req.ContainerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "disconnected"})
}

func (h *NetworkHandler) PruneNetworks(w http.ResponseWriter, r *http.Request) {
	report, err := h.dockerService.PruneNetworks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"deleted": report.NetworksDeleted,
	}

	respondJSON(w, http.StatusOK, response)
}
