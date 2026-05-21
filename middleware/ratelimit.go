package middleware

import (
	"net/http"
	"sync"
	"time"

	"ehubgo/cache"
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

// RateLimitMiddleware applies a rate limit per IP. 
// Uses Redis for distributed rate limiting if available, otherwise falls back to in-memory.
func RateLimitMiddleware(redisStore cache.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if redisStore != nil && redisStore.IsAvailable() {
			// Distributed Rate Limiting via Redis
			key := "ratelimit:" + ip
			count, err := redisStore.Incr(c.Request.Context(), key)
			if err == nil {
				if count == 1 {
					// First request in the window, set expiration
					redisStore.Set(c.Request.Context(), key, count, time.Minute)
				}
				if count > 100 {
					c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
						"error": "Too many requests. Please try again later.",
					})
					return
				}
				c.Next()
				return
			}
			// If Redis fails, fall through to in-memory fallback
		}

		// In-Memory Fallback
		mu.Lock()
		client, exists := clients[ip]
		if !exists {
			client = &Client{
				limiter: rate.NewLimiter(rate.Every(time.Minute/100), 200),
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
