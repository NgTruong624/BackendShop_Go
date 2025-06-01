# Project Backend

A RESTful API backend service built with Go, Gin, and PostgreSQL.

## Features

- User authentication (Register/Login) with JWT
- Role-based access control (Admin/User)
- Product management (CRUD operations)
- Product image upload (admin only)
- Pagination and filtering for product listing
- Secure password hashing
- Environment variable configuration
- Database seeder for sample data

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
│   └── utils/       # Utilities (response, etc.)
├── static/uploads/  # Uploaded product images
├── .env             # Environment variables
├── go.mod, go.sum   # Go modules
└── README.md
```

## Prerequisites

- Go 1.23 or higher
- PostgreSQL
- Git

## Setup & Run

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

### Auth
- `POST /api/v1/register` – Register new user
- `POST /api/v1/login` – Login and get JWT token

### Products
- `GET /api/v1/products` – List products (with pagination/filter)
- `GET /api/v1/products/:id` – Get product details
- `POST /api/v1/products` – Create product (admin only)
- `PUT /api/v1/products/:id` – Update product (admin only)
- `DELETE /api/v1/products/:id` – Delete product (admin only)
- `POST /api/v1/products/:id/upload` – Upload product image (admin only, multipart/form-data, field: image)

### Static Files
- Uploaded images are served at `/uploads/<filename>`

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
