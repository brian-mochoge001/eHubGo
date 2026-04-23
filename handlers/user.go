package handlers

import (
	"database/sql"
	"net/http"

	"ehub/backend/db"
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
