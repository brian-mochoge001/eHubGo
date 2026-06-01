package handlers

import (
	"database/sql"
	"net/http"
	"os"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PayHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewPayHandler(queries *db.Queries, dbConn *sql.DB) *PayHandler {
	return &PayHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *PayHandler) GetWalletBalance(c *gin.Context) {
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

// ProcessMockPayment simulates a successful payment response
func (h *PayHandler) ProcessMockPayment(c *gin.Context) {
	if os.Getenv("PAYMENT_MODE") != "mock" && os.Getenv("GO_ENV") == "production" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Mock payments disabled"})
		return
	}

	var req struct {
		OrderID string `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(string)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		// 1. Update order status to 'paid'
		_, err := tx.ExecContext(c.Request.Context(), "UPDATE orders SET status = 'paid', updated_at = CURRENT_TIMESTAMP WHERE id = $1", req.OrderID)
		if err != nil {
			return err
		}

		// 2. Record transaction
		var totalAmount string
		err = tx.QueryRowContext(c.Request.Context(), "SELECT total_amount FROM orders WHERE id = $1", req.OrderID).Scan(&totalAmount)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(c.Request.Context(), `
			INSERT INTO transactions (id, user_id, order_id, amount, status)
			VALUES ($1, $2, $3, $4, 'completed')
		`, uuid.New().String(), userID, req.OrderID, totalAmount)
		
		return err
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mock payment processed successfully"})
}
