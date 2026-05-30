package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type BusinessStaffHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewBusinessStaffHandler(queries *db.Queries, dbConn *sql.DB) *BusinessStaffHandler {
	return &BusinessStaffHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

type InviteStaffRequest struct {
	BusinessID string `json:"business_id" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Role       string `json:"role" binding:"required"` // manager, cashier, etc.
}

// InviteStaff allows a store owner to invite a user to join their staff.
func (h *BusinessStaffHandler) InviteStaff(c *gin.Context) {
	var req InviteStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ownerID := c.GetString("user_id")

	// Verify ownership of the business
	business, err := h.Queries.GetBusinessByID(c.Request.Context(), req.BusinessID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Business not found"})
		return
	}

	if business.OwnerID != ownerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the store owner can invite staff"})
		return
	}

	// Check if user exists
	user, err := h.Queries.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User with this email not found. They must create an account first."})
		return
	}

	// Create staff record
	token := uuid.New().String()
	_, err = h.Queries.CreateBusinessStaff(c.Request.Context(), db.CreateBusinessStaffParams{
		ID:              uuid.New().String(),
		BusinessID:      req.BusinessID,
		UserID:          user.ID,
		Role:            req.Role,
		Permissions:     pqtype.NullRawMessage{RawMessage: json.RawMessage("{}"), Valid: true},
		InvitedBy:       sql.NullString{String: ownerID, Valid: true},
		InvitationToken: sql.NullString{String: token, Valid: true},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create staff record: " + err.Error()})
		return
	}

	// In a real app, send an email with the token here
	c.JSON(http.StatusCreated, gin.H{
		"message": "Staff invited successfully",
		"token":   token,
	})
}

// ListStaff returns all staff members for a specific business.
func (h *BusinessStaffHandler) ListStaff(c *gin.Context) {
	businessID := c.Param("id")
	ownerID := c.GetString("user_id")

	// Verify ownership or admin status
	business, err := h.Queries.GetBusinessByID(c.Request.Context(), businessID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Business not found"})
		return
	}

	if business.OwnerID != ownerID && !c.GetBool("is_admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	staff, err := h.Queries.ListBusinessStaffByBusiness(c.Request.Context(), businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, staff)
}
