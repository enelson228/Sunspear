package api

import (
	"database/sql"
	"net/http"
	"strings"
	"sunspear/api/handlers"
	"sunspear/api/middleware"
	"sunspear/config"
	"sunspear/services"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func NewRouter(
	cfg *config.Config,
	db *sql.DB,
	dockerService *services.DockerService,
	monitorService *services.MonitoringService,
	marketplaceService *services.MarketplaceService,
	composeService *services.ComposeService,
) http.Handler {
	r := mux.NewRouter()

	// Initialize handlers
	containerHandler := handlers.NewContainerHandler(dockerService)
	imageHandler := handlers.NewImageHandler(dockerService)
	systemHandler := handlers.NewSystemHandler(dockerService, monitorService)
	appHandler := handlers.NewAppHandler(marketplaceService, dockerService)
	authHandler := handlers.NewAuthHandler(cfg, db)
	wsHandler := handlers.NewWSHandler(dockerService, monitorService)
	volumeHandler := handlers.NewVolumeHandler(dockerService)
	networkHandler := handlers.NewNetworkHandler(dockerService)
	composeHandler := handlers.NewComposeHandler(composeService)
	settingsHandler := handlers.NewSettingsHandler(cfg, db)

	// Public routes
	r.HandleFunc("/health", healthCheck).Methods("GET")
	r.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	r.HandleFunc("/api/auth/setup", authHandler.Setup).Methods("POST")
	r.HandleFunc("/api/auth/setup/status", authHandler.SetupStatus).Methods("GET")

	// Protected routes
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	// Container routes (bulk routes before {id} routes)
	api.HandleFunc("/containers", containerHandler.ListContainers).Methods("GET")
	api.HandleFunc("/containers", containerHandler.CreateContainer).Methods("POST")
	api.HandleFunc("/containers/bulk/stop", containerHandler.BulkStopContainers).Methods("POST")
	api.HandleFunc("/containers/bulk/restart", containerHandler.BulkRestartContainers).Methods("POST")
	api.HandleFunc("/containers/{id}", containerHandler.GetContainer).Methods("GET")
	api.HandleFunc("/containers/{id}/start", containerHandler.StartContainer).Methods("POST")
	api.HandleFunc("/containers/{id}/stop", containerHandler.StopContainer).Methods("POST")
	api.HandleFunc("/containers/{id}/restart", containerHandler.RestartContainer).Methods("POST")
	api.HandleFunc("/containers/{id}/rename", containerHandler.RenameContainer).Methods("POST")
	api.HandleFunc("/containers/{id}/remove", containerHandler.RemoveContainer).Methods("DELETE")
	api.HandleFunc("/containers/{id}/logs", containerHandler.GetLogs).Methods("GET")
	api.HandleFunc("/containers/{id}/stats", containerHandler.GetStats).Methods("GET")

	// Image routes (static routes before {id} routes)
	api.HandleFunc("/images", imageHandler.ListImages).Methods("GET")
	api.HandleFunc("/images/pull", imageHandler.PullImage).Methods("POST")
	api.HandleFunc("/images/build", imageHandler.BuildImage).Methods("POST")
	api.HandleFunc("/images/prune", imageHandler.PruneImages).Methods("POST")
	api.HandleFunc("/images/search", imageHandler.SearchImages).Methods("GET")
	api.HandleFunc("/images/{id}", imageHandler.InspectImage).Methods("GET")
	api.HandleFunc("/images/{id}/tag", imageHandler.TagImage).Methods("POST")
	api.HandleFunc("/images/{id}/history", imageHandler.GetImageHistory).Methods("GET")
	api.HandleFunc("/images/{id}/remove", imageHandler.RemoveImage).Methods("DELETE")

	// System routes
	api.HandleFunc("/system/metrics", systemHandler.GetMetrics).Methods("GET")
	api.HandleFunc("/system/info", systemHandler.GetInfo).Methods("GET")
	api.HandleFunc("/system/version", systemHandler.GetVersion).Methods("GET")

	// App marketplace routes
	api.HandleFunc("/apps", appHandler.ListApps).Methods("GET")
	api.HandleFunc("/apps/installed", appHandler.ListInstalledApps).Methods("GET")
	api.HandleFunc("/apps/installed/{id}", appHandler.GetInstalledApp).Methods("GET")
	api.HandleFunc("/apps/installed/{id}/uninstall", appHandler.UninstallApp).Methods("POST")
	api.HandleFunc("/apps/{id}", appHandler.GetApp).Methods("GET")
	api.HandleFunc("/apps/{id}/install", appHandler.InstallApp).Methods("POST")

	// WebSocket routes
	api.HandleFunc("/ws/events", wsHandler.StreamEvents).Methods("GET")
	api.HandleFunc("/ws/logs/{id}", wsHandler.StreamLogs).Methods("GET")
	api.HandleFunc("/ws/metrics", wsHandler.StreamMetrics).Methods("GET")

	// Volume routes (static before {name})
	api.HandleFunc("/volumes", volumeHandler.ListVolumes).Methods("GET")
	api.HandleFunc("/volumes", volumeHandler.CreateVolume).Methods("POST")
	api.HandleFunc("/volumes/prune", volumeHandler.PruneVolumes).Methods("POST")
	api.HandleFunc("/volumes/{name}", volumeHandler.InspectVolume).Methods("GET")
	api.HandleFunc("/volumes/{name}", volumeHandler.RemoveVolume).Methods("DELETE")

	// Network routes (static before {id})
	api.HandleFunc("/networks", networkHandler.ListNetworks).Methods("GET")
	api.HandleFunc("/networks", networkHandler.CreateNetwork).Methods("POST")
	api.HandleFunc("/networks/prune", networkHandler.PruneNetworks).Methods("POST")
	api.HandleFunc("/networks/{id}", networkHandler.InspectNetwork).Methods("GET")
	api.HandleFunc("/networks/{id}", networkHandler.RemoveNetwork).Methods("DELETE")
	api.HandleFunc("/networks/{id}/connect", networkHandler.ConnectContainer).Methods("POST")
	api.HandleFunc("/networks/{id}/disconnect", networkHandler.DisconnectContainer).Methods("POST")

	// Compose routes (static before {id})
	api.HandleFunc("/compose/projects", composeHandler.ListProjects).Methods("GET")
	api.HandleFunc("/compose/projects", composeHandler.DeployProject).Methods("POST")
	api.HandleFunc("/compose/validate", composeHandler.ValidateYAML).Methods("POST")
	api.HandleFunc("/compose/templates", composeHandler.ListTemplates).Methods("GET")
	api.HandleFunc("/compose/templates/{name}", composeHandler.GetTemplate).Methods("GET")
	api.HandleFunc("/compose/projects/{id}", composeHandler.GetProject).Methods("GET")
	api.HandleFunc("/compose/projects/{id}", composeHandler.DeleteProject).Methods("DELETE")
	api.HandleFunc("/compose/projects/{id}/start", composeHandler.StartProject).Methods("POST")
	api.HandleFunc("/compose/projects/{id}/stop", composeHandler.StopProject).Methods("POST")
	api.HandleFunc("/compose/projects/{id}/restart", composeHandler.RestartProject).Methods("POST")

	// Auth info routes
	api.HandleFunc("/auth/verify", authHandler.Verify).Methods("GET")
	api.HandleFunc("/auth/me", authHandler.Me).Methods("GET")

	// Settings routes
	api.HandleFunc("/settings", settingsHandler.GetSettings).Methods("GET")
	api.HandleFunc("/settings", settingsHandler.UpdateSettings).Methods("PUT")

	// User management routes
	api.HandleFunc("/users", settingsHandler.ListUsers).Methods("GET")
	api.HandleFunc("/users", settingsHandler.CreateUser).Methods("POST")
	api.HandleFunc("/users/{id}", settingsHandler.DeleteUser).Methods("DELETE")
	api.HandleFunc("/users/{id}/password", settingsHandler.ChangePassword).Methods("PUT")

	// CORS configuration — support comma-separated origins
	allowedOrigins := strings.Split(cfg.FrontendURL, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            false,
	})

	return c.Handler(r)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
