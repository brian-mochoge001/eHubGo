package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AgeGateMiddleware enforces that the user is at least 20 years old.
// Assumes DateOfBirth is available in user context or profile object.
func AgeGateMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mock: In production, fetch DOB from DB or User Profile in context
		dobStr := c.GetHeader("X-User-DOB") // Format: YYYY-MM-DD
		if dobStr == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Date of birth required for age verification"})
			return
		}

		dob, err := time.Parse("2006-01-02", dobStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid DOB format"})
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
