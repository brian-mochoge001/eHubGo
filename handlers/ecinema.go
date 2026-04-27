package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
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

func (h *CinemaHandler) ListMovieShowtimes(c *gin.Context) {
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
