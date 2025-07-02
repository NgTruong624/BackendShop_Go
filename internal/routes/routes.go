package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/NgTruong624/project_backend/internal/handlers"
	"github.com/NgTruong624/project_backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// adminMiddleware kiểm tra quyền admin
func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Permission denied",
				"message": "Admin access required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// SetupRouter configures all the routes for the application
func SetupRouter(
	authHandler *handlers.AuthHandler,
	productHandler *handlers.ProductHandler,
	adminHandler *handlers.AdminHandler,
	jwtMiddleware *middleware.JWTMiddleware,
) *gin.Engine {
	router := gin.Default()

	// Khởi tạo rate limiter
	middleware.InitGlobalRateLimiter()
	router.Use(middleware.RateLimitMiddleware())
	// Cấu hình static file serving
	router.Static("/uploads", "./static/uploads")

	// Middleware cho static files
	router.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/uploads/") {
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("X-Frame-Options", "DENY")
			c.Header("Content-Security-Policy", "default-src 'self'")
			c.Header("Cache-Control", "public, max-age=31536000")
			c.Header("Expires", time.Now().AddDate(1, 0, 0).Format(time.RFC1123))
		}
		c.Next()
	})

	// API v1 group
	api := router.Group("/api/v1")
	{
		// Status route
		api.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Rate limit stats route (admin only)
		api.GET("/rate-limit-stats", func(c *gin.Context) {
			stats := middleware.GetGlobalRateLimiter().GetStats()
			c.JSON(http.StatusOK, gin.H{
				"rate_limit_stats": stats,
			})
		})

		// Auth routes (Public)
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		// Protected routes
		authorized := api.Group("/")
		authorized.Use(jwtMiddleware.AuthMiddleware())
		{
			// User routes
			authorized.PUT("/users/change-password", authHandler.ChangePassword)

			// Product routes (Admin only)
			adminProducts := authorized.Group("/products")
			adminProducts.Use(adminMiddleware())
			{
				adminProducts.POST("", productHandler.CreateProduct)
				adminProducts.PUT("/:id", productHandler.UpdateProduct)
				adminProducts.DELETE("/:id", productHandler.DeleteProduct)

				// Upload routes (Admin only)
				uploadGroup := adminProducts.Group("/:id")
				uploadGroup.POST("/upload", productHandler.UploadProductImage)
			}

			// Admin routes
			admin := authorized.Group("/admin")
			admin.Use(adminMiddleware())
			{
				admin.GET("/users", adminHandler.GetUsersList)
			}
		}

		// Public product routes
		publicProductRoutes := api.Group("/products")
		{
			publicProductRoutes.GET("", productHandler.GetProducts)
			publicProductRoutes.GET("/:id", productHandler.GetProduct)
		}
	}

	return router
}
