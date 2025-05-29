# Project Backend

A RESTful API backend service built with Go, Gin, and PostgreSQL.

## Features

- User authentication (Register/Login) with JWT
- Role-based access control (Admin/User)
- Product management (CRUD operations)
- Pagination and filtering for product listing
- Secure password hashing
- Environment variable configuration

## Prerequisites

- Go 1.23 or higher
- PostgreSQL
- Git

## Installation

1. Clone the repository:
```bash
git clone https://github.com/your-username/project_backend.git
cd project_backend
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file in the root directory with the following variables:
```env
DB_HOST=localhost
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name
DB_PORT=5432
JWT_SECRET=your_jwt_secret
PORT=8080
```

4. Run the application:
```bash
go run cmd/api/main.go
```

## API Endpoints

### Public Routes
- `POST /api/v1/register` - Register a new user
- `POST /api/v1/login` - Login user
- `GET /api/v1/products` - Get all products (with pagination and filtering)
- `GET /api/v1/products/:id` - Get product by ID

### Protected Routes (Requires JWT Token)
- `POST /api/v1/products` - Create new product (Admin only)
- `PUT /api/v1/products/:id` - Update product (Admin only)
- `DELETE /api/v1/products/:id` - Delete product (Admin only)

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── handlers/
│   │   ├── auth_handler.go
│   │   └── product_handler.go
│   ├── middleware/
│   │   └── auth.go
│   ├── models/
│   │   ├── product.go
│   │   └── user.go
│   └── utils/
│       └── response.go
├── .env
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## License

This project is licensed under the MIT License. 