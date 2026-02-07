package handlers

import (
	"encoding/json"
	"net/http"
	"sunspear/services"

	"github.com/gorilla/mux"
)

type VolumeHandler struct {
	dockerService *services.DockerService
}

func NewVolumeHandler(dockerService *services.DockerService) *VolumeHandler {
	return &VolumeHandler{dockerService: dockerService}
}

func (h *VolumeHandler) ListVolumes(w http.ResponseWriter, r *http.Request) {
	volumes, err := h.dockerService.ListVolumes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, volumes)
}

func (h *VolumeHandler) CreateVolume(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string            `json:"name"`
		Driver string            `json:"driver"`
		Labels map[string]string `json:"labels"`
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
		req.Driver = "local"
	}

	volume, err := h.dockerService.CreateVolume(r.Context(), req.Name, req.Driver, req.Labels)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, volume)
}

func (h *VolumeHandler) InspectVolume(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	volume, err := h.dockerService.InspectVolume(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, volume)
}

func (h *VolumeHandler) RemoveVolume(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	force := r.URL.Query().Get("force") == "true"

	if err := h.dockerService.RemoveVolume(r.Context(), name, force); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (h *VolumeHandler) PruneVolumes(w http.ResponseWriter, r *http.Request) {
	report, err := h.dockerService.PruneVolumes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"deleted":        report.VolumesDeleted,
		"spaceReclaimed": report.SpaceReclaimed,
	}

	respondJSON(w, http.StatusOK, response)
}
