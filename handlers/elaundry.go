package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type LaundryHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewLaundryHandler(queries *db.Queries, dbConn *sql.DB) *LaundryHandler {
	return &LaundryHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *LaundryHandler) ListLaundryServices(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		services, err := qtx.ListServicesByType(c.Request.Context(), "laundry")
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, services)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
