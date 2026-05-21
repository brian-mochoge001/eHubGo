package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AgeGateMiddleware enforces that the user is at least 20 years old.
// Now securely fetches DOB from authenticated user context.
func AgeGateMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		dobVal, exists := c.Get("user_dob")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Date of birth verification required. Please update your profile."})
			return
		}

		dobStr := dobVal.(string)
		dob, err := time.Parse("2006-01-02", dobStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to process age verification"})
			return
		}

		// Calculate age
		now := time.Now()
		years := now.Year() - dob.Year()
		if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
			years--
		}

		if years < 20 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You must be 20 years or older to access this service"})
			return
		}

		c.Next()
	}
}
