package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type TravelHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewTravelHandler(queries *db.Queries, dbConn *sql.DB) *TravelHandler {
	return &TravelHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *TravelHandler) ListTours(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		tours, err := qtx.ListTours(c.Request.Context())
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, tours)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
