# 🐳 Docker Implementation Guide

## Tổng quan

Dự án này đã được container hóa sử dụng Docker và Docker Compose, cho phép:
- Môi trường phát triển nhất quán
- Triển khai dễ dàng
- Quản lý dependencies tự động
- Cách ly ứng dụng khỏi hệ thống host

## 📋 Yêu cầu hệ thống

- Docker 20.10+ 
- Docker Compose 2.0+
- Tối thiểu 4GB RAM
- 10GB dung lượng trống

## 🚀 Cài đặt và Khởi chạy

### Bước 1: Chuẩn bị môi trường
```bash
# Clone repository (nếu chưa có)
git clone <your-repo-url>
cd Project_backend_Go

# Tạo file .env từ template
cp .env.example .env

# Chỉnh sửa .env theo nhu cầu
nano .env
```

### Bước 2: Khởi chạy lần đầu
```bash
# Cách 1: Sử dụng Make (khuyến nghị)
make start

# Cách 2: Sử dụng Docker Compose trực tiếp
docker-compose up -d

# Cách 3: Sử dụng script
chmod +x docker-dev.sh
./docker-dev.sh start
```

### Bước 3: Kiểm tra trạng thái
```bash
# Kiểm tra services
make status

# Xem logs
make logs

# Kiểm tra health
curl http://localhost:8080/api/v1/status
```

## 🛠️ Các lệnh quản lý chính

### Make Commands (Khuyến nghị)
```bash
make help          # Hiển thị tất cả lệnh
make start          # Khởi động development
make stop           # Dừng services
make restart        # Khởi động lại
make logs           # Xem logs
make status         # Kiểm tra trạng thái
make shell          # Mở shell trong container API
make clean          # Dọn dẹp hoàn toàn
make dev-up         # Khởi động với pgAdmin
make build          # Build lại images
```

### Docker Compose Commands
```bash
docker-compose up -d               # Khởi động background
docker-compose down                # Dừng và xóa containers
docker-compose logs -f            # Theo dõi logs real-time
docker-compose ps                 # Liệt kê containers
docker-compose exec api sh        # Mở shell trong API container
docker-compose exec postgres psql -U project_user -d project_db
```

## 🌍 Môi trường Development

### Services được khởi động:
- **API Server**: http://localhost:8080
- **PostgreSQL**: localhost:5432
- **pgAdmin** (optional): http://localhost:5050

### Credentials mặc định:
- **Database**: project_user / project_password
- **pgAdmin**: admin@admin.com / admin

### Kiểm tra API:
```bash
# Health check
curl http://localhost:8080/api/v1/status

# Test API với seeded data
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}'
```

## 🏭 Môi trường Production

### Chuẩn bị:
```bash
# Tạo file .env.prod với thông tin production
cp .env.example .env.prod

# Chỉnh sửa với thông tin thật
nano .env.prod
```

### Khởi chạy production:
```bash
# Sử dụng file production
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d

# Hoặc với Make
make prod-up
```

### Tính năng Production:
- Resource limits
- Health checks
- No development tools
- Optimized logging
- Nginx reverse proxy (optional)

## 📁 Cấu trúc Files

```
Project_backend_Go/
├── Dockerfile                 # Multi-stage build cho Go app
├── docker-compose.yml         # Development environment
├── docker-compose.prod.yml    # Production environment
├── .dockerignore             # Exclude files from build
├── .env.example              # Template cho environment variables
├── Makefile                  # Management commands
├── docker-dev.sh            # Development script
└── DOCKER_GUIDE.md          # Tài liệu này
```

## 🔧 Customization

### Database Configuration
Chỉnh sửa trong `docker-compose.yml`:
```yaml
postgres:
  environment:
    POSTGRES_USER: your_user
    POSTGRES_PASSWORD: your_password
    POSTGRES_DB: your_db
```

### Application Settings
Chỉnh sửa environment variables trong service `api`:
```yaml
api:
  environment:
    JWT_SECRET: your_secret_key
    RUN_SEEDER: "true"  # Auto seed on startup
```

### Ports
Thay đổi port mapping:
```yaml
api:
  ports:
    - "3000:8080"  # Host:Container
```

## 🚨 Troubleshooting

### Lỗi thường gặp:

1. **Port đã được sử dụng**
   ```bash
   # Kiểm tra port đang sử dụng
   lsof -i :8080
   # Thay đổi port trong docker-compose.yml
   ```

2. **Database connection failed**
   ```bash
   # Kiểm tra PostgreSQL health
   docker-compose exec postgres pg_isready -U project_user
   # Xem logs database
   docker-compose logs postgres
   ```

3. **Build failed**
   ```bash
   # Clean và rebuild
   make clean
   make build
   ```

4. **Permissions issues**
   ```bash
   # Fix uploads directory
   sudo chown -R $(id -u):$(id -g) static/uploads
   chmod 755 static/uploads
   ```

### Debug Commands:
```bash
# Vào container để debug
docker-compose exec api sh

# Kiểm tra logs chi tiết
docker-compose logs --tail=100 api

# Kiểm tra network
docker network ls
docker network inspect project_backend_go_project_network
```

## 📊 Monitoring

### Health Checks:
```bash
# API health
curl http://localhost:8080/api/v1/status

# Database health
docker-compose exec postgres pg_isready -U project_user -d project_db
```

### Performance Monitoring:
```bash
# Resource usage
docker stats

# Container logs
docker-compose logs -f --tail=50 api
```

## 🔄 Updates & Maintenance

### Cập nhật code:
```bash
# Pull latest code
git pull origin main

# Rebuild và restart
make restart
```

### Backup Database:
```bash
# Create backup
docker-compose exec postgres pg_dump -U project_user project_db > backup.sql

# Restore backup
docker-compose exec -T postgres psql -U project_user -d project_db < backup.sql
```

### Clean up:
```bash
# Xóa tất cả containers và volumes
make clean

# Xóa unused images
docker image prune -a
```

## 🛡️ Security Notes

1. **Thay đổi default passwords** trong production
2. **Sử dụng HTTPS** cho production deployment
3. **Không expose database port** ra ngoài trong production
4. **Sử dụng secrets management** cho sensitive data
5. **Regular security updates** cho base images

## 📝 Logs

### Xem logs:
```bash
# Tất cả services
make logs

# Specific service
docker-compose logs -f api
docker-compose logs -f postgres
```

### Log rotation (Production):
- Cấu hình log rotation cho Docker
- Sử dụng external logging solutions như ELK stack

## 🤝 Contribute

Khi thêm tính năng mới:
1. Update Dockerfile nếu cần
2. Update docker-compose.yml
3. Update tài liệu này
4. Test trong môi trường Docker

---

## 📞 Support

Nếu gặp vấn đề:
1. Kiểm tra logs: `make logs`
2. Kiểm tra status: `make status`
3. Restart services: `make restart`
4. Clean và rebuild: `make clean && make build && make start`