package main

import (
	"fmt"
	"log"
	"os"

	"github.com/NgTruong624/project_backend/internal/handlers"
	"github.com/NgTruong624/project_backend/internal/middleware"
	"github.com/NgTruong624/project_backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Kết nối database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate models
	if err := db.AutoMigrate(&models.User{}, &models.Product{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Khởi tạo router
	r := gin.Default()

	// Khởi tạo handlers và middleware
	jwtSecret := os.Getenv("JWT_SECRET")
	authHandler := handlers.NewAuthHandler(db, jwtSecret)
	productHandler := handlers.NewProductHandler(db)
	jwtMiddleware := middleware.NewJWTMiddleware(jwtSecret)

	// Public routes
	public := r.Group("/api/v1")
	{
		// Auth routes
		public.POST("/register", authHandler.Register)
		public.POST("/login", authHandler.Login)

		// Product routes (public)
		public.GET("/products", productHandler.GetProducts)
		public.GET("/products/:id", productHandler.GetProduct)
	}

	// Protected routes
	protected := r.Group("/api/v1")
	protected.Use(jwtMiddleware.AuthMiddleware())
	{
		// Product routes (private - admin only)
		protected.POST("/products", productHandler.CreateProduct)
		protected.PUT("/products/:id", productHandler.UpdateProduct)
		protected.DELETE("/products/:id", productHandler.DeleteProduct)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
