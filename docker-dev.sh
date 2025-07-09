#!/bin/bash

# Development Docker Management Script

case "$1" in
  "start")
    echo "🚀 Starting development environment..."
    docker-compose up -d
    echo "✅ Services started!"
    echo "🌐 API: http://localhost:8080"
    echo "🗄️  pgAdmin: http://localhost:5050 (admin@admin.com / admin)"
    ;;
  "stop")
    echo "🛑 Stopping services..."
    docker-compose down
    echo "✅ Services stopped!"
    ;;
  "restart")
    echo "🔄 Restarting services..."
    docker-compose restart
    ;;
  "logs")
    if [ -z "$2" ]; then
      docker-compose logs -f
    else
      docker-compose logs -f "$2"
    fi
    ;;
  "build")
    echo "🔨 Building application..."
    docker-compose build --no-cache
    ;;
  "clean")
    echo "🧹 Cleaning up..."
    docker-compose down -v
    docker system prune -f
    echo "✅ Cleanup completed!"
    ;;
  "status")
    echo "📊 Service status:"
    docker-compose ps
    ;;
  "shell")
    if [ "$2" = "api" ]; then
      docker-compose exec api sh
    elif [ "$2" = "db" ]; then
      docker-compose exec postgres psql -U project_user -d project_db
    else
      echo "Usage: $0 shell [api|db]"
    fi
    ;;
  *)
    echo "🐳 Docker Development Management"
    echo "Usage: $0 {start|stop|restart|logs|build|clean|status|shell}"
    echo ""
    echo "Commands:"
    echo "  start   - Start all services"
    echo "  stop    - Stop all services"
    echo "  restart - Restart all services"
    echo "  logs    - Show logs (add service name for specific service)"
    echo "  build   - Rebuild application image"
    echo "  clean   - Stop services and clean up volumes"
    echo "  status  - Show service status"
    echo "  shell   - Open shell (shell api|db)"
    ;;
esac