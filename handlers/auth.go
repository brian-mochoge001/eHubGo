package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"ehubgo/db"
	firebase "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Queries    *db.Queries
	DB         *sql.DB
	AuthClient *firebase.Client
}

func NewAuthHandler(queries *db.Queries, dbConn *sql.DB, authClient *firebase.Client) *AuthHandler {
	return &AuthHandler{
		Queries:    queries,
		DB:         dbConn,
		AuthClient: authClient,
	}
}

type SyncUserRequest struct {
	IDToken string `json:"id_token" binding:"required"`
	Role    string `json:"role"` // Optional role selection
}

// SyncUser verifies the Firebase ID token and synchronizes the user record in the local database.
func (h *AuthHandler) SyncUser(c *gin.Context) {
	var req SyncUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 1. Verify Firebase ID token
	fbToken, err := h.AuthClient.VerifyIDToken(c.Request.Context(), req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Firebase token"})
		return
	}

	// 2. Fetch or Create User in local database
	user, err := h.Queries.GetUserByEmail(c.Request.Context(), fbToken.Claims["email"].(string))
	if err != nil {
		if err == sql.ErrNoRows {
			// Create user
			user, err = h.Queries.CreateUser(c.Request.Context(), db.CreateUserParams{
				ID:        fbToken.UID,
				Email:     fbToken.Claims["email"].(string),
				FirstName: "User", // Can be updated later
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
	}

	// 3. Assign Role (if provided)
	if req.Role != "" {
		allowedRoles := []string{"user", "vendor", "driver", "host", "c2c_seller", "staff", "admin", "executive_admin"}
		isValid := false
		for _, r := range allowedRoles {
			if r == req.Role {
				isValid = true
				break
			}
		}
		if isValid {
			err = h.Queries.AssignRoleToUser(c.Request.Context(), db.AssignRoleToUserParams{
				UserID: user.ID,
				Role:   db.UserRoleType(req.Role),
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role"})
				return
			}
		}
	}

	// 4. Fetch user roles
	roles, err := h.fetchUserRoles(c.Request.Context(), user.ID)
	if err != nil {
		roles = []string{"user"}
	}

	// Set session cookie
	c.SetCookie("session_token", req.IDToken, 3600*24*7, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{
		"user_id": user.ID,
		"email":   user.Email,
		"roles":   roles,
	})
}

// Logout clears the session cookie.
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("session_token", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// fetchUserRoles retrieves all roles for a user.
func (h *AuthHandler) fetchUserRoles(ctx context.Context, userID string) ([]string, error) {
	var rolesStr sql.NullString
	err := h.DB.QueryRowContext(ctx,
		"SELECT string_agg(role::text, ',') FROM user_roles WHERE user_id = $1",
		userID).Scan(&rolesStr)

	if err != nil || !rolesStr.Valid {
		return []string{"user"}, nil
	}

	roles := strings.Split(rolesStr.String, ",")
	for i, r := range roles {
		roles[i] = strings.TrimSpace(r)
	}
	return roles, nil
}
