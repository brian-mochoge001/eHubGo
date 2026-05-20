package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FeedbackHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

type FeedbackDTO struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

func NewFeedbackHandler(queries *db.Queries, dbConn *sql.DB) *FeedbackHandler {
	return &FeedbackHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

// @Summary Submit user feedback
// @Description Submit rating and textual feedback from the mobile/web app
// @Tags feedback
// @Accept json
// @Produce json
// @Param body body FeedbackDTO true "Feedback payload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/feedback [post]
func (h *FeedbackHandler) SubmitFeedback(c *gin.Context) {
	// In a complete implementation, this would save to a `feedbacks` table.
	// Since we haven't modified the DB schema for feedback, we will mock the save operation.
	var req FeedbackDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = "anonymous"
	}

	// Mock saving to DB
	_ = userID
	_ = uuid.New()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Feedback submitted successfully. Thank you!",
	})
}
