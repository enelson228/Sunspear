package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sunspear/services"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/gorilla/mux"
)

type ContainerHandler struct {
	dockerService *services.DockerService
}

func NewContainerHandler(dockerService *services.DockerService) *ContainerHandler {
	return &ContainerHandler{dockerService: dockerService}
}

func (h *ContainerHandler) ListContainers(w http.ResponseWriter, r *http.Request) {
	all := r.URL.Query().Get("all") == "true"

	containers, err := h.dockerService.ListContainers(r.Context(), all)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, containers)
}

func (h *ContainerHandler) GetContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]

	container, err := h.dockerService.GetContainer(r.Context(), containerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, container)
}

func (h *ContainerHandler) StartContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]

	if err := h.dockerService.StartContainer(r.Context(), containerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func (h *ContainerHandler) StopContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]

	timeout := 10
	if t := r.URL.Query().Get("timeout"); t != "" {
		if parsed, err := strconv.Atoi(t); err == nil {
			timeout = parsed
		}
	}

	if err := h.dockerService.StopContainer(r.Context(), containerID, timeout); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (h *ContainerHandler) RestartContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]

	timeout := 10
	if t := r.URL.Query().Get("timeout"); t != "" {
		if parsed, err := strconv.Atoi(t); err == nil {
			timeout = parsed
		}
	}

	if err := h.dockerService.RestartContainer(r.Context(), containerID, timeout); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "restarted"})
}

func (h *ContainerHandler) RemoveContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]

	force := r.URL.Query().Get("force") == "true"

	if err := h.dockerService.RemoveContainer(r.Context(), containerID, force); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (h *ContainerHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]

	tail := r.URL.Query().Get("tail")
	if tail == "" {
		tail = "100"
	}

	logs, err := h.dockerService.GetContainerLogs(r.Context(), containerID, tail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer logs.Close()

	w.Header().Set("Content-Type", "text/plain")
	io.Copy(w, logs)
}

func (h *ContainerHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]

	stats, err := h.dockerService.GetContainerStats(r.Context(), containerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stats.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, stats.Body)
}

func (h *ContainerHandler) CreateContainer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Image         string            `json:"image"`
		Name          string            `json:"name"`
		Ports         map[string]string `json:"ports"`
		Volumes       map[string]string `json:"volumes"`
		Env           []string          `json:"env"`
		RestartPolicy string            `json:"restartPolicy"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Prepare port bindings
	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}
	for containerPort, hostPort := range req.Ports {
		if hostPort == "" {
			continue
		}
		portStr := containerPort
		if !strings.Contains(portStr, "/") {
			portStr = portStr + "/tcp"
		}
		port := nat.Port(portStr)
		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: hostPort},
		}
	}

	// Prepare volume binds
	var binds []string
	for containerPath, hostPath := range req.Volumes {
		if hostPath != "" {
			binds = append(binds, fmt.Sprintf("%s:%s", hostPath, containerPath))
		}
	}

	config := &container.Config{
		Image:        req.Image,
		Env:          req.Env,
		ExposedPorts: exposedPorts,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Binds:        binds,
	}

	// Configure restart policy
	if req.RestartPolicy != "" {
		policy := container.RestartPolicy{}
		setRestartPolicyName(&policy, req.RestartPolicy)
		hostConfig.RestartPolicy = policy
	}

	response, err := h.dockerService.CreateContainer(r.Context(), config, hostConfig, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, response)
}

func (h *ContainerHandler) RenameContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if err := h.dockerService.RenameContainer(r.Context(), containerID, req.Name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "renamed", "name": req.Name})
}

func (h *ContainerHandler) BulkStopContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := h.dockerService.ListContainers(r.Context(), false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeout := 10
	stoppedCount := 0
	var errors []string

	for _, c := range containers {
		if c.State == "running" {
			if err := h.dockerService.StopContainer(r.Context(), c.ID, timeout); err != nil {
				errors = append(errors, c.ID[:12]+": "+err.Error())
			} else {
				stoppedCount++
			}
		}
	}

	response := map[string]interface{}{
		"stopped": stoppedCount,
		"total":   len(containers),
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *ContainerHandler) BulkRestartContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := h.dockerService.ListContainers(r.Context(), false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeout := 10
	restartedCount := 0
	var errors []string

	for _, c := range containers {
		if c.State == "running" {
			if err := h.dockerService.RestartContainer(r.Context(), c.ID, timeout); err != nil {
				errors = append(errors, c.ID[:12]+": "+err.Error())
			} else {
				restartedCount++
			}
		}
	}

	response := map[string]interface{}{
		"restarted": restartedCount,
		"total":     len(containers),
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	respondJSON(w, http.StatusOK, response)
}

func setRestartPolicyName(policy *container.RestartPolicy, name string) {
	if policy == nil {
		return
	}

	policyValue := reflect.ValueOf(policy).Elem()
	nameField := policyValue.FieldByName("Name")
	if nameField.IsValid() && nameField.CanSet() && nameField.Kind() == reflect.String {
		nameField.SetString(name)
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
