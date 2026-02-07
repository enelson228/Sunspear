package api

import (
	"database/sql"
	"net/http"
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
) http.Handler {
	r := mux.NewRouter()

	// Initialize handlers
	containerHandler := handlers.NewContainerHandler(dockerService)
	imageHandler := handlers.NewImageHandler(dockerService)
	systemHandler := handlers.NewSystemHandler(dockerService, monitorService)
	appHandler := handlers.NewAppHandler(marketplaceService, dockerService)
	authHandler := handlers.NewAuthHandler(cfg, db)
	wsHandler := handlers.NewWSHandler(dockerService, monitorService)

	// Public routes
	r.HandleFunc("/health", healthCheck).Methods("GET")
	r.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	r.HandleFunc("/api/auth/setup", authHandler.Setup).Methods("POST")
	r.HandleFunc("/api/auth/setup/status", authHandler.SetupStatus).Methods("GET")
	r.HandleFunc("/api/auth/verify", authHandler.Verify).Methods("GET")

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

	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
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
