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

type ClientInfo struct {
	Limiter      *rate.Limiter
	LastSeen     time.Time
	UserID       string
	RequestCount int64
}

type RateLimitConfig struct {
	Rate  rate.Limit
	Burst int
}

type RateLimiter struct {
	clients       map[string]*ClientInfo
	mu            sync.RWMutex
	configs       map[string]RateLimitConfig
	cleanupTicker *time.Ticker
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewRateLimiter() *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())

	rl := &RateLimiter{
		clients: make(map[string]*ClientInfo),
		configs: make(map[string]RateLimitConfig),
		ctx:     ctx,
		cancel:  cancel,
	}

	rl.configs["default"] = RateLimitConfig{Rate: 50, Burst: 100}

	rl.configs["auth"] = RateLimitConfig{Rate: rate.Every(20 * time.Second), Burst: 5}

	rl.configs["public"] = RateLimitConfig{Rate: 10, Burst: 100}
	rl.configs["admin"] = RateLimitConfig{Rate: 5, Burst: 50}
	rl.configs["product_read"] = RateLimitConfig{Rate: 20, Burst: 200}
	rl.configs["product_write"] = RateLimitConfig{Rate: 1, Burst: 20}

	rl.startCleanup()

	return rl
}

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

func (rl *RateLimiter) getConfigForEndpoint(path string) RateLimitConfig {
	cleanPath := strings.TrimSuffix(path, "/")

	if cleanPath == "/api/v1/auth/login" || cleanPath == "/api/v1/auth/register" {
		return rl.configs["auth"]
	}

	if cleanPath == "/api/v1/admin/users" {
		return rl.configs["admin"]
	}

	if cleanPath == "/api/v1/products" || cleanPath == "/api/v1/status" {
		return rl.configs["public"]
	}

	return rl.configs["default"]
}

func (rl *RateLimiter) getClientKey(c *gin.Context) string {
	ip := c.ClientIP()
	userID := c.GetString("user_id")

	if userID != "" {
		return fmt.Sprintf("%s:%s", ip, userID)
	}
	return ip
}

func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientKey := rl.getClientKey(c)
		config := rl.getConfigForEndpoint(c.Request.URL.Path)

		if c.Request.URL.Path == "/api/v1/products" {
			if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" {
				config = rl.configs["product_write"]
			} else {
				config = rl.configs["product_read"]
			}
		}

		rl.mu.Lock()

		client, exists := rl.clients[clientKey]
		if !exists {
			client = &ClientInfo{
				Limiter:  rate.NewLimiter(config.Rate, config.Burst),
				LastSeen: time.Now(),
				UserID:   c.GetString("user_id"),
			}
			rl.clients[clientKey] = client
		} else {
			currentRate := client.Limiter.Limit()
			currentBurst := client.Limiter.Burst()
			if currentRate != config.Rate || currentBurst != config.Burst {
				client.Limiter = rate.NewLimiter(config.Rate, config.Burst)
			}
		}

		client.LastSeen = time.Now()
		client.RequestCount++

		if !client.Limiter.Allow() {
			res := client.Limiter.Reserve()
			retry := res.Delay()
			res.CancelAt(time.Now())

			rl.mu.Unlock()

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%.4f", float64(config.Rate)))
			c.Header("X-RateLimit-Burst", fmt.Sprintf("%d", config.Burst))
			c.Header("X-RateLimit-Remaining", "0")
			if retry <= 0 {
				retry = time.Second
			}
			c.Header("Retry-After", fmt.Sprintf("%.0f", retry.Seconds()))

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too Many Requests",
				"message":     "Rate limit exceeded. Please try again later.",
				"retry_after": int(retry.Seconds()),
				"limit":       config.Rate,
				"burst":       config.Burst,
			})
			return
		}

		rl.mu.Unlock()

		res := client.Limiter.Reserve()
		delay := res.Delay()
		res.CancelAt(time.Now())

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%.4f", float64(config.Rate)))
		c.Header("X-RateLimit-Burst", fmt.Sprintf("%d", config.Burst))
		c.Header("X-RateLimit-Remaining", "-")
		if delay > 0 {
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%.0f", delay.Seconds()))
		} else {
			c.Header("X-RateLimit-Reset", "0")
		}

		c.Next()
	}
}

func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := map[string]interface{}{
		"total_clients": len(rl.clients),
		"clients":       make(map[string]interface{}),
	}

	for key, client := range rl.clients {
		stats["clients"].(map[string]interface{})[key] = map[string]interface{}{
			"last_seen":     client.LastSeen,
			"request_count": client.RequestCount,
			"user_id":       client.UserID,
			"note":          "tokens_remaining not available via public API",
		}
	}

	return stats
}

func (rl *RateLimiter) ClearAllClients() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.clients = make(map[string]*ClientInfo)
}

func (rl *RateLimiter) Close() {
	if rl.cleanupTicker != nil {
		rl.cleanupTicker.Stop()
	}
	rl.cancel()
}

var globalRateLimiter *RateLimiter

func InitGlobalRateLimiter() {
	globalRateLimiter = NewRateLimiter()
}

func ResetGlobalRateLimiter() {
	if globalRateLimiter != nil {
		globalRateLimiter.Close()
	}
	globalRateLimiter = NewRateLimiter()
}

func GetGlobalRateLimiter() *RateLimiter {
	if globalRateLimiter == nil {
		InitGlobalRateLimiter()
	}
	return globalRateLimiter
}

func RateLimitMiddleware() gin.HandlerFunc {
	return GetGlobalRateLimiter().RateLimitMiddleware()
}

func ClearAllClientsGlobal() {
	if globalRateLimiter != nil {
		globalRateLimiter.ClearAllClients()
	}
}

func ResetRateLimiterGlobal() {
	ResetGlobalRateLimiter()
}
