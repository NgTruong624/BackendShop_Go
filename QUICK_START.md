# 🚀 Quick Start Guide

## Kiểm tra nhanh Docker implementation

### 1. Cài đặt Docker (nếu chưa có)
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install docker.io docker-compose

# CentOS/RHEL
sudo yum install docker docker-compose

# macOS (với Homebrew)
brew install docker docker-compose
```

### 2. Test ngay lập tức
```bash
# Di chuyển vào thư mục project
cd Project_backend_Go

# Khởi động lần đầu
make start

# Hoặc
docker-compose up -d
```

### 3. Kiểm tra services
```bash
# Xem trạng thái
make status

# Kiểm tra API
curl http://localhost:8080/api/v1/status

# Test login với admin account
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}'
```

### 4. Truy cập services
- **API**: http://localhost:8080
- **pgAdmin**: http://localhost:5050 (admin@admin.com / admin)
- **Database**: localhost:5432 (project_user / project_password)

### 5. Commands cơ bản
```bash
make help          # Xem tất cả lệnh
make logs           # Xem logs
make stop           # Dừng services
make clean          # Dọn dẹp hoàn toàn
```

---

## ✅ Checklist triển khai thành công

- [ ] Docker services start successfully
- [ ] API responds on http://localhost:8080/api/v1/status
- [ ] Database migrations run automatically
- [ ] Sample data seeds correctly
- [ ] Login works with admin account
- [ ] File uploads work (test product image upload)
- [ ] pgAdmin can connect to database

## 🆘 Nếu có lỗi

1. **Xem logs**: `make logs`
2. **Restart**: `make restart`
3. **Clean rebuild**: `make clean && make start`
4. **Kiểm tra ports**: `lsof -i :8080`