package handlers

import (
	"database/sql"
	"net/http"
	"strings"

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
		Name            string `json:"name" binding:"required"`
		Description     string `json:"description"`
		MiniserviceType string `json:"miniservice_type" binding:"required"` // liquor, hotel, etc.
		LogoURL         string `json:"logo_url"`
		BannerURL       string `json:"banner_url"`
		PhoneNumber     string `json:"phone_number"`
		Email           string `json:"email"`
		AddressID       string `json:"address_id"`
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

// ListBusinesses returns all businesses, optionally filtered by type or status
func (h *BusinessHandler) ListBusinesses(c *gin.Context) {
	businessType := c.Query("type")
	ownerID := c.Query("owner_id")
	status := c.Query("status")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		
		if status != "" {
			rows, err := qtx.ListBusinessesByStatus(c.Request.Context(), db.NullBusinessVerificationStatus{
				BusinessVerificationStatus: db.BusinessVerificationStatus(status),
				Valid:                     true,
			})
			if err == nil {
				c.JSON(http.StatusOK, rows)
				return nil
			}
		}

		if ownerID != "" {
			biz, err := qtx.GetBusinessesByOwnerID(c.Request.Context(), ownerID)
			if err == nil {
				c.JSON(http.StatusOK, biz)
				return nil
			}
		}

		businesses, err := qtx.ListAllBusinesses(c.Request.Context())
		if err != nil {
			return err
		}

		if businessType != "" {
			var filtered []db.ListAllBusinessesRow
			for _, b := range businesses {
				if b.MiniserviceType == businessType {
					filtered = append(filtered, b)
				}
			}
			c.JSON(http.StatusOK, filtered)
			return nil
		}

		c.JSON(http.StatusOK, businesses)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// UpdateBusinessStatus allows admins to approve/reject/suspend a business
func (h *BusinessHandler) UpdateBusinessStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status db.BusinessVerificationStatus `json:"status" binding:"required"`
		Reason string                     `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Permission check: only admin/staff can update status
	rolesVal, _ := c.Get("user_roles")
	roles := rolesVal.(string)
	isAdmin := false
	for _, r := range []string{"admin", "executive_admin", "staff"} {
		if strings.Contains(roles, r) {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins or staff can update business status"})
		return
	}

	err := h.DB.QueryRow("SELECT id FROM businesses WHERE id = $1", id).Scan(&id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "business not found"})
		return
	}

	err = WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		business, err := qtx.UpdateBusinessStatus(c.Request.Context(), db.UpdateBusinessStatusParams{
			ID:                 id,
			VerificationStatus: db.NullBusinessVerificationStatus{
				BusinessVerificationStatus: req.Status,
				Valid:                      true,
			},
		})
		if err != nil {
			return err
		}

		c.JSON(http.StatusOK, business)
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
