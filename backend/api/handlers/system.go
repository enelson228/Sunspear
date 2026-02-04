package handlers

import (
	"net/http"
	"sunspear/services"
)

type SystemHandler struct {
	dockerService  *services.DockerService
	monitorService *services.MonitoringService
}

func NewSystemHandler(dockerService *services.DockerService, monitorService *services.MonitoringService) *SystemHandler {
	return &SystemHandler{
		dockerService:  dockerService,
		monitorService: monitorService,
	}
}

func (h *SystemHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.monitorService.GetMetrics()
	respondJSON(w, http.StatusOK, metrics)
}

func (h *SystemHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	info, err := h.dockerService.GetSystemInfo(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, info)
}

func (h *SystemHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	version, err := h.dockerService.GetVersion(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, version)
}
