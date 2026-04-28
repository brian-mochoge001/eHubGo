package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func (h *FlightsHandler) BookFlightTicket(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	var req struct {
		FlightID   string  `json:"flight_id" binding:"required"`
		SeatNumber string  `json:"seat_number" binding:"required"`
		Tier       string  `json:"tier" binding:"required"` // e.g., 'economy', 'business', 'first'
		TotalAmount float64 `json:"total_amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		ticket, err := qtx.CreateFlightTicket(c.Request.Context(), db.CreateFlightTicketParams{
			ID:           uuid.New().String(),
			UserID:       userID,
			ShowtimeID:   req.FlightID, // Reusing showtime_id for flight reference
			SeatNumber:   req.SeatNumber,
			TicketNumber: uuid.New().String()[:8],
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
