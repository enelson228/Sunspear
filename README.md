# Sunspear

A CasaOS-styled Docker management dashboard with Halo Reach military HUD aesthetics. Self-host and manage Docker containers, browse an app marketplace, and monitor system metrics through an elegant interface.

![Sunspear](https://img.shields.io/badge/status-alpha-orange)
![Go](https://img.shields.io/badge/go-1.21-blue)
![Vue](https://img.shields.io/badge/vue-3.4-green)

## Features

- **Docker Container Management** - Start, stop, restart, and remove containers
- **Image Management** - Pull images, view details, and manage your image library
- **App Marketplace** - One-click installation of popular self-hosted applications
- **System Monitoring** - Real-time CPU, RAM, disk, and network metrics
- **Halo Reach Theme** - Military HUD aesthetic with scanlines, corner brackets, and animations
- **Responsive Design** - Works on desktop and mobile devices

## Tech Stack

**Backend:**
- Go 1.21+
- Docker SDK
- Gorilla Mux (routing)
- SQLite (database)
- JWT authentication

**Frontend:**
- Vue 3 with Composition API
- Vite (build tool)
- Pinia (state management)
- Vue Router
- Axios

## Prerequisites

- Docker and Docker Compose installed
- Fresh VPS or local machine with Docker access
- Port 3000 and 8080 available

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

Edit `.env` and set a secure JWT secret:

```env
JWT_SECRET=your-very-secure-random-string-here
```

### 3. Build and start

```bash
make build
make up
```

Or using Docker Compose directly:

```bash
docker-compose build
docker-compose up -d
```

### 4. Access the dashboard

Open your browser and navigate to:

```
http://localhost:3000
```

On first run, you'll be prompted to create an admin account.

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

**Important:** Change the `JWT_SECRET` in production and use a strong admin password.

## Roadmap

- [x] Foundation and architecture
- [x] Backend API with Docker SDK
- [x] Frontend with Vue 3 and Reach theme
- [x] Authentication system
- [x] System monitoring
- [ ] Full container management UI
- [ ] Log viewer with streaming
- [ ] Container creation form
- [ ] App marketplace implementation
- [ ] Docker Compose template system
- [ ] WebSocket for real-time updates
- [ ] Volume and network management
- [ ] Multi-container app deployments

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
