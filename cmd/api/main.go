package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/NgTruong624/project_backend/internal/handlers"
	"github.com/NgTruong624/project_backend/internal/middleware"
	"github.com/NgTruong624/project_backend/internal/models"
	"github.com/NgTruong624/project_backend/internal/routes"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file - không crash nếu không tìm thấy
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Could not load .env file, using environment variables")
	} else {
		log.Println("Successfully loaded .env file")
	}

	// Kết nối database
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=redis",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	fmt.Println("DSN String:", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate models
	if err := db.AutoMigrate(&models.User{}, &models.Product{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Seed data nếu được cấu hình
	if os.Getenv("RUN_SEEDER") == "true" {
		if err := seedData(db); err != nil {
			log.Printf("Warning: Failed to seed data: %v", err)
		} else {
			log.Println("Successfully seeded database")
		}
	}

	// Khởi tạo handlers và middleware
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	
	authHandler := handlers.NewAuthHandler(db, jwtSecret)
	productHandler := handlers.NewProductHandler(db)
	adminHandler := handlers.NewAdminHandler(db)
	jwtMiddleware := middleware.NewJWTMiddleware(jwtSecret)

	// Setup router với tất cả routes
	router := routes.SetupRouter(authHandler, productHandler, adminHandler, jwtMiddleware)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// seedData tạo dữ liệu mẫu cho database
func seedData(db *gorm.DB) error {
	// Tạo password hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Seed users
	users := []models.User{
		{
			Username:  "admin",
			Email:     "admin@example.com",
			Password:  string(hashedPassword),
			FullName:  "Admin User",
			Role:      "admin",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Username:  "user1",
			Email:     "user1@example.com",
			Password:  string(hashedPassword),
			FullName:  "Normal User",
			Role:      "user",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Insert users
	for _, user := range users {
		if err := db.FirstOrCreate(&user, models.User{Email: user.Email}).Error; err != nil {
			return err
		}
	}

	// Seed products
	products := []models.Product{
		{
			Name:        "Laptop Gaming",
			Description: "Laptop gaming cấu hình cao",
			Price:       25000000,
			Stock:       10,
			Category:    "Electronics",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "Smartphone",
			Description: "Điện thoại thông minh mới nhất",
			Price:       15000000,
			Stock:       20,
			Category:    "Electronics",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "Headphone",
			Description: "Tai nghe không dây",
			Price:       2000000,
			Stock:       50,
			Category:    "Accessories",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Insert products
	for _, product := range products {
		if err := db.FirstOrCreate(&product, models.Product{Name: product.Name}).Error; err != nil {
			return err
		}
	}

	return nil
}