package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type FoodHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewFoodHandler(queries *db.Queries, dbConn *sql.DB) *FoodHandler {
	return &FoodHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *FoodHandler) ListAllFoodItems(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		items, err := qtx.ListAllFoodItems(c.Request.Context())
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
