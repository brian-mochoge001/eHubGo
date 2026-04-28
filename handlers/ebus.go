package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type BusHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewBusHandler(queries *db.Queries, dbConn *sql.DB) *BusHandler {
	return &BusHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *BusHandler) ListBusRoutes(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		routes, err := qtx.ListBusRoutes(c.Request.Context(), db.ListBusRoutesParams{})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, routes)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
