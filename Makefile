.PHONY: build up down logs restart clean dev-backend dev-frontend help

help:
	@echo "Sunspear - Docker Management Dashboard"
	@echo ""
	@echo "Available commands:"
	@echo "  make build          - Build all Docker containers"
	@echo "  make up             - Start all services"
	@echo "  make down           - Stop all services"
	@echo "  make logs           - View logs from all services"
	@echo "  make restart        - Restart all services"
	@echo "  make clean          - Stop and remove all containers and volumes"
	@echo "  make dev-backend    - Run backend in development mode"
	@echo "  make dev-frontend   - Run frontend in development mode"

build:
	docker-compose build

up:
	docker-compose up -d
	@echo "Sunspear is running!"
	@echo "Frontend: http://localhost:3000"
	@echo "Backend API: http://localhost:8080"

down:
	docker-compose down

logs:
	docker-compose logs -f

restart:
	docker-compose restart

clean:
	docker-compose down -v
	@echo "All containers and volumes removed"

dev-backend:
	@echo "Starting backend in development mode..."
	@echo "Make sure you have Go installed and run: cd backend && go run main.go"

dev-frontend:
	@echo "Starting frontend in development mode..."
	@echo "Make sure you have Node.js installed and run: cd frontend && npm run dev"
