package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewUserHandler(queries *db.Queries, dbConn *sql.DB) *UserHandler {
	return &UserHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *UserHandler) GetWalletBalance(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		wallet, err := qtx.GetWalletBalance(c.Request.Context(), userID)
		if err != nil {
			if err == sql.ErrNoRows {
				// Initialize wallet if not exists
				wallet, err = qtx.UpdateWalletBalance(c.Request.Context(), db.UpdateWalletBalanceParams{
					UserID:  userID,
					Balance: "0.00",
				})
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		c.JSON(http.StatusOK, wallet)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *UserHandler) ListMessages(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		messages, err := qtx.ListUserMessages(c.Request.Context(), userID)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, messages)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *UserHandler) ListNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		notifications, err := qtx.ListUserNotifications(c.Request.Context(), userID)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, notifications)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *UserHandler) ListBills(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		bills, err := qtx.ListUserBills(c.Request.Context(), userID)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, bills)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	uid := userID.(string)

	var user struct {
		ID         string `json:"id"`
		Email      string `json:"email"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		ProfileUrl string `json:"profile_picture_url"`
	}

	err := h.DB.QueryRowContext(c.Request.Context(),
		"SELECT id, email, first_name, COALESCE(last_name, ''), COALESCE(profile_picture_url, '') FROM users WHERE id = $1",
		uid).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.ProfileUrl)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data"})
		return
	}

	rolesVal, exists := c.Get("user_roles")
	var roles []string
	if exists {
		rolesStr := rolesVal.(string)
		if rolesStr != "" {
			roles = strings.Split(rolesStr, ",")
		}
	}
	if len(roles) == 0 {
		roles = []string{"user"}
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"roles": roles,
	})
}

// ListUsers returns all users in the system (Admin only)
func (h *UserHandler) ListUsers(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		users, err := qtx.ListUsers(c.Request.Context())
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, users)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

