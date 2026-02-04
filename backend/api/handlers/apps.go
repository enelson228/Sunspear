package handlers

import (
	"net/http"
	"sunspear/services"

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

	// TODO: Implement app installation logic
	// 1. Read compose template
	// 2. Substitute variables
	// 3. Deploy containers
	// 4. Track in database

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "App installation initiated",
		"appId":  appID,
	})
}

func (h *AppHandler) ListInstalledApps(w http.ResponseWriter, r *http.Request) {
	// TODO: Query database for installed apps
	respondJSON(w, http.StatusOK, []interface{}{})
}
