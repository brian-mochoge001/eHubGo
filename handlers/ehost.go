package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type HostHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewHostHandler(queries *db.Queries, dbConn *sql.DB) *HostHandler {
	return &HostHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *HostHandler) ListProperties(c *gin.Context) {
	city := c.Query("city")
	priceStr := c.Query("max_price")
	roomsStr := c.Query("min_rooms")

	var minRooms sql.NullInt32
	if r, err := strconv.Atoi(roomsStr); err == nil {
		minRooms = sql.NullInt32{Int32: int32(r), Valid: true}
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		properties, err := qtx.ListProperties(c.Request.Context(), db.ListPropertiesParams{
			City:     sql.NullString{String: city, Valid: city != ""},
			MaxPrice: sql.NullString{String: priceStr, Valid: priceStr != ""},
			MinRooms: minRooms,
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, properties)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *HostHandler) GetProperty(c *gin.Context) {
	id := c.Param("id")
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		property, err := qtx.GetPropertyByID(c.Request.Context(), id)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, property)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
	}
}
