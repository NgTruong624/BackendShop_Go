.PHONY: help build start stop restart logs clean status shell test dev-up dev-down prod-up prod-down

# Default help command
help:
	@echo "🐳 Docker Project Management"
	@echo "Available commands:"
	@echo "  make build     - Build Docker images"
	@echo "  make start     - Start development environment"
	@echo "  make stop      - Stop all services"
	@echo "  make restart   - Restart all services"
	@echo "  make logs      - Show logs from all services"
	@echo "  make clean     - Clean up containers, volumes, and images"
	@echo "  make status    - Show service status"
	@echo "  make shell     - Open shell in API container"
	@echo "  make test      - Run tests"
	@echo "  make dev-up    - Start with pgAdmin (development mode)"
	@echo "  make dev-down  - Stop development environment"
	@echo "  make prod-up   - Start production environment"
	@echo "  make prod-down - Stop production environment"

# Build Docker images
build:
	@echo "🔨 Building Docker images..."
	docker-compose build --no-cache

# Start development environment
start:
	@echo "🚀 Starting development environment..."
	docker-compose up -d
	@echo "✅ Services started!"
	@echo "🌐 API: http://localhost:8080"
	@echo "📊 Health check: http://localhost:8080/api/v1/status"

# Stop all services
stop:
	@echo "🛑 Stopping all services..."
	docker-compose down

# Restart all services
restart:
	@echo "🔄 Restarting services..."
	docker-compose restart

# Show logs
logs:
	docker-compose logs -f

# Clean up everything
clean:
	@echo "🧹 Cleaning up..."
	docker-compose down -v --remove-orphans
	docker system prune -af --volumes
	@echo "✅ Cleanup completed!"

# Show service status
status:
	@echo "📊 Service status:"
	docker-compose ps

# Open shell in API container
shell:
	docker-compose exec api sh

# Run tests
test:
	@echo "🧪 Running tests..."
	go test ./...

# Development mode with pgAdmin
dev-up:
	@echo "🚀 Starting development environment with pgAdmin..."
	docker-compose --profile dev up -d
	@echo "✅ Services started!"
	@echo "🌐 API: http://localhost:8080"
	@echo "🗄️  pgAdmin: http://localhost:5050 (admin@admin.com / admin)"

dev-down:
	docker-compose --profile dev down

# Production mode
prod-up:
	@echo "🚀 Starting production environment..."
	docker-compose -f docker-compose.prod.yml up -d

prod-down:
	docker-compose -f docker-compose.prod.yml down