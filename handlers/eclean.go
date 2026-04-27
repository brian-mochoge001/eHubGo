package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type CleanHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewCleanHandler(queries *db.Queries, dbConn *sql.DB) *CleanHandler {
	return &CleanHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *CleanHandler) ListCleaningServices(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		services, err := qtx.ListServicesByType(c.Request.Context(), "cleaning")
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
