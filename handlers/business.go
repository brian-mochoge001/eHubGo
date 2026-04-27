package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BusinessHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewBusinessHandler(queries *db.Queries, dbConn *sql.DB) *BusinessHandler {
	return &BusinessHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

// RegisterBusiness allows a user to open a new "stall" (business) in the mall
func (h *BusinessHandler) RegisterBusiness(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var req struct {
		Name           string `json:"name" binding:"required"`
		Description    string `json:"description"`
		MiniserviceType string `json:"miniservice_type" binding:"required"` // liquor, hotel, etc.
		LogoURL        string `json:"logo_url"`
		BannerURL      string `json:"banner_url"`
		PhoneNumber    string `json:"phone_number"`
		Email          string `json:"email"`
		AddressID      string `json:"address_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)

		// 1. Create the business
		business, err := qtx.CreateBusiness(c.Request.Context(), db.CreateBusinessParams{
			ID:              uuid.New().String(),
			OwnerID:         userID,
			Name:            req.Name,
			Description:     sql.NullString{String: req.Description, Valid: req.Description != ""},
			LogoUrl:         sql.NullString{String: req.LogoURL, Valid: req.LogoURL != ""},
			BannerUrl:       sql.NullString{String: req.BannerURL, Valid: req.BannerURL != ""},
			MiniserviceType: req.MiniserviceType,
			AddressID:       sql.NullString{String: req.AddressID, Valid: req.AddressID != ""},
			PhoneNumber:     sql.NullString{String: req.PhoneNumber, Valid: req.PhoneNumber != ""},
			Email:           sql.NullString{String: req.Email, Valid: req.Email != ""},
		})
		if err != nil {
			return err
		}

		// 2. Ensure user has 'vendor' role
		_, _ = qtx.AssignRoleToUser(c.Request.Context(), db.AssignRoleToUserParams{
			UserID: userID,
			Role:   db.UserRoleTypeVendor,
		})

		c.JSON(http.StatusCreated, business)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetMyMall returns all businesses owned by the current user
func (h *BusinessHandler) GetMyMall(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		businesses, err := qtx.GetBusinessesByOwnerID(c.Request.Context(), userID)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, businesses)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// ListBusinesses returns all businesses, optionally filtered by type
func (h *BusinessHandler) ListBusinesses(c *gin.Context) {
	businessType := c.Query("type")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		var businesses []db.Business
		var err error

		if businessType != "" {
			businesses, err = qtx.ListBusinessesByType(c.Request.Context(), businessType)
		} else {
			// If no type, we might want to list all approved ones
			// For now, reuse ListBusinessesByType with empty or add a ListAllApproved
			businesses, err = qtx.ListBusinessesByType(c.Request.Context(), businessType)
		}

		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, businesses)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetBusinessProfile returns details of a specific business
func (h *BusinessHandler) GetBusinessProfile(c *gin.Context) {
	id := c.Param("id")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		business, err := qtx.GetBusinessByID(c.Request.Context(), id)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, business)
		return nil
	})

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "business not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}
