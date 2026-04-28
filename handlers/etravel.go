package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func (h *TravelHandler) ListBusRoutes(c *gin.Context) {
	origin := c.Query("origin")
	destination := c.Query("destination")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		routes, err := qtx.ListBusRoutes(c.Request.Context(), db.ListBusRoutesParams{
			Origin:      sql.NullString{String: origin, Valid: origin != ""},
			Destination: sql.NullString{String: destination, Valid: destination != ""},
		})
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

func (h *TravelHandler) BookBusTicket(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	var req struct {
		ShowtimeID string  `json:"showtime_id" binding:"required"`
		SeatNumber string  `json:"seat_number" binding:"required"`
		TotalAmount float64 `json:"total_amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		ticket, err := qtx.CreateBusTicket(c.Request.Context(), db.CreateBusTicketParams{
			ID:           uuid.New().String(),
			UserID:       userID,
			ShowtimeID:   req.ShowtimeID,
			SeatNumber:   req.SeatNumber,
			TicketNumber: uuid.New().String()[:8], // Simple ticket number
			QrCodeData:   "qr_" + uuid.New().String(),
			TotalAmount:  fmt.Sprintf("%.2f", req.TotalAmount),
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusCreated, ticket)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
