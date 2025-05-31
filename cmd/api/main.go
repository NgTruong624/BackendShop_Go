package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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

	// Cấu hình static file serving
	// Cấu hình cho thư mục uploads với các tùy chọn bảo mật
	r.Static("/uploads", "./static/uploads")
	
	// Cấu hình cho các file tĩnh khác (nếu có)
	r.Static("/static", "./static")

	// Thêm middleware để kiểm soát truy cập file
	r.Use(func(c *gin.Context) {
		// Kiểm tra nếu request là truy cập file trong thư mục uploads
		if strings.HasPrefix(c.Request.URL.Path, "/uploads/") {
			// Thêm header bảo mật
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("X-Frame-Options", "DENY")
			c.Header("Content-Security-Policy", "default-src 'self'")
			
			// Cấu hình cache cho file
			c.Header("Cache-Control", "public, max-age=31536000") // Cache 1 năm
			c.Header("Expires", time.Now().AddDate(1, 0, 0).Format(time.RFC1123))
		}
		c.Next()
	})

	// Khởi tạo handlers và middleware
	jwtSecret := os.Getenv("JWT_SECRET")
	authHandler := handlers.NewAuthHandler(db, jwtSecret)
	productHandler := handlers.NewProductHandler(db)
	jwtMiddleware := middleware.NewJWTMiddleware(jwtSecret)

	// API Routes
	api := r.Group("/api/v1")
	{
		// Auth routes
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)

		// Product routes
		products := api.Group("/products")
		{
			// Public routes
			products.GET("", productHandler.GetProducts)    // Lấy danh sách sản phẩm
			products.GET("/:id", productHandler.GetProduct) // Xem chi tiết sản phẩm

			// Protected routes (yêu cầu JWT token và quyền admin)
			adminProducts := products.Group("")
			adminProducts.Use(jwtMiddleware.AuthMiddleware())
			{
				adminProducts.POST("", productHandler.CreateProduct)       // Tạo sản phẩm mới
				adminProducts.PUT("/:id", productHandler.UpdateProduct)    // Cập nhật sản phẩm
				adminProducts.DELETE("/:id", productHandler.DeleteProduct) // Xóa sản phẩm
				
				// Route upload ảnh với middleware bảo mật
				uploadGroup := adminProducts.Group("/:id")
				uploadGroup.Use(func(c *gin.Context) {
					// Kiểm tra quyền admin
					role := c.GetString("role")
					if role != "admin" {
						c.JSON(http.StatusForbidden, gin.H{
							"error": "Permission denied",
							"message": "Only admin can upload product images",
						})
						c.Abort()
						return
					}
					c.Next()
				})
				uploadGroup.POST("/upload", productHandler.UploadProductImage) // Upload ảnh sản phẩm
			}
		}
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
