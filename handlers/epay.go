package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
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
