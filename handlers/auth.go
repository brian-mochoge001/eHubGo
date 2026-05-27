package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Queries *db.Queries
	DB      *sql.DB
	JWTKey  []byte
	Expiry  time.Duration
}

func NewAuthHandler(queries *db.Queries, dbConn *sql.DB, jwtKey []byte, expiryMinutes int) *AuthHandler {
	return &AuthHandler{
		Queries: queries,
		DB:      dbConn,
		JWTKey:  jwtKey,
		Expiry:  time.Duration(expiryMinutes) * time.Minute,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required,min=3"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type AuthResponse struct {
	Token  string   `json:"token"`
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
}

type Claims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// Login authenticates a user and returns a JWT token.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password format"})
		return
	}

	// Fetch user by email
	user, err := h.Queries.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Fetch user roles
	roles, err := h.fetchUserRoles(c.Request.Context(), user.ID)
	if err != nil {
		roles = []string{"user"}
	}

	// Generate JWT
	token, err := h.generateJWT(user.ID, user.Email, roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token:  token,
		UserID: user.ID,
		Email:  user.Email,
		Roles:  roles,
	})
}

// Register creates a new user account with hashed password.
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid registration data"})
		return
	}

	// Check if email already exists
	_, err := h.Queries.GetUserByEmail(c.Request.Context(), req.Email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}
	if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Create user in database
	user, err := h.Queries.CreateUser(c.Request.Context(), db.CreateUserParams{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     sql.NullString{String: req.LastName, Valid: req.LastName != ""},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Assign default "user" role
	_, err = h.Queries.AssignRoleToUser(c.Request.Context(), db.AssignRoleToUserParams{
		UserID: user.ID,
		Role:   "user",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role"})
		return
	}

	// Generate JWT
	roles := []string{"user"}
	token, err := h.generateJWT(user.ID, user.Email, roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token:  token,
		UserID: user.ID,
		Email:  user.Email,
		Roles:  roles,
	})
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

// generateJWT creates a signed JWT token.
func (h *AuthHandler) generateJWT(userID, email string, roles []string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.Expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.JWTKey)
}
