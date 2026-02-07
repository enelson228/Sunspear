package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sunspear/services"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/gorilla/mux"
)

type AppHandler struct {
	marketplaceService *services.MarketplaceService
	dockerService      *services.DockerService
}

func NewAppHandler(marketplaceService *services.MarketplaceService, dockerService *services.DockerService) *AppHandler {
	return &AppHandler{
		marketplaceService: marketplaceService,
		dockerService:      dockerService,
	}
}

func (h *AppHandler) ListApps(w http.ResponseWriter, r *http.Request) {
	apps := h.marketplaceService.GetApps()
	respondJSON(w, http.StatusOK, apps)
}

func (h *AppHandler) GetApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appID := vars["id"]

	app := h.marketplaceService.GetApp(appID)
	if app == nil {
		http.Error(w, "App not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, app)
}

func (h *AppHandler) InstallApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appID := vars["id"]

	app := h.marketplaceService.GetApp(appID)
	if app == nil {
		http.Error(w, "App not found", http.StatusNotFound)
		return
	}

	// Parse request body
	var installReq struct {
		EnvVars map[string]string `json:"envVars"`
		Config  map[string]string `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&installReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required env vars
	for _, required := range app.EnvVars.Required {
		if _, ok := installReq.EnvVars[required]; !ok {
			http.Error(w, fmt.Sprintf("Missing required environment variable: %s", required), http.StatusBadRequest)
			return
		}
	}

	// Build image name
	imageName := fmt.Sprintf("%s:%s", appID, app.Version)

	// Pull the image
	pullReader, err := h.dockerService.PullImage(r.Context(), imageName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to pull image: %v", err), http.StatusInternalServerError)
		return
	}
	// Consume the pull output
	io.Copy(io.Discard, pullReader)
	pullReader.Close()

	// Prepare environment variables
	env := []string{}
	for k, v := range installReq.EnvVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Prepare port bindings
	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}
	for _, port := range app.Ports {
		portStr := fmt.Sprintf("%d/tcp", port)
		exposedPorts[nat.Port(portStr)] = struct{}{}
		portBindings[nat.Port(portStr)] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: strconv.Itoa(port),
			},
		}
	}

	// Prepare volumes
	binds := []string{}
	for _, vol := range app.Volumes {
		// Create named volumes or bind mounts
		binds = append(binds, fmt.Sprintf("%s-%s:%s", appID, vol, fmt.Sprintf("/data/%s", vol)))
	}

	// Container name
	containerName := fmt.Sprintf("%s-app", appID)

	// Create container
	containerConfig := &container.Config{
		Image:        imageName,
		Env:          env,
		ExposedPorts: exposedPorts,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Binds:        binds,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	createResp, err := h.dockerService.CreateContainer(r.Context(), containerConfig, hostConfig, containerName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create container: %v", err), http.StatusInternalServerError)
		return
	}

	// Start container
	if err := h.dockerService.StartContainer(r.Context(), createResp.ID); err != nil {
		// Cleanup: remove the created container
		h.dockerService.RemoveContainer(r.Context(), createResp.ID, true)
		http.Error(w, fmt.Sprintf("Failed to start container: %v", err), http.StatusInternalServerError)
		return
	}

	// Record in database
	installedApp, err := h.marketplaceService.InstallApp(appID, []string{createResp.ID}, installReq.Config)
	if err != nil {
		// Note: Container is already running, but DB tracking failed
		// In production, you might want to implement rollback here
		http.Error(w, fmt.Sprintf("Container started but failed to track in database: %v", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, installedApp)
}

func (h *AppHandler) ListInstalledApps(w http.ResponseWriter, r *http.Request) {
	apps, err := h.marketplaceService.GetInstalledApps()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get installed apps: %v", err), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, apps)
}

func (h *AppHandler) GetInstalledApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid app ID", http.StatusBadRequest)
		return
	}

	app, err := h.marketplaceService.GetInstalledApp(id)
	if err != nil {
		http.Error(w, "App not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, app)
}

func (h *AppHandler) UninstallApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid app ID", http.StatusBadRequest)
		return
	}

	// Get installed app info
	installedApp, err := h.marketplaceService.GetInstalledApp(id)
	if err != nil {
		http.Error(w, "App not found", http.StatusNotFound)
		return
	}

	// Parse container IDs
	var containerIDs []string
	if err := json.Unmarshal([]byte(installedApp.ContainerIDs), &containerIDs); err != nil {
		http.Error(w, "Failed to parse container IDs", http.StatusInternalServerError)
		return
	}

	// Stop and remove each container
	for _, containerID := range containerIDs {
		// Stop container with 10 second timeout
		if err := h.dockerService.StopContainer(r.Context(), containerID, 10); err != nil {
			// Log error but continue with removal
		}

		// Remove container
		if err := h.dockerService.RemoveContainer(r.Context(), containerID, true); err != nil {
			// Log error but continue
		}
	}

	// Delete from database
	if err := h.marketplaceService.UninstallApp(id); err != nil {
		http.Error(w, fmt.Sprintf("Failed to uninstall app: %v", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "App uninstalled successfully",
	})
}
