package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type C2CHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewC2CHandler(queries *db.Queries, dbConn *sql.DB) *C2CHandler {
	return &C2CHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

// ListC2CListings returns all available C2C items
func (h *C2CHandler) ListC2CListings(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		listings, err := qtx.ListC2CListings(c.Request.Context())
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, listings)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// CreateC2CListing allows a user to list an item for sale (Second-hand)
func (h *C2CHandler) CreateC2CListing(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var req struct {
		Title        string   `json:"title" binding:"required"`
		Description  string   `json:"description"`
		Price        float64  `json:"price" binding:"required"`
		Currency     string   `json:"currency" default:"Ksh"`
		IsNegotiable bool     `json:"is_negotiable"`
		Location     string   `json:"location"`
		ImageUrls    []string `json:"image_urls"`
		Condition    string   `json:"condition" default:"used"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)

		seller, err := qtx.GetC2CSellerByUserID(c.Request.Context(), userID)
		if err != nil {
			if err == sql.ErrNoRows {
				seller, err = qtx.CreateC2CSeller(c.Request.Context(), db.CreateC2CSellerParams{
					ID:     uuid.New().String(),
					UserID: userID,
				})
				if err != nil {
					return err
				}
				_ = qtx.AssignRoleToUser(c.Request.Context(), db.AssignRoleToUserParams{
					UserID: userID,
					Role:   db.UserRoleTypeC2cSeller,
				})
			} else {
				return err
			}
		}

		listing, err := qtx.CreateC2CListing(c.Request.Context(), db.CreateC2CListingParams{
			ID:           uuid.New().String(),
			SellerID:     seller.ID,
			Title:        req.Title,
			Description:  sql.NullString{String: req.Description, Valid: req.Description != ""},
			Price:        fmt.Sprintf("%.2f", req.Price),
			Currency:     req.Currency,
			IsNegotiable: sql.NullBool{Bool: req.IsNegotiable, Valid: true},
			Location:     sql.NullString{String: req.Location, Valid: req.Location != ""},
			Condition:    sql.NullString{String: req.Condition, Valid: req.Condition != ""},
			ImageUrls:    req.ImageUrls,
			Status:       "available",
		})
		if err != nil {
			return err
		}

		c.JSON(http.StatusCreated, listing)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetC2CListing returns details for a single C2C item
func (h *C2CHandler) GetC2CListing(c *gin.Context) {
	id := c.Param("id")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		listing, err := qtx.GetC2CListingByID(c.Request.Context(), id)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, listing)
		return nil
	})

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}
