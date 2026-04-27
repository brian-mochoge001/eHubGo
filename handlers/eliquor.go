package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type LiquorHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewLiquorHandler(queries *db.Queries, dbConn *sql.DB) *LiquorHandler {
	return &LiquorHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *LiquorHandler) ListLiquorItems(c *gin.Context) {
	businessID := c.Query("business_id")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		var items []db.LiquorItem
		var err error

		if businessID != "" {
			items, err = qtx.ListLiquorItemsByBusiness(c.Request.Context(), businessID)
		} else {
			rows, err := qtx.ListLiquorItems(c.Request.Context())
			if err != nil {
				return err
			}
			c.JSON(http.StatusOK, rows)
			return nil
		}

		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, items)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
