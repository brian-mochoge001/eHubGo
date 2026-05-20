package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type Client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]*Client)
)

func init() {
	go cleanupClients()
}

func cleanupClients() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, client := range clients {
			if time.Since(client.lastSeen) > 3*time.Minute {
				delete(clients, ip)
			}
		}
		mu.Unlock()
	}
}

// RateLimitMiddleware applies a rate limit per IP. Default config: 100 requests per minute with a burst of 200.
func RateLimitMiddleware() gin.HandlerFunc {
	// 100 requests per minute = ~1.67 requests per second. Burst = 200
	limit := rate.Every(time.Minute / 100)
	
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		mu.Lock()
		client, exists := clients[ip]
		if !exists {
			client = &Client{
				limiter: rate.NewLimiter(limit, 200),
			}
			clients[ip] = client
		}
		client.lastSeen = time.Now()
		limiter := client.limiter
		mu.Unlock()

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			return
		}

		c.Next()
	}
}
