package handlers

import (
	"database/sql"
	"net/http"

	"ehub/backend/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReviewHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewReviewHandler(queries *db.Queries, dbConn *sql.DB) *ReviewHandler {
	return &ReviewHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	var req struct {
		TargetID   string `json:"target_id" binding:"required"`
		TargetType string `json:"target_type" binding:"required"`
		Rating     int32  `json:"rating" binding:"required,min=1,max=5"`
		Comment    string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		params := db.AddReviewParams{
			ID:         uuid.New().String(),
			TargetID:   req.TargetID,
			TargetType: req.TargetType,
			UserID:     userID,
			Rating:     req.Rating,
			Comment:    sql.NullString{String: req.Comment, Valid: req.Comment != ""},
		}

		review, err := qtx.AddReview(c.Request.Context(), params)
		if err != nil {
			return err
		}
		c.JSON(http.StatusCreated, review)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *ReviewHandler) GetReviewsByTarget(c *gin.Context) {
	targetID := c.Query("target_id")
	targetType := c.Query("target_type")
	
	if targetID == "" || targetType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_id and target_type required"})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		reviews, err := qtx.GetReviewsByTarget(c.Request.Context(), db.GetReviewsByTargetParams{
			TargetID:   targetID,
			TargetType: targetType,
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, reviews)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
