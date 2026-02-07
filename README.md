# Sunspear

A CasaOS-styled Docker management dashboard with Halo Reach military HUD aesthetics. Self-host and manage Docker containers, browse an app marketplace, and monitor system metrics through an elegant interface.

![Sunspear](https://img.shields.io/badge/status-alpha-orange)
![Go](https://img.shields.io/badge/go-1.25-blue)
![Vue](https://img.shields.io/badge/vue-3.5-green)
![Vite](https://img.shields.io/badge/vite-7-blueviolet)

## Features

- **Docker Container Management** - Start, stop, restart, and remove containers
- **Image Management** - Pull images, view details, and manage your image library
- **App Marketplace** - One-click installation of popular self-hosted applications
- **System Monitoring** - Real-time CPU, RAM, disk, and network metrics
- **Halo Reach Theme** - Military HUD aesthetic with scanlines, corner brackets, and animations
- **Responsive Design** - Works on desktop and mobile devices

## Tech Stack

**Backend:**
- Go 1.25+
- Docker SDK
- Gorilla Mux (routing)
- SQLite (database)
- JWT authentication

**Frontend:**
- Vue 3 with Composition API
- Vite 7 (build tool)
- Pinia (state management)
- Vue Router
- Axios

## Prerequisites

- Docker and Docker Compose installed
- Fresh VPS or local machine with Docker access
- Public domain pointing to your server (for HTTPS)
- Ports 80 and 443 available

## Quick Start

### 1. Clone the repository

```bash
git clone <your-repo-url>
cd Sunspear
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` and set required values:

```env
PUBLIC_DOMAIN=your-domain.com
JWT_SECRET=your-very-secure-random-string-here
ADMIN_PASSWORD_HASH=your-bcrypt-hash
```

If you want to change the domain, update `PUBLIC_DOMAIN` and `Caddyfile`.
`ADMIN_PASSWORD_HASH` is optional if you plan to create the first user via `/api/auth/setup`.

### 3. Build and start

```bash
docker compose up -d --build
```

### 4. Access the dashboard

Open your browser and navigate to:

```
https://your-domain.com
```

On first run, you'll be prompted to create an admin account.

## Reverse Proxy (Caddy)

The default setup uses Caddy as a reverse proxy with automatic HTTPS:

- `/` -> frontend
- `/api/*` -> backend

If you change domains, update `PUBLIC_DOMAIN` in `.env` and `Caddyfile`.

## Development

### Backend Development

```bash
cd backend
go mod download
go run main.go
```

The backend API will be available at `http://localhost:8080`.

### Frontend Development

```bash
cd frontend
npm install
npm run dev
```

The frontend dev server will run at `http://localhost:5173` with hot module replacement.

## Project Structure

```
Sunspear/
├── backend/               # Go API server
│   ├── api/              # HTTP handlers and routing
│   ├── services/         # Business logic
│   ├── config/           # Configuration and database
│   ├── models/           # Data structures
│   └── data/             # App marketplace and database
│
├── frontend/             # Vue 3 SPA
│   ├── src/
│   │   ├── assets/       # Styles (Reach theme)
│   │   ├── components/   # Reusable components
│   │   ├── views/        # Page components
│   │   ├── stores/       # Pinia stores
│   │   ├── router/       # Vue Router
│   │   └── composables/  # Composition functions
│   └── public/           # Static assets
│
└── docker-compose.yml    # Container orchestration
```

## Available Commands

```bash
make build      # Build Docker containers
make up         # Start all services
make down       # Stop all services
make logs       # View logs
make restart    # Restart all services
make clean      # Stop and remove volumes
```

## API Endpoints

### Authentication
- `POST /api/auth/login` - Login
- `POST /api/auth/setup` - First-run setup
- `GET /api/auth/verify` - Verify token

### Containers
- `GET /api/containers` - List containers
- `GET /api/containers/:id` - Get container details
- `POST /api/containers/:id/start` - Start container
- `POST /api/containers/:id/stop` - Stop container
- `POST /api/containers/:id/restart` - Restart container
- `DELETE /api/containers/:id/remove` - Remove container
- `GET /api/containers/:id/logs` - Get container logs
- `POST /api/containers` - Create container

### Images
- `GET /api/images` - List images
- `POST /api/images/pull` - Pull image
- `DELETE /api/images/:id/remove` - Remove image
- `GET /api/images/search` - Search Docker Hub

### System
- `GET /api/system/metrics` - Current system metrics
- `GET /api/system/info` - Docker system info
- `GET /api/system/version` - Docker version

### Apps
- `GET /api/apps` - List marketplace apps
- `GET /api/apps/:id` - Get app details
- `POST /api/apps/:id/install` - Install app

## Design System

Sunspear uses the **Halo Reach military HUD aesthetic** adapted from the Infinity project:

- **Colors:** Reach Slate (#1a1f2e), Amber (#f6a623), Cyan (#22d3ee)
- **Typography:** Rajdhani (display), Inter (body), JetBrains Mono (technical)
- **Effects:** Scanline overlay, noise texture, corner brackets, pulsing status indicators
- **Components:** Glass-morphism nav, HUD-styled cards, monospace inputs

## Security

- JWT authentication with secure tokens
- bcrypt password hashing
- CORS protection
- Docker socket access limited to backend container
- No default credentials

**Important:** Set `JWT_SECRET` in production. `ADMIN_PASSWORD_HASH` is optional if you use the setup endpoint to create the first user.

## Roadmap

### Phase 1 - Foundation (complete)

- [x] Go backend with Gorilla Mux routing and SQLite
- [x] Docker SDK integration with API version negotiation
- [x] Vue 3 SPA with Pinia state management and Vue Router
- [x] Halo Reach HUD design system (scanlines, corner brackets, glass-morphism)
- [x] JWT authentication with bcrypt password hashing
- [x] First-run setup flow (create admin via `/api/auth/setup`)
- [x] System monitoring (CPU, RAM, disk via gopsutil with polling)
- [x] Docker Compose orchestration with Caddy reverse proxy and auto-HTTPS
- [x] Health check endpoints and scripts

### Phase 2 - Container Management (complete)

- [x] Container list with filtering (all / running) and search
- [x] Container detail view (general info, network, state, environment variables)
- [x] Container actions (start, stop, restart, remove with confirmation)
- [x] Container log viewer with tail and download
- [x] Container creation form (image, name, ports, volumes, env vars, restart policy)
- [x] Container resource stats display (CPU / memory usage per container)
- [x] Bulk container actions (stop all, restart all)
- [x] Container rename

### Phase 3 - Image Management (complete)

- [x] Image list view with details
- [x] Pull images from registry
- [x] Remove images
- [x] Docker Hub search from App Store page
- [x] Image tag management
- [x] Image build from Dockerfile (paste content)
- [x] Image layer inspection
- [x] Prune unused images (with confirmation)

### Phase 4 - Log Streaming and Real-Time Updates

- [ ] WebSocket server in Go backend
- [ ] Real-time container log streaming (replace polling with WebSocket)
- [ ] Live container status updates via Docker event stream (`GetEvents` already in `docker.go`)
- [ ] Dashboard auto-refresh without polling (WebSocket push for metrics)
- [ ] Toast notifications for container state changes (started, stopped, crashed)

### Phase 5 - App Marketplace

- [ ] App installation backend (read compose template, substitute variables, deploy)
- [ ] Installation wizard UI (configure env vars, ports, volumes before deploy)
- [ ] Compose template engine (parse and execute `data/apps/compose-templates/*.yml`)
- [ ] Installed apps tracking in SQLite
- [ ] App update and uninstall workflows
- [ ] Expand default app catalog (currently: Uptime Kuma, Jellyfin, Vaultwarden)
- [ ] Community app catalog support (import from URL)

### Phase 6 - Volume and Network Management

- [ ] Volume list / create / remove API endpoints
- [ ] Volume browser UI with size and mount info
- [ ] Network list / create / remove API endpoints
- [ ] Network inspector UI (connected containers, subnets, gateways)
- [ ] Attach / detach containers from networks

### Phase 7 - Multi-Container and Compose

- [ ] Docker Compose file upload and deploy
- [ ] Compose project view (group containers by project)
- [ ] Compose project lifecycle (up / down / restart as a unit)
- [ ] Compose file editor with syntax highlighting
- [ ] Stack templates (predefined multi-service stacks)

### Phase 8 - Polish and Hardening

- [ ] Global navbar refactor (extract shared layout component)
- [ ] Responsive mobile navigation (hamburger menu)
- [ ] Settings page (change password, theme toggle, API URL config)
- [ ] User management (multiple accounts, role-based access)
- [ ] Rate limiting and request validation on API endpoints
- [ ] Automated backup and restore for SQLite database
- [ ] Error boundary and global error handling in frontend
- [ ] Loading skeletons for all views

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Design inspired by Halo Reach military HUD
- Based on CasaOS architecture
- Built with Docker SDK and Vue 3

---

**Sunspear** - Manage your containers with style 🎯
