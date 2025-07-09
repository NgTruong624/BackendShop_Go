package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ClientInfo lưu trữ thông tin chi tiết về client
type ClientInfo struct {
	Limiter      *rate.Limiter
	LastSeen     time.Time
	UserID       string // Cho authenticated requests
	RequestCount int64
}

// RateLimitConfig cấu hình rate limit cho từng loại endpoint
type RateLimitConfig struct {
	Rate  rate.Limit
	Burst int
}

// RateLimiter middleware với nhiều tính năng nâng cao
type RateLimiter struct {
	clients       map[string]*ClientInfo
	mu            sync.RWMutex
	configs       map[string]RateLimitConfig
	cleanupTicker *time.Ticker
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewRateLimiter tạo một rate limiter mới
func NewRateLimiter() *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())

	rl := &RateLimiter{
		clients: make(map[string]*ClientInfo),
		configs: make(map[string]RateLimitConfig),
		ctx:     ctx,
		cancel:  cancel,
	}

	// Cấu hình mặc định cho các loại endpoint
	rl.configs["default"] = RateLimitConfig{Rate: 10, Burst: 20}
	rl.configs["auth"] = RateLimitConfig{Rate: 0.1, Burst: 5}
	rl.configs["public"] = RateLimitConfig{Rate: 0.1, Burst: 100}
	rl.configs["admin"] = RateLimitConfig{Rate: 0.1, Burst: 50}

	// Bắt đầu cleanup routine
	rl.startCleanup()

	return rl
}

// startCleanup bắt đầu routine dọn dẹp clients không active
func (rl *RateLimiter) startCleanup() {
	rl.cleanupTicker = time.NewTicker(10 * time.Minute)
	go func() {
		for {
			select {
			case <-rl.cleanupTicker.C:
				rl.cleanupInactiveClients()
			case <-rl.ctx.Done():
				return
			}
		}
	}()
}

// cleanupInactiveClients xóa clients không active trong 1 giờ
func (rl *RateLimiter) cleanupInactiveClients() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour)
	for ip, client := range rl.clients {
		if client.LastSeen.Before(cutoff) {
			delete(rl.clients, ip)
		}
	}
}

// getConfigForEndpoint lấy cấu hình rate limit cho endpoint
func (rl *RateLimiter) getConfigForEndpoint(path string) RateLimitConfig {
	// Loại bỏ dấu '/' cuối nếu có
	cleanPath := strings.TrimSuffix(path, "/")

	// Auth endpoints
	if cleanPath == "/api/v1/auth/login" || cleanPath == "/api/v1/auth/register" {
		return rl.configs["auth"]
	}

	// Admin endpoints
	if cleanPath == "/api/v1/admin/users" {
		return rl.configs["admin"]
	}

	// Public endpoints
	if cleanPath == "/api/v1/products" || cleanPath == "/api/v1/status" {
		return rl.configs["public"]
	}

	return rl.configs["default"]
}

// getClientKey tạo key duy nhất cho client (IP + UserID nếu có)
func (rl *RateLimiter) getClientKey(c *gin.Context) string {
	ip := c.ClientIP()
	userID := c.GetString("user_id")

	if userID != "" {
		return fmt.Sprintf("%s:%s", ip, userID)
	}
	return ip
}

// RateLimitMiddleware tạo middleware với cấu hình linh hoạt
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientKey := rl.getClientKey(c)
		config := rl.getConfigForEndpoint(c.Request.URL.Path)

		// Điều chỉnh config dựa trên HTTP method
		if c.Request.URL.Path == "/api/v1/products" {
			if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" {
				config = rl.configs["admin"] // Admin endpoints có rate limit nghiêm ngặt hơn
			}
		}

		// Lock để truy cập map
		rl.mu.Lock()

		// Kiểm tra hoặc tạo client info
		client, exists := rl.clients[clientKey]
		if !exists {
			client = &ClientInfo{
				Limiter:  rate.NewLimiter(config.Rate, config.Burst),
				LastSeen: time.Now(),
				UserID:   c.GetString("user_id"),
			}
			rl.clients[clientKey] = client
		}

		// Cập nhật thời gian cuối
		client.LastSeen = time.Now()
		client.RequestCount++

		// Kiểm tra rate limit
		if !client.Limiter.Allow() {
			rl.mu.Unlock()

			// Trả về response với thông tin chi tiết
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%.2f", float64(config.Rate)))
			c.Header("X-RateLimit-Burst", fmt.Sprintf("%d", config.Burst))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "60")

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too Many Requests",
				"message":     "Rate limit exceeded. Please try again later.",
				"retry_after": 60,
				"limit":       config.Rate,
				"burst":       config.Burst,
			})
			return
		}

		rl.mu.Unlock()

		// Thêm headers cho response
		remaining := int(client.Limiter.TokensAt(time.Now()))
		reset := client.Limiter.Reserve().Delay() / time.Second
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.Burst))
		c.Header("X-RateLimit-Burst", fmt.Sprintf("%d", config.Burst))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", reset))

		c.Next()
	}
}

// GetStats trả về thống kê rate limiting
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := map[string]interface{}{
		"total_clients": len(rl.clients),
		"clients":       make(map[string]interface{}),
	}

	for key, client := range rl.clients {
		stats["clients"].(map[string]interface{})[key] = map[string]interface{}{
			"last_seen":        client.LastSeen,
			"request_count":    client.RequestCount,
			"user_id":          client.UserID,
			"tokens_remaining": client.Limiter.TokensAt(time.Now()),
		}
	}

	return stats
}

// Close dọn dẹp resources
func (rl *RateLimiter) Close() {
	if rl.cleanupTicker != nil {
		rl.cleanupTicker.Stop()
	}
	rl.cancel()
}

// Global instance để sử dụng trong toàn bộ ứng dụng
var globalRateLimiter *RateLimiter

// InitGlobalRateLimiter khởi tạo global rate limiter
func InitGlobalRateLimiter() {
	globalRateLimiter = NewRateLimiter()
}

// GetGlobalRateLimiter trả về global rate limiter
func GetGlobalRateLimiter() *RateLimiter {
	if globalRateLimiter == nil {
		InitGlobalRateLimiter()
	}
	return globalRateLimiter
}

// RateLimitMiddleware tạo middleware sử dụng global rate limiter
func RateLimitMiddleware() gin.HandlerFunc {
	return GetGlobalRateLimiter().RateLimitMiddleware()
}
