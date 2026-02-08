package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"sunspear/services"

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

	// Parse request body matching frontend payload format
	var installReq struct {
		Name    string            `json:"name"`
		Env     []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"env"`
		Ports   map[string]string `json:"ports"`
		Volumes map[string]string `json:"volumes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&installReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required env vars
	envMap := make(map[string]string)
	for _, e := range installReq.Env {
		envMap[e.Name] = e.Value
	}
	for _, required := range app.EnvVars.Required {
		if val, ok := envMap[required.Name]; !ok || val == "" {
			http.Error(w, fmt.Sprintf("Missing required environment variable: %s", required.Name), http.StatusBadRequest)
			return
		}
	}

	// Use the app's Image field for the actual Docker Hub image
	imageName := app.Image

	// Pull the image
	pullReader, err := h.dockerService.PullImage(r.Context(), imageName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to pull image: %v", err), http.StatusInternalServerError)
		return
	}
	io.Copy(io.Discard, pullReader)
	pullReader.Close()

	// Prepare environment variables
	env := []string{}
	for _, e := range installReq.Env {
		if e.Value != "" {
			env = append(env, fmt.Sprintf("%s=%s", e.Name, e.Value))
		}
	}

	// Prepare port bindings from user-configured ports
	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}
	for label, hostPortStr := range installReq.Ports {
		// Get the container port from the app catalog using the label
		containerPort, ok := app.Ports[label]
		if !ok {
			continue
		}
		portStr := fmt.Sprintf("%d/tcp", containerPort)
		exposedPorts[nat.Port(portStr)] = struct{}{}
		portBindings[nat.Port(portStr)] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPortStr,
			},
		}
	}

	// Prepare volume binds from user-configured paths
	binds := []string{}
	for containerPath, hostPath := range installReq.Volumes {
		if hostPath != "" {
			binds = append(binds, fmt.Sprintf("%s:%s", hostPath, containerPath))
		}
	}

	// Use user-specified container name, fallback to app ID
	containerName := installReq.Name
	if containerName == "" {
		containerName = fmt.Sprintf("%s-app", appID)
	}

	// Create container
	containerConfig := &container.Config{
		Image:        imageName,
		Env:          env,
		ExposedPorts: exposedPorts,
	}

	restartPolicy := container.RestartPolicy{}
	setAppRestartPolicyName(&restartPolicy, "unless-stopped")
	hostConfig := &container.HostConfig{
		PortBindings:  portBindings,
		Binds:         binds,
		RestartPolicy: restartPolicy,
	}

	createResp, err := h.dockerService.CreateContainer(r.Context(), containerConfig, hostConfig, containerName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create container: %v", err), http.StatusInternalServerError)
		return
	}

	// Start container
	if err := h.dockerService.StartContainer(r.Context(), createResp.ID); err != nil {
		h.dockerService.RemoveContainer(r.Context(), createResp.ID, true)
		http.Error(w, fmt.Sprintf("Failed to start container: %v", err), http.StatusInternalServerError)
		return
	}

	// Record in database
	configMap := make(map[string]string)
	configMap["containerName"] = containerName
	installedApp, err := h.marketplaceService.InstallApp(appID, []string{createResp.ID}, configMap)
	if err != nil {
		http.Error(w, fmt.Sprintf("Container started but failed to track in database: %v", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id":          installedApp.ID,
		"appId":       installedApp.AppID,
		"appName":     installedApp.AppName,
		"containerId": createResp.ID,
		"status":      "running",
	})
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
		if err := h.dockerService.StopContainer(r.Context(), containerID, 10); err != nil {
			// Log error but continue with removal
		}

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

func setAppRestartPolicyName(policy *container.RestartPolicy, name string) {
	if policy == nil {
		return
	}
	policyValue := reflect.ValueOf(policy).Elem()
	nameField := policyValue.FieldByName("Name")
	if nameField.IsValid() && nameField.CanSet() && nameField.Kind() == reflect.String {
		nameField.SetString(name)
	}
}
