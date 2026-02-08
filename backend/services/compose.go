package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"gopkg.in/yaml.v3"
)

// ComposeSpec represents a docker-compose YAML file
type ComposeSpec struct {
	Version  string                     `yaml:"version"`
	Services map[string]ComposeServiceSpec `yaml:"services"`
	Networks map[string]interface{}     `yaml:"networks"`
	Volumes  map[string]interface{}     `yaml:"volumes"`
}

// ComposeServiceSpec represents a service definition in docker-compose
type ComposeServiceSpec struct {
	Image       string      `yaml:"image"`
	Ports       []string    `yaml:"ports"`
	Environment interface{} `yaml:"environment"`
	Volumes     []string    `yaml:"volumes"`
	Labels      interface{} `yaml:"labels"`
	Command     interface{} `yaml:"command"`
	Restart     string      `yaml:"restart"`
	DependsOn   interface{} `yaml:"depends_on"`
	Networks    []string    `yaml:"networks"`
}

// ComposeProject represents a deployed compose project
type ComposeProject struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	YAMLContent  string `json:"yamlContent"`
	Status       string `json:"status"`
	ContainerIDs string `json:"containerIds"`
	NetworkIDs   string `json:"networkIds"`
	VolumeNames  string `json:"volumeNames"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// StackTemplate represents a predefined compose template
type StackTemplate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	YAML        string `json:"yaml"`
}

// ComposeService manages docker-compose deployments
type ComposeService struct {
	db            *sql.DB
	dockerService *DockerService
}

func NewComposeService(db *sql.DB, dockerService *DockerService) *ComposeService {
	service := &ComposeService{
		db:            db,
		dockerService: dockerService,
	}
	service.createDefaultTemplates()
	return service
}

// ParseYAML parses and validates a docker-compose YAML string
func (s *ComposeService) ParseYAML(yamlContent string) (*ComposeSpec, error) {
	var spec ComposeSpec
	if err := yaml.Unmarshal([]byte(yamlContent), &spec); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if len(spec.Services) == 0 {
		return nil, fmt.Errorf("no services defined in compose file")
	}

	return &spec, nil
}

// Deploy creates and starts a new compose project
func (s *ComposeService) Deploy(ctx context.Context, name, description, yamlContent string) (*ComposeProject, error) {
	// Parse YAML
	spec, err := s.ParseYAML(yamlContent)
	if err != nil {
		return nil, err
	}

	// Create project network
	networkName := "sunspear-" + name
	networkResp, err := s.dockerService.CreateNetwork(ctx, networkName, "bridge", false)
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	var containerIDs []string
	var volumeNames []string
	networkIDs := []string{networkResp.ID}

	// Resolve service order
	serviceOrder, err := s.resolveServiceOrder(spec.Services)
	if err != nil {
		s.dockerService.RemoveNetwork(ctx, networkResp.ID)
		return nil, fmt.Errorf("failed to resolve service order: %w", err)
	}

	// Deploy services in order
	for _, serviceName := range serviceOrder {
		serviceSpec := spec.Services[serviceName]

		// Pull image
		pullReader, err := s.dockerService.PullImage(ctx, serviceSpec.Image)
		if err != nil {
			s.rollback(ctx, containerIDs, networkIDs)
			return nil, fmt.Errorf("failed to pull image %s: %w", serviceSpec.Image, err)
		}
		io.Copy(io.Discard, pullReader)
		pullReader.Close()

		// Parse configuration
		env := s.parseEnvironment(serviceSpec.Environment)
		exposedPorts, portBindings := s.parsePorts(serviceSpec.Ports)
		volumes := s.parseVolumes(serviceSpec.Volumes, name)
		labels := s.parseLabels(serviceSpec.Labels)
		command := s.parseCommand(serviceSpec.Command)

		// Add project labels
		if labels == nil {
			labels = make(map[string]string)
		}
		labels["com.sunspear.project"] = name
		labels["com.sunspear.service"] = serviceName

		// Create container config
		config := &container.Config{
			Image:        serviceSpec.Image,
			Env:          env,
			ExposedPorts: exposedPorts,
			Labels:       labels,
		}

		if len(command) > 0 {
			config.Cmd = command
		}

		// Create host config
		hostConfig := &container.HostConfig{
			PortBindings: portBindings,
			Binds:        volumes,
		}

		// Set restart policy
		if serviceSpec.Restart != "" {
			policy := container.RestartPolicy{}
			setRestartPolicyName(&policy, serviceSpec.Restart)
			hostConfig.RestartPolicy = policy
		}

		// Create container
		containerName := name + "-" + serviceName
		resp, err := s.dockerService.CreateContainer(ctx, config, hostConfig, containerName)
		if err != nil {
			s.rollback(ctx, containerIDs, networkIDs)
			return nil, fmt.Errorf("failed to create container %s: %w", serviceName, err)
		}

		containerIDs = append(containerIDs, resp.ID)

		// Connect to network
		if err := s.dockerService.ConnectNetwork(ctx, networkResp.ID, resp.ID); err != nil {
			s.rollback(ctx, containerIDs, networkIDs)
			return nil, fmt.Errorf("failed to connect %s to network: %w", serviceName, err)
		}

		// Start container
		if err := s.dockerService.StartContainer(ctx, resp.ID); err != nil {
			s.rollback(ctx, containerIDs, networkIDs)
			return nil, fmt.Errorf("failed to start container %s: %w", serviceName, err)
		}

		// Track volumes
		for _, vol := range volumes {
			if !strings.Contains(vol, ":") {
				volumeNames = append(volumeNames, vol)
			}
		}
	}

	// Save project to database
	containerIDsJSON, _ := json.Marshal(containerIDs)
	networkIDsJSON, _ := json.Marshal(networkIDs)
	volumeNamesJSON, _ := json.Marshal(volumeNames)

	result, err := s.db.Exec(`
		INSERT INTO compose_projects (name, description, yaml_content, status, container_ids, network_ids, volume_names)
		VALUES (?, ?, ?, 'running', ?, ?, ?)
	`, name, description, yamlContent, string(containerIDsJSON), string(networkIDsJSON), string(volumeNamesJSON))

	if err != nil {
		s.rollback(ctx, containerIDs, networkIDs)
		return nil, fmt.Errorf("failed to save project: %w", err)
	}

	projectID, _ := result.LastInsertId()

	return s.GetProject(int(projectID))
}

// StopProject stops all containers in a project
func (s *ComposeService) StopProject(ctx context.Context, id int) error {
	project, err := s.GetProject(id)
	if err != nil {
		return err
	}

	var containerIDs []string
	if err := json.Unmarshal([]byte(project.ContainerIDs), &containerIDs); err != nil {
		return fmt.Errorf("failed to parse container IDs: %w", err)
	}

	for _, containerID := range containerIDs {
		s.dockerService.StopContainer(ctx, containerID, 10)
	}

	_, err = s.db.Exec("UPDATE compose_projects SET status = 'stopped', updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}

// StartProject starts all containers in a project
func (s *ComposeService) StartProject(ctx context.Context, id int) error {
	project, err := s.GetProject(id)
	if err != nil {
		return err
	}

	var containerIDs []string
	if err := json.Unmarshal([]byte(project.ContainerIDs), &containerIDs); err != nil {
		return fmt.Errorf("failed to parse container IDs: %w", err)
	}

	for _, containerID := range containerIDs {
		if err := s.dockerService.StartContainer(ctx, containerID); err != nil {
			return err
		}
	}

	_, err = s.db.Exec("UPDATE compose_projects SET status = 'running', updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}

// RestartProject restarts all containers in a project
func (s *ComposeService) RestartProject(ctx context.Context, id int) error {
	if err := s.StopProject(ctx, id); err != nil {
		return err
	}
	time.Sleep(time.Second)
	return s.StartProject(ctx, id)
}

// DeleteProject tears down and removes a project
func (s *ComposeService) DeleteProject(ctx context.Context, id int) error {
	project, err := s.GetProject(id)
	if err != nil {
		return err
	}

	// Parse IDs
	var containerIDs []string
	var networkIDs []string
	json.Unmarshal([]byte(project.ContainerIDs), &containerIDs)
	json.Unmarshal([]byte(project.NetworkIDs), &networkIDs)

	// Stop and remove containers
	for _, containerID := range containerIDs {
		s.dockerService.StopContainer(ctx, containerID, 10)
		s.dockerService.RemoveContainer(ctx, containerID, true)
	}

	// Remove networks
	for _, networkID := range networkIDs {
		s.dockerService.RemoveNetwork(ctx, networkID)
	}

	// Remove from database
	_, err = s.db.Exec("DELETE FROM compose_projects WHERE id = ?", id)
	return err
}

// ListProjects returns all compose projects
func (s *ComposeService) ListProjects() ([]ComposeProject, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, yaml_content, status, container_ids, network_ids, volume_names, created_at, updated_at
		FROM compose_projects
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []ComposeProject{}
	for rows.Next() {
		var p ComposeProject
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.YAMLContent, &p.Status, &p.ContainerIDs, &p.NetworkIDs, &p.VolumeNames, &p.CreatedAt, &p.UpdatedAt); err != nil {
			continue
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// GetProject returns a single project by ID
func (s *ComposeService) GetProject(id int) (*ComposeProject, error) {
	var p ComposeProject
	err := s.db.QueryRow(`
		SELECT id, name, description, yaml_content, status, container_ids, network_ids, volume_names, created_at, updated_at
		FROM compose_projects
		WHERE id = ?
	`, id).Scan(&p.ID, &p.Name, &p.Description, &p.YAMLContent, &p.Status, &p.ContainerIDs, &p.NetworkIDs, &p.VolumeNames, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

// ListTemplates returns all available compose templates
func (s *ComposeService) ListTemplates() ([]StackTemplate, error) {
	templatesDir := "./data/apps/compose-templates"
	files, err := os.ReadDir(templatesDir)
	if err != nil {
		return []StackTemplate{}, nil
	}

	var templates []StackTemplate
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yml") {
			template, err := s.GetTemplate(strings.TrimSuffix(file.Name(), ".yml"))
			if err == nil {
				templates = append(templates, *template)
			}
		}
	}

	return templates, nil
}

// GetTemplate returns a specific template
func (s *ComposeService) GetTemplate(name string) (*StackTemplate, error) {
	// Sanitize name to prevent path traversal
	if strings.ContainsAny(name, "/\\..") || name != filepath.Base(name) {
		return nil, fmt.Errorf("invalid template name")
	}
	filePath := filepath.Join("./data/apps/compose-templates", name+".yml")
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid template path")
	}
	templatesDir, _ := filepath.Abs("./data/apps/compose-templates")
	if !strings.HasPrefix(absPath, templatesDir) {
		return nil, fmt.Errorf("invalid template name")
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Parse to get description from comments
	// Capitalize first letter for description
	displayName := name
	if len(displayName) > 0 {
		displayName = strings.ToUpper(displayName[:1]) + displayName[1:]
	}
	description := fmt.Sprintf("%s compose stack", displayName)

	return &StackTemplate{
		Name:        name,
		Description: description,
		YAML:        string(data),
	}, nil
}

// rollback removes created containers and networks on failure
func (s *ComposeService) rollback(ctx context.Context, containerIDs []string, networkIDs []string) {
	for _, containerID := range containerIDs {
		s.dockerService.RemoveContainer(ctx, containerID, true)
	}
	for _, networkID := range networkIDs {
		s.dockerService.RemoveNetwork(ctx, networkID)
	}
}

// parseEnvironment converts environment interface to string slice
func (s *ComposeService) parseEnvironment(env interface{}) []string {
	if env == nil {
		return nil
	}

	switch v := env.(type) {
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	case map[string]interface{}:
		result := make([]string, 0, len(v))
		for key, val := range v {
			result = append(result, fmt.Sprintf("%s=%v", key, val))
		}
		return result
	}

	return nil
}

// parsePorts converts port definitions to Docker format
func (s *ComposeService) parsePorts(ports []string) (nat.PortSet, nat.PortMap) {
	if len(ports) == 0 {
		return nil, nil
	}

	exposedPorts := make(nat.PortSet)
	portBindings := make(nat.PortMap)

	for _, portDef := range ports {
		parts := strings.Split(portDef, ":")
		var containerPort, hostPort string

		if len(parts) == 2 {
			hostPort = parts[0]
			containerPort = parts[1]
		} else {
			containerPort = parts[0]
			hostPort = parts[0]
		}

		// Add /tcp if not specified
		if !strings.Contains(containerPort, "/") {
			containerPort = containerPort + "/tcp"
		}

		port := nat.Port(containerPort)
		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{
			{HostPort: hostPort},
		}
	}

	return exposedPorts, portBindings
}

// parseVolumes converts volume definitions to Docker format
func (s *ComposeService) parseVolumes(volumes []string, projectName string) []string {
	if len(volumes) == 0 {
		return nil
	}

	result := make([]string, len(volumes))
	for i, vol := range volumes {
		// If it doesn't contain :, it's a named volume
		if !strings.Contains(vol, ":") {
			result[i] = projectName + "-" + vol
		} else {
			// Check if it's a named volume or bind mount
			parts := strings.Split(vol, ":")
			if !strings.HasPrefix(parts[0], "/") && !strings.HasPrefix(parts[0], ".") {
				// Named volume, prefix with project name
				parts[0] = projectName + "-" + parts[0]
				result[i] = strings.Join(parts, ":")
			} else {
				// Bind mount, keep as is
				result[i] = vol
			}
		}
	}

	return result
}

// parseLabels converts labels interface to map
func (s *ComposeService) parseLabels(labels interface{}) map[string]string {
	if labels == nil {
		return nil
	}

	switch v := labels.(type) {
	case []interface{}:
		result := make(map[string]string)
		for _, item := range v {
			if str, ok := item.(string); ok {
				parts := strings.SplitN(str, "=", 2)
				if len(parts) == 2 {
					result[parts[0]] = parts[1]
				}
			}
		}
		return result
	case map[string]interface{}:
		result := make(map[string]string)
		for key, val := range v {
			result[key] = fmt.Sprintf("%v", val)
		}
		return result
	}

	return nil
}

// parseCommand converts command interface to string slice
func (s *ComposeService) parseCommand(cmd interface{}) []string {
	if cmd == nil {
		return nil
	}

	switch v := cmd.(type) {
	case string:
		return strings.Fields(v)
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}

	return nil
}

// parseDependsOn converts depends_on interface to string slice
func (s *ComposeService) parseDependsOn(deps interface{}) []string {
	if deps == nil {
		return nil
	}

	switch v := deps.(type) {
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	case map[string]interface{}:
		result := make([]string, 0, len(v))
		for key := range v {
			result = append(result, key)
		}
		return result
	}

	return nil
}

// resolveServiceOrder performs topological sort based on depends_on
func (s *ComposeService) resolveServiceOrder(services map[string]ComposeServiceSpec) ([]string, error) {
	// Build dependency graph
	inDegree := make(map[string]int)
	dependents := make(map[string][]string)

	for name := range services {
		inDegree[name] = 0
	}

	for name, service := range services {
		deps := s.parseDependsOn(service.DependsOn)
		for _, dep := range deps {
			if _, exists := services[dep]; !exists {
				return nil, fmt.Errorf("service %s depends on undefined service %s", name, dep)
			}
			dependents[dep] = append(dependents[dep], name)
			inDegree[name]++
		}
	}

	// Topological sort using Kahn's algorithm
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	var result []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		for _, dependent := range dependents[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	if len(result) != len(services) {
		return nil, fmt.Errorf("circular dependency detected in service definitions")
	}

	return result, nil
}

// setRestartPolicyName uses reflection to set the restart policy name
// to handle different Docker SDK versions where Name can be string or typed
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

// createDefaultTemplates creates the template directory and default templates
func (s *ComposeService) createDefaultTemplates() {
	templatesDir := "./data/apps/compose-templates"
	os.MkdirAll(templatesDir, 0755)

	templates := map[string]string{
		"wordpress.yml": `version: "3.8"
services:
  wordpress:
    image: wordpress:latest
    ports:
      - "8080:80"
    environment:
      WORDPRESS_DB_HOST: db:3306
      WORDPRESS_DB_USER: wordpress
      WORDPRESS_DB_PASSWORD: wordpress
      WORDPRESS_DB_NAME: wordpress
    volumes:
      - wordpress_data:/var/www/html
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: mysql:5.7
    environment:
      MYSQL_DATABASE: wordpress
      MYSQL_USER: wordpress
      MYSQL_PASSWORD: wordpress
      MYSQL_ROOT_PASSWORD: rootpassword
    volumes:
      - db_data:/var/lib/mysql
    restart: unless-stopped

volumes:
  wordpress_data:
  db_data:
`,
		"gitea-postgres.yml": `version: "3.8"
services:
  gitea:
    image: gitea/gitea:latest
    ports:
      - "3000:3000"
      - "2222:22"
    environment:
      GITEA__database__DB_TYPE: postgres
      GITEA__database__HOST: db:5432
      GITEA__database__NAME: gitea
      GITEA__database__USER: gitea
      GITEA__database__PASSWD: gitea
    volumes:
      - gitea_data:/data
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: postgres:13
    environment:
      POSTGRES_USER: gitea
      POSTGRES_PASSWORD: gitea
      POSTGRES_DB: gitea
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  gitea_data:
  postgres_data:
`,
		"monitoring.yml": `version: "3.8"
services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - prometheus_data:/prometheus
    command:
      - "--config.file=/etc/prometheus/prometheus.yml"
      - "--storage.tsdb.path=/prometheus"
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
    volumes:
      - grafana_data:/var/lib/grafana
    depends_on:
      - prometheus
    restart: unless-stopped

volumes:
  prometheus_data:
  grafana_data:
`,
		"nextcloud-mariadb.yml": `version: "3.8"
services:
  nextcloud:
    image: nextcloud:latest
    ports:
      - "8081:80"
    environment:
      MYSQL_HOST: db
      MYSQL_DATABASE: nextcloud
      MYSQL_USER: nextcloud
      MYSQL_PASSWORD: nextcloud
      REDIS_HOST: redis
    volumes:
      - nextcloud_data:/var/www/html
    depends_on:
      - db
      - redis
    restart: unless-stopped

  db:
    image: mariadb:10.6
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: nextcloud
      MYSQL_USER: nextcloud
      MYSQL_PASSWORD: nextcloud
    volumes:
      - db_data:/var/lib/mysql
    restart: unless-stopped

  redis:
    image: redis:alpine
    restart: unless-stopped

volumes:
  nextcloud_data:
  db_data:
`,
		"planka.yml": `version: "3.8"
services:
  planka:
    image: ghcr.io/plankanban/planka:latest
    ports:
      - "3002:1337"
    environment:
      DATABASE_URL: postgresql://planka:planka@db:5432/planka
      SECRET_KEY: changeme-secret-key-minimum-32-chars
      BASE_URL: http://localhost:3002
    volumes:
      - planka_data:/app/public/user-avatars
      - planka_attachments:/app/public/project-background-images
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: postgres:14-alpine
    environment:
      POSTGRES_USER: planka
      POSTGRES_PASSWORD: planka
      POSTGRES_DB: planka
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  planka_data:
  planka_attachments:
  postgres_data:
`,
	}

	for filename, content := range templates {
		filePath := filepath.Join(templatesDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			os.WriteFile(filePath, []byte(content), 0644)
		}
	}
}
