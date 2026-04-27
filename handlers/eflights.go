package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type FlightsHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewFlightsHandler(queries *db.Queries, dbConn *sql.DB) *FlightsHandler {
	return &FlightsHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *FlightsHandler) ListFlights(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		flights, err := qtx.ListFlights(c.Request.Context())
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, flights)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
