package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type ServicesExtraHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewServicesExtraHandler(queries *db.Queries, dbConn *sql.DB) *ServicesExtraHandler {
	return &ServicesExtraHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

// eBus
func (h *ServicesExtraHandler) ListBusRoutes(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		routes, err := qtx.ListBusRoutes(c.Request.Context())
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

// eCinema
func (h *ServicesExtraHandler) ListNowPlayingMovies(c *gin.Context) {
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

func (h *ServicesExtraHandler) ListComingSoonMovies(c *gin.Context) {
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

func (h *ServicesExtraHandler) ListMovieShowtimes(c *gin.Context) {
	movieID := c.Param("id")
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		showtimes, err := qtx.ListMovieShowtimesByMovie(c.Request.Context(), movieID)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, showtimes)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// eFlights
func (h *ServicesExtraHandler) ListFlights(c *gin.Context) {
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

// eJobs
func (h *ServicesExtraHandler) ListJobs(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		jobs, err := qtx.ListActiveJobs(c.Request.Context())
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, jobs)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// eTravel
func (h *ServicesExtraHandler) ListTours(c *gin.Context) {
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
