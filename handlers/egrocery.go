package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type GroceryHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewGroceryHandler(queries *db.Queries, dbConn *sql.DB) *GroceryHandler {
	return &GroceryHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *GroceryHandler) ListGroceryItems(c *gin.Context) {
	businessID := c.Query("business_id")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		var items []db.GroceryItem
		var err error

		if businessID != "" {
			items, err = qtx.ListGroceryItemsByBusiness(c.Request.Context(), businessID)
		} else {
			rows, err := qtx.ListGroceryItems(c.Request.Context())
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
