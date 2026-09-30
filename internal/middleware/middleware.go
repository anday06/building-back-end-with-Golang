package middleware

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"log"
	"sync"
	"task-management-api/internal/config"
	"task-management-api/pkg/response"
	"task-management-api/pkg/token"
	"time"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("%s %s %d %s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start))
	}
}

func RateLimit() gin.HandlerFunc {
	var mu sync.Mutex
	clients := make(map[string]*rate.Limiter)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		limiter, ok := clients[ip]
		if !ok {
			limiter = rate.NewLimiter(rate.Limit(10), 20)
			clients[ip] = limiter
		}
		allowed := limiter.Allow()
		mu.Unlock()
		if !allowed {
			response.Error(c, 429, "rate limit exceeded")
			c.Abort()
			return
		}
		c.Next()
	}
}

func Auth(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		value := c.GetHeader("Authorization")
		if len(value) < 8 || value[:7] != "Bearer " {
			response.Error(c, 401, "missing or invalid authorization header")
			c.Abort()
			return
		}
		claims, err := token.Parse(value[7:], cfg.JWTSecret)
		if err != nil {
			response.Error(c, 401, "invalid or expired token")
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if role, _ := c.Get("role"); role != "admin" {
			response.Error(c, 403, "admin role required")
			c.Abort()
			return
		}
		c.Next()
	}
}
