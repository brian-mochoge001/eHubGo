package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"math/rand"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CinemaHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewCinemaHandler(queries *db.Queries, dbConn *sql.DB) *CinemaHandler {
	return &CinemaHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *CinemaHandler) ListNowPlayingMovies(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		movies, err := qtx.ListNowPlayingMovies(c.Request.Context())
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, movies)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *CinemaHandler) ListComingSoonMovies(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		movies, err := qtx.ListComingSoonMovies(c.Request.Context())
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, movies)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *CinemaHandler) GetMovieDetails(c *gin.Context) {
	movieID := c.Param("id")
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		movie, err := qtx.GetMovieDetails(c.Request.Context(), movieID)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, movie)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
	}
}

func (h *CinemaHandler) ListRefreshments(c *gin.Context) {
	businessID := c.Query("business_id")
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		refreshments, err := qtx.ListRefreshmentsByCinema(c.Request.Context(), businessID)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, refreshments)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *CinemaHandler) BookTicket(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	var req struct {
		ShowtimeID     string   `json:"showtime_id" binding:"required"`
		SeatNumber     string   `json:"seat_number" binding:"required"`
		RefreshmentIDs []string `json:"refreshment_ids"`
		TotalAmount    float64  `json:"total_amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketNum := fmt.Sprintf("TCK-%d", rand.Intn(1000000))
	qrCode := fmt.Sprintf("QR-%s", ticketNum)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		ticket, err := qtx.CreateTicket(c.Request.Context(), db.CreateTicketParams{
			ID:             uuid.New().String(),
			UserID:         userID,
			ShowtimeID:     req.ShowtimeID,
			SeatNumber:     req.SeatNumber,
			TicketNumber:   ticketNum,
			QrCodeData:     qrCode,
			RefreshmentIds: req.RefreshmentIDs,
			TotalAmount:    fmt.Sprintf("%.2f", req.TotalAmount),
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
