package services

import (
	"database/sql"
	"encoding/json"
	"os"
)

type App struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Icon        string            `json:"icon"`
	Category    string            `json:"category"`
	Version     string            `json:"version"`
	Ports       []int             `json:"ports"`
	Volumes     []string          `json:"volumes"`
	EnvVars     AppEnvVars        `json:"envVars"`
	ComposeFile string            `json:"composeFile"`
}

type AppEnvVars struct {
	Required []string `json:"required"`
	Optional []string `json:"optional"`
}

type AppCatalog struct {
	Apps []App `json:"apps"`
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

func (s *MarketplaceService) createDefaultCatalog() error {
	// Create default app catalog
	defaultCatalog := AppCatalog{
		Apps: []App{
			{
				ID:          "uptime-kuma",
				Name:        "Uptime Kuma",
				Description: "Self-hosted monitoring tool",
				Icon:        "📊",
				Category:    "monitoring",
				Version:     "latest",
				Ports:       []int{3001},
				Volumes:     []string{"data"},
				EnvVars: AppEnvVars{
					Required: []string{},
					Optional: []string{},
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
				Ports:       []int{8096},
				Volumes:     []string{"config", "media"},
				EnvVars: AppEnvVars{
					Required: []string{},
					Optional: []string{"TZ"},
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
				Ports:       []int{80},
				Volumes:     []string{"data"},
				EnvVars: AppEnvVars{
					Required: []string{},
					Optional: []string{"ADMIN_TOKEN"},
				},
				ComposeFile: "vaultwarden.yml",
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
