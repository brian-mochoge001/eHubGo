package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type JobsHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewJobsHandler(queries *db.Queries, dbConn *sql.DB) *JobsHandler {
	return &JobsHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *JobsHandler) ListJobs(c *gin.Context) {
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

func (h *JobsHandler) GetCompanyProfile(c *gin.Context) {
	businessID := c.Param("business_id")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		company, err := qtx.GetBusinessByID(c.Request.Context(), businessID)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, company)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
	}
}
