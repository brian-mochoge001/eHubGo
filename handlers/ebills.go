package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type BillsHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewBillsHandler(queries *db.Queries, dbConn *sql.DB) *BillsHandler {
	return &BillsHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *BillsHandler) ListUserBills(c *gin.Context) {
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
