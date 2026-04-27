package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type RepairHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewRepairHandler(queries *db.Queries, dbConn *sql.DB) *RepairHandler {
	return &RepairHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *RepairHandler) ListRepairServices(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		services, err := qtx.ListServicesByType(c.Request.Context(), "repair")
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
