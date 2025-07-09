# Project Backend

## Project Structure Guide

This project is a RESTful API backend built with Go, Gin, and PostgreSQL.

### Main Entry Points
- `cmd/api/main.go`: Main API server entry point. Loads configuration from `.env` and starts the HTTP server.
- `cmd/seeder/main.go`: Seeder entry point for populating the database with sample data.

### Configuration
- `.env`: Environment variables for database, JWT, and server configuration. See `.env.example` for template.

### Key Directories
- `internal/handlers/`: HTTP handlers for authentication, product, and admin endpoints.
- `internal/middleware/`: Middleware, including JWT authentication.
- `internal/models/`: Data models for users and products, plus request/response schemas.
- `internal/repository/`: Data access layer for users and products.
- `internal/utils/`: Utility functions (e.g., response formatting, error handling).
- `static/uploads/`: Uploaded product images, served at `/uploads/<filename>`.

### Development
- Run the API: `go run cmd/api/main.go`
- Seed the database: `go run cmd/seeder/main.go` or set `RUN_SEEDER=true` in `.env` and start the API.

---

## Features

- User authentication (Register/Login) with JWT
- Password management (change password with validation)
- Role-based access control (Admin/User)
- Product management (CRUD operations)
- Product image upload (admin only)
- Pagination and filtering for product listing
- Secure password hashing
- Environment variable configuration
- Database seeder for sample data
- Enhanced error handling with detailed validation messages

## Project Structure

```
Project_backend_Go/
├── cmd/
│   ├── api/         # Main API server
│   └── seeder/      # Seeder for sample data
├── internal/
│   ├── handlers/    # HTTP handlers
│   ├── middleware/  # Middleware (JWT, etc.)
│   ├── models/      # Data models
│   ├── repository/  # Data access layer
│   └── utils/       # Utilities (response, error handling)
├── static/uploads/  # Uploaded product images
├── .env             # Environment variables
├── go.mod, go.sum   # Go modules
└── README.md
```

## Prerequisites

### For Traditional Setup:
- Go 1.23 or higher
- PostgreSQL
- Git

### For Docker Setup (Recommended):
- Docker 20.10+
- Docker Compose 2.0+
- Git

## Setup & Run

### 🐳 Docker Setup (Recommended)

1. Clone the repository:
   ```sh
   git clone <repo-url>
   cd Project_backend_Go
   ```

2. Quick start with Docker:
   ```sh
   # Start all services (API + PostgreSQL + pgAdmin)
   make start
   
   # Or using docker-compose directly
   docker-compose up -d
   ```

3. Access services:
   - **API**: http://localhost:8080
   - **pgAdmin**: http://localhost:5050 (admin@admin.com / admin)
   - **API Status**: http://localhost:8080/api/v1/status

4. Management commands:
   ```sh
   make help          # Show all available commands
   make status        # Check service status
   make logs          # View logs
   make stop          # Stop services
   make clean         # Clean up everything
   ```

📖 **For detailed Docker documentation, see [DOCKER_GUIDE.md](DOCKER_GUIDE.md)**

### 🔧 Traditional Setup

1. Clone the repository:
   ```sh
   git clone <repo-url>
   cd Project_backend_Go
   ```
2. Create a `.env` file with your database and JWT settings:
   ```env
   DB_HOST=localhost
   DB_USER=youruser
   DB_PASSWORD=yourpassword
   DB_NAME=yourdb
   DB_PORT=5432
   JWT_SECRET=your_jwt_secret
   PORT=8080
   # Optional: RUN_SEEDER=true to seed data on API start
   ```
3. Install dependencies:
   ```sh
   go mod download
   ```
4. Run database migrations and seed sample data:
   ```sh
   go run cmd/seeder/main.go
   ```
5. Start the API server:
   ```sh
   go run cmd/api/main.go
   ```

## API Endpoints

### Authentication & User Management
- `POST /api/v1/auth/register` – Register new user
- `POST /api/v1/auth/login` – Login and get JWT token
- `PUT /api/v1/users/change-password` – Change user password (requires authentication)

### Products (Public)
- `GET /api/v1/products` – List all products
- `GET /api/v1/products/:id` – Get product details by ID

### Products (Admin Only)
- `POST /api/v1/products` – Create new product
- `PUT /api/v1/products/:id` – Update existing product
- `DELETE /api/v1/products/:id` – Delete product
- `POST /api/v1/products/:id/upload` – Upload product image (multipart/form-data, field: image)

### Admin Management
- `GET /api/v1/admin/users` – Get list of all users (admin only)

### Static Files
- Uploaded images are served at `/uploads/<filename>`
- Static files are served with security headers:
  - X-Content-Type-Options: nosniff
  - X-Frame-Options: DENY
  - Content-Security-Policy: default-src 'self'
  - Cache-Control: public, max-age=31536000

### API Status
- `GET /api/v1/status` – Check API health status

## Authentication

The API uses JWT (JSON Web Tokens) for authentication. To access protected endpoints:

1. Login using `/api/v1/auth/login` to get a JWT token
2. Include the token in subsequent requests in the Authorization header:
   ```
   Authorization: Bearer <your_jwt_token>
   ```

### Role-Based Access Control
- Regular users can only access public endpoints and their own user data
- Admin users have additional access to:
  - Product management (CRUD operations)
  - User management
  - Product image uploads

## Error Handling

The API provides detailed error responses for various scenarios:

1. Validation Errors:
   ```json
   {
     "status": 400,
     "message": "Validation failed",
     "error": {
       "current_password": "Current password is required.",
       "new_password": "New password must be at least 6 characters long.",
       "confirm_new_password": "Confirm password must match new password."
     }
   }
   ```

2. Authentication Errors:
   ```json
   {
     "status": 401,
     "message": "User not authenticated",
     "error": ""
   }
   ```

3. Business Logic Errors:
   ```json
   {
     "status": 400,
     "message": "Current password is incorrect",
     "error": ""
   }
   ```

## Seeder
- To seed sample users and products, run:
  ```sh
  go run cmd/seeder/main.go
  ```
- Or set `RUN_SEEDER=true` in `.env` to seed automatically when starting the API.

## Product Image Upload
- Only admin users can upload images for products.
- Use endpoint: `POST /api/v1/products/:id/upload` with form-data field `image` (accepts jpg, png, gif).
- Uploaded files are saved in `static/uploads/` and accessible via `/uploads/<filename>`.

## License

This project is licensed under the MIT License. 
