package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"ehubgo/cache"
	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CinemaHandler struct {
	Queries *db.Queries
	DB      *sql.DB
	Cache   cache.Store
}

func NewCinemaHandler(queries *db.Queries, dbConn *sql.DB, c cache.Store) *CinemaHandler {
	return &CinemaHandler{
		Queries: queries,
		DB:      dbConn,
		Cache:   c,
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

func (h *CinemaHandler) ListMovieShowtimes(c *gin.Context) {
	movieID := c.Query("movie_id")
	cacheKey := fmt.Sprintf("cinema:showtimes:%s", movieID)

	var showtimes []db.ListMovieShowtimesRow
	found, err := h.Cache.GetJSON(c.Request.Context(), cacheKey, &showtimes)
	if err == nil && found {
		c.JSON(http.StatusOK, showtimes)
		return
	}

	err = WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		var err error
		if movieID != "" {
			showtimes, err = qtx.ListMovieShowtimes(c.Request.Context(), movieID)
		} else {
			showtimes = []db.ListMovieShowtimesRow{}
		}
		return err
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = h.Cache.SetJSON(c.Request.Context(), cacheKey, showtimes, 15*time.Minute)
	c.JSON(http.StatusOK, showtimes)
}

func (h *CinemaHandler) CreateMovieShowtime(c *gin.Context) {
    // Note: Assuming a similar structure to other handlers
    var req struct {
		MovieID string `json:"movie_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
    
    // ... logic ...
    
    // After successful creation:
    _ = h.Cache.Delete(c.Request.Context(), fmt.Sprintf("cinema:showtimes:%s", req.MovieID))
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

	ticketNum := fmt.Sprintf("TCK-%s", uuid.New().String()[:8])
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
