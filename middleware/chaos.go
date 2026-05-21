package middleware

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ChaosMiddleware injects random latency or HTTP 500 errors if CHAOS_MODE=true is set in the environment.
// It is intended to test client retry logic and resiliency.
func ChaosMiddleware() gin.HandlerFunc {
	chaosMode := os.Getenv("CHAOS_MODE") == "true" && os.Getenv("GO_ENV") != "production"

	return func(c *gin.Context) {
		if !chaosMode {
			c.Next()
			return
		}

		// 5% chance to inject an artificial 500 Internal Server Error
		if rand.Float32() < 0.05 {
			log.Printf("[CHAOS] Injecting HTTP 500 for request: %s %s", c.Request.Method, c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Chaos Monkey injected an internal error!",
			})
			return
		}

		// 30% chance to inject latency between 500ms and 3000ms
		if rand.Float32() < 0.30 {
			delay := time.Duration(rand.Intn(2500)+500) * time.Millisecond
			log.Printf("[CHAOS] Injecting %v latency for request: %s %s", delay, c.Request.Method, c.Request.URL.Path)
			time.Sleep(delay)
		}

		c.Next()
	}
}
