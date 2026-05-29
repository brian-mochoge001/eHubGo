package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ComplaintHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewComplaintHandler(queries *db.Queries, dbConn *sql.DB) *ComplaintHandler {
	return &ComplaintHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

// SubmitComplaint allows a user or driver to submit a complaint
func (h *ComplaintHandler) SubmitComplaint(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	var req struct {
		TargetID   string `json:"target_id" binding:"required"`
		TargetType string `json:"target_type" binding:"required"` // 'user' or 'driver'
		Reason     string `json:"reason" binding:"required"`
		Details    string `json:"details"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(c.Request.Context(), `
			INSERT INTO complaints (id, reporter_id, target_id, target_type, reason, details)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New().String(), userID, req.TargetID, req.TargetType, req.Reason, sql.NullString{String: req.Details, Valid: req.Details != ""})
		return err
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Complaint submitted successfully"})
}
