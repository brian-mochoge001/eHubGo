package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type DeliveryHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewDeliveryHandler(queries *db.Queries, dbConn *sql.DB) *DeliveryHandler {
	return &DeliveryHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *DeliveryHandler) ListDeliveryOptions(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		services, err := qtx.ListServicesByType(c.Request.Context(), "delivery")
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
