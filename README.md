# Project Backend

A RESTful API backend built with Go, Gin, and PostgreSQL, fully containerized with Docker for easy setup and deployment.

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

## Prerequisites

- **Docker & Docker Compose**: For running the application stack.
- **Git**: For cloning the repository.
- **Go & PostgreSQL** (Optional): For local development outside of Docker.

## 🚀 Quick Start with Docker

This is the recommended way to run the project for development and production.

1.  **Clone the repository:**
    ```sh
    git clone <repo-url>
    cd Project_backend_Go
    ```

2.  **Start the services:**
    Use the `Makefile` for convenience:
    ```sh
    # Start all services (API + PostgreSQL + pgAdmin) in the background
    make start
    ```
    Alternatively, use Docker Compose directly:
    ```sh
    # Start in detached mode
    docker-compose up -d
    ```

3.  **Access the services:**
    - **API Server**: `http://localhost:8080`
    - **API Status**: `http://localhost:8080/api/v1/status`
    - **pgAdmin (Database UI)**: `http://localhost:5050` (Login: `admin@admin.com` / `admin`)

4.  **Manage the application:**
    ```sh
    make help          # Show all available commands
    make status        # Check service status
    make logs          # View application logs
    make stop          # Stop all services
    make clean         # Stop and remove all containers, networks, and volumes
    ```

📖 **For more details on the Docker setup, see [DOCKER_GUIDE.md](DOCKER_GUIDE.md).**

---

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
- `POST /api/v1/products/:id/upload` – Upload product image (multipart/form-data, field: `image`)

### Admin Management
- `GET /api/v1/admin/users` – Get list of all users (admin only)

### Static Files & Security
- Uploaded images are served from `/uploads/<filename>`.
- The static file server includes security headers like `X-Content-Type-Options`, `X-Frame-Options`, and a strict `Content-Security-Policy`.

### API Status
- `GET /api/v1/status` – Check API health status.

---

## Project Details

### Authentication
The API uses JWT for authentication. After logging in, include the token in the `Authorization` header as a Bearer token.
```
Authorization: Bearer <your_jwt_token>
```
Access is role-based (Admin/User). Admins have extended privileges for managing products and users.

### Error Handling
The API returns detailed JSON error responses for validation, authentication, and business logic errors, including a `status`, `message`, and structured `error` field.

### Database Seeder
The database is automatically seeded with sample users and products when the application starts with `RUN_SEEDER=true` (the default in `docker-compose.yml`). You can also run the seeder manually.

### Project Structure
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
├── docker-compose.yml # Docker services definition
├── Dockerfile       # Docker build instructions for the Go app
├── .env.example     # Environment variable template
└── README.md
```

---

## 🔧 Development without Docker (Manual Setup)

If you prefer to run the Go application directly on your host machine:

1.  **Prerequisites**:
    - Go 1.23 or higher
    - PostgreSQL
    - Git

2.  **Setup**:
    ```sh
    # Clone the repo
    git clone <repo-url>
    cd Project_backend_Go

    # Create and configure your .env file
    cp .env.example .env
    # Edit .env with your local database credentials

    # Install Go dependencies
    go mod download
    ```

3.  **Run the application**:
    ```sh
    # Run database migrations and seeder
    go run cmd/seeder/main.go

    # Start the API server
    go run cmd/api/main.go
    ```

## License

This project is licensed under the MIT License.
 
