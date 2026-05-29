package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// CustomClaims reflects the structure of the custom JWT issued by AuthHandler
type CustomClaims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// AuthMiddleware verifies either Firebase ID tokens or internal JWTs and enriches the request context.
func AuthMiddleware(authClient *firebase.Client, dbConn *sql.DB, jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			if os.Getenv("DEV_AUTH_BYPASS") == "true" && os.Getenv("GO_ENV") != "production" {
				c.Set("user_id", "test-user-id")
				c.Set("user_email", "dev@example.com")
				c.Set("user_roles", "executive_admin,admin,vendor,customer")
				c.Set("user_dob", "2000-01-01")
			}
			c.Next()
			return
		}

		tokenString := strings.TrimSpace(strings.Replace(authHeader, "Bearer", "", 1))
		if tokenString == "" {
			c.Next()
			return
		}

		// 1. Try Firebase validation first
		fbToken, err := authClient.VerifyIDToken(c.Request.Context(), tokenString)
		if err == nil {
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
			return
		}

		// 2. If Firebase fails, try custom JWT validation
		if len(jwtSecret) > 0 {
			claims := &CustomClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return jwtSecret, nil
			})

			if err == nil && token.Valid {
				c.Set("user_id", claims.UserID)
				c.Set("user_email", claims.Email)
				c.Set("user_roles", strings.Join(claims.Roles, ","))
				// Note: DOB is not in custom claims, but could be fetched from DB if needed
				c.Next()
				return
			}
		}

		// 3. Both failed
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
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
