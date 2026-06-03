package middleware

import (
	"database/sql"
	"net/http"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifies Firebase ID tokens and enriches the request context.
func AuthMiddleware(authClient *firebase.Client, dbConn *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString != "" {
			tokenString = strings.TrimSpace(strings.Replace(tokenString, "Bearer", "", 1))
		} else {
			// Fallback to cookie
			var err error
			tokenString, err = c.Cookie("session_token")
			if err != nil {
				// No token found
				if os.Getenv("DEV_AUTH_BYPASS") == "true" && os.Getenv("GO_ENV") != "production" {
					c.Set("user_id", "test-user-id")
					c.Set("user_email", "dev@example.com")
					c.Set("user_roles", "executive_admin,admin,vendor,customer")
					c.Set("user_dob", "2000-01-01")
				}
				c.Next()
				return
			}
		}

		if tokenString == "" {
			c.Next()
			return
		}

		// Strictly verify Firebase token
		fbToken, err := authClient.VerifyIDToken(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired Firebase token"})
			return
		}
		
		c.Set("user_id", fbToken.UID)

		// Fetch roles and DOB from DB for Firebase users
		if dbConn != nil {
			var roles sql.NullString
			var dob sql.NullTime
			err = dbConn.QueryRowContext(c.Request.Context(),
				"SELECT (SELECT string_agg(role::text, ',') FROM user_roles WHERE user_id = $1), date_of_birth FROM users WHERE id = $1",
				fbToken.UID).Scan(&roles, &dob)

			if err == nil {
				if roles.Valid && roles.String != "" {
					c.Set("user_roles", roles.String)
				} else {
					c.Set("user_roles", "user")
				}
				if dob.Valid {
					c.Set("user_dob", dob.Time.Format("2006-01-02"))
				}
			} else {
				c.Set("user_roles", "user")
			}
		}
		c.Next()
	}
}

// RequireAuth enforces that a user is authenticated (has user_id set in context)
func RequireAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        _, exists := c.Get("user_id")
        if !exists {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
            return
        }
        c.Next()
    }
}

// RequireRole ensures the current user has at least one of the provided roles.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        rolesVal, exists := c.Get("user_roles")
        if !exists {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied: missing roles"})
            return
        }
        rolesStr := rolesVal.(string)
        for _, r := range strings.Split(rolesStr, ",") {
            for _, want := range allowedRoles {
                if strings.TrimSpace(r) == want {
                    c.Next()
                    return
                }
            }
        }
        c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied"})
    }
}
