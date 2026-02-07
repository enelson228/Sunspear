package services

import (
	"database/sql"
	"encoding/json"
	"os"
)

type App struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	Category    string     `json:"category"`
	Version     string     `json:"version"`
	Image       string     `json:"image"`
	Ports       map[string]int `json:"ports"`
	Volumes     []AppVolume    `json:"volumes"`
	EnvVars     AppEnvVars     `json:"envVars"`
	ComposeFile string     `json:"composeFile"`
}

type AppVolume struct {
	Container string `json:"container"`
	Host      string `json:"host"`
}

type AppEnvVar struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Default     string `json:"default,omitempty"`
}

type AppEnvVars struct {
	Required []AppEnvVar `json:"required"`
	Optional []AppEnvVar `json:"optional"`
}

type AppCatalog struct {
	Apps []App `json:"apps"`
}

type InstalledApp struct {
	ID           int    `json:"id"`
	AppID        string `json:"appId"`
	AppName      string `json:"appName"`
	ContainerIDs string `json:"containerIds"`
	Config       string `json:"config"`
	InstalledAt  string `json:"installedAt"`
	Status       string `json:"status"`
}

type MarketplaceService struct {
	db      *sql.DB
	catalog AppCatalog
}

func NewMarketplaceService(db *sql.DB) *MarketplaceService {
	return &MarketplaceService{
		db: db,
	}
}

func (s *MarketplaceService) LoadApps() error {
	// Read apps.json
	data, err := os.ReadFile("./data/apps/apps.json")
	if err != nil {
		// Create default apps.json if it doesn't exist
		if os.IsNotExist(err) {
			return s.createDefaultCatalog()
		}
		return err
	}

	// Parse JSON
	if err := json.Unmarshal(data, &s.catalog); err != nil {
		return err
	}

	return nil
}

func (s *MarketplaceService) GetApps() []App {
	return s.catalog.Apps
}

func (s *MarketplaceService) GetApp(appID string) *App {
	for _, app := range s.catalog.Apps {
		if app.ID == appID {
			return &app
		}
	}
	return nil
}

func (s *MarketplaceService) InstallApp(appID string, containerIDs []string, config map[string]string) (*InstalledApp, error) {
	configJSON, _ := json.Marshal(config)
	idsJSON, _ := json.Marshal(containerIDs)

	app := s.GetApp(appID)
	appName := appID
	if app != nil {
		appName = app.Name
	}

	result, err := s.db.Exec(
		"INSERT INTO installed_apps (app_id, app_name, container_ids, config) VALUES (?, ?, ?, ?)",
		appID, appName, string(idsJSON), string(configJSON),
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &InstalledApp{
		ID:           int(id),
		AppID:        appID,
		AppName:      appName,
		ContainerIDs: string(idsJSON),
		Config:       string(configJSON),
		Status:       "running",
	}, nil
}

func (s *MarketplaceService) GetInstalledApps() ([]InstalledApp, error) {
	rows, err := s.db.Query("SELECT id, app_id, app_name, container_ids, config, installed_at FROM installed_apps")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []InstalledApp
	for rows.Next() {
		var app InstalledApp
		if err := rows.Scan(&app.ID, &app.AppID, &app.AppName, &app.ContainerIDs, &app.Config, &app.InstalledAt); err != nil {
			continue
		}
		apps = append(apps, app)
	}
	if apps == nil {
		apps = []InstalledApp{}
	}
	return apps, nil
}

func (s *MarketplaceService) GetInstalledApp(id int) (*InstalledApp, error) {
	var app InstalledApp
	err := s.db.QueryRow("SELECT id, app_id, app_name, container_ids, config, installed_at FROM installed_apps WHERE id = ?", id).
		Scan(&app.ID, &app.AppID, &app.AppName, &app.ContainerIDs, &app.Config, &app.InstalledAt)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (s *MarketplaceService) UninstallApp(id int) error {
	_, err := s.db.Exec("DELETE FROM installed_apps WHERE id = ?", id)
	return err
}

func (s *MarketplaceService) createDefaultCatalog() error {
	defaultCatalog := AppCatalog{
		Apps: []App{
			{
				ID:          "uptime-kuma",
				Name:        "Uptime Kuma",
				Description: "Self-hosted monitoring tool",
				Icon:        "📊",
				Category:    "monitoring",
				Version:     "1",
				Image:       "louislam/uptime-kuma:1",
				Ports:       map[string]int{"web": 3001},
				Volumes: []AppVolume{
					{Container: "/app/data", Host: "./uptime-kuma-data"},
				},
				EnvVars: AppEnvVars{
					Required: []AppEnvVar{},
					Optional: []AppEnvVar{},
				},
				ComposeFile: "uptime-kuma.yml",
			},
			{
				ID:          "jellyfin",
				Name:        "Jellyfin",
				Description: "Free media server",
				Icon:        "🎬",
				Category:    "media",
				Version:     "latest",
				Image:       "jellyfin/jellyfin:latest",
				Ports:       map[string]int{"web": 8096},
				Volumes: []AppVolume{
					{Container: "/config", Host: "./jellyfin-config"},
					{Container: "/media", Host: "./jellyfin-media"},
				},
				EnvVars: AppEnvVars{
					Required: []AppEnvVar{},
					Optional: []AppEnvVar{
						{Name: "TZ", Description: "Timezone", Default: "America/New_York"},
					},
				},
				ComposeFile: "jellyfin.yml",
			},
			{
				ID:          "vaultwarden",
				Name:        "Vaultwarden",
				Description: "Self-hosted password manager",
				Icon:        "🔐",
				Category:    "security",
				Version:     "latest",
				Image:       "vaultwarden/server:latest",
				Ports:       map[string]int{"web": 80},
				Volumes: []AppVolume{
					{Container: "/data", Host: "./vaultwarden-data"},
				},
				EnvVars: AppEnvVars{
					Required: []AppEnvVar{},
					Optional: []AppEnvVar{
						{Name: "ADMIN_TOKEN", Description: "Admin panel token"},
					},
				},
				ComposeFile: "vaultwarden.yml",
			},
			{
				ID:          "portainer",
				Name:        "Portainer",
				Description: "Container management platform",
				Icon:        "🐳",
				Category:    "tools",
				Version:     "latest",
				Image:       "portainer/portainer-ce:latest",
				Ports:       map[string]int{"web": 9000},
				Volumes: []AppVolume{
					{Container: "/data", Host: "portainer-data"},
					{Container: "/var/run/docker.sock", Host: "/var/run/docker.sock"},
				},
				EnvVars: AppEnvVars{
					Required: []AppEnvVar{},
					Optional: []AppEnvVar{},
				},
				ComposeFile: "portainer.yml",
			},
			{
				ID:          "nextcloud",
				Name:        "Nextcloud",
				Description: "Self-hosted cloud storage and collaboration platform",
				Icon:        "☁️",
				Category:    "productivity",
				Version:     "latest",
				Image:       "nextcloud:latest",
				Ports:       map[string]int{"web": 80},
				Volumes: []AppVolume{
					{Container: "/var/www/html", Host: "./nextcloud-data"},
				},
				EnvVars: AppEnvVars{
					Required: []AppEnvVar{
						{Name: "NEXTCLOUD_ADMIN_USER", Description: "Admin username"},
						{Name: "NEXTCLOUD_ADMIN_PASSWORD", Description: "Admin password"},
					},
					Optional: []AppEnvVar{},
				},
				ComposeFile: "nextcloud.yml",
			},
			{
				ID:          "gitea",
				Name:        "Gitea",
				Description: "Lightweight self-hosted Git service",
				Icon:        "🌿",
				Category:    "development",
				Version:     "latest",
				Image:       "gitea/gitea:latest",
				Ports:       map[string]int{"web": 3000, "ssh": 22},
				Volumes: []AppVolume{
					{Container: "/data", Host: "./gitea-data"},
				},
				EnvVars: AppEnvVars{
					Required: []AppEnvVar{},
					Optional: []AppEnvVar{
						{Name: "USER_UID", Description: "UID for Gitea user", Default: "1000"},
						{Name: "USER_GID", Description: "GID for Gitea user", Default: "1000"},
					},
				},
				ComposeFile: "gitea.yml",
			},
			{
				ID:          "pihole",
				Name:        "Pi-hole",
				Description: "Network-wide ad blocking DNS server",
				Icon:        "🛡️",
				Category:    "networking",
				Version:     "latest",
				Image:       "pihole/pihole:latest",
				Ports:       map[string]int{"web": 80, "dns": 53},
				Volumes: []AppVolume{
					{Container: "/etc/pihole", Host: "./pihole-config"},
					{Container: "/etc/dnsmasq.d", Host: "./pihole-dnsmasq"},
				},
				EnvVars: AppEnvVars{
					Required: []AppEnvVar{},
					Optional: []AppEnvVar{
						{Name: "WEBPASSWORD", Description: "Web interface password"},
					},
				},
				ComposeFile: "pihole.yml",
			},
			{
				ID:          "heimdall",
				Name:        "Heimdall",
				Description: "Application dashboard and launcher",
				Icon:        "🏠",
				Category:    "tools",
				Version:     "latest",
				Image:       "linuxserver/heimdall:latest",
				Ports:       map[string]int{"web": 80},
				Volumes: []AppVolume{
					{Container: "/config", Host: "./heimdall-config"},
				},
				EnvVars: AppEnvVars{
					Required: []AppEnvVar{},
					Optional: []AppEnvVar{
						{Name: "PUID", Description: "User ID", Default: "1000"},
						{Name: "PGID", Description: "Group ID", Default: "1000"},
						{Name: "TZ", Description: "Timezone", Default: "America/New_York"},
					},
				},
				ComposeFile: "heimdall.yml",
			},
			{
				ID:          "grafana",
				Name:        "Grafana",
				Description: "Analytics and monitoring dashboard",
				Icon:        "📈",
				Category:    "monitoring",
				Version:     "latest",
				Image:       "grafana/grafana:latest",
				Ports:       map[string]int{"web": 3000},
				Volumes: []AppVolume{
					{Container: "/var/lib/grafana", Host: "grafana-data"},
				},
				EnvVars: AppEnvVars{
					Required: []AppEnvVar{},
					Optional: []AppEnvVar{
						{Name: "GF_SECURITY_ADMIN_PASSWORD", Description: "Admin password", Default: "admin"},
					},
				},
				ComposeFile: "grafana.yml",
			},
		},
	}

	// Write to file
	data, err := json.MarshalIndent(defaultCatalog, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll("./data/apps", 0755); err != nil {
		return err
	}

	if err := os.WriteFile("./data/apps/apps.json", data, 0644); err != nil {
		return err
	}

	s.catalog = defaultCatalog
	return nil
}
