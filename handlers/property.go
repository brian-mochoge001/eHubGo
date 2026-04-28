package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PropertyHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewPropertyHandler(queries *db.Queries, dbConn *sql.DB) *PropertyHandler {
	return &PropertyHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

// ListProperties returns a list of available properties
func (h *PropertyHandler) ListProperties(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		properties, err := qtx.ListProperties(c.Request.Context(), db.ListPropertiesParams{})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, properties)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetProperty returns details of a single property
func (h *PropertyHandler) GetProperty(c *gin.Context) {
	id := c.Param("id")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		property, err := qtx.GetPropertyByID(c.Request.Context(), id)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, property)
		return nil
	})

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "property not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

// CreatePropertyListing allows a host/vendor to list a new property (Stay/Hotel)
func (h *PropertyHandler) CreatePropertyListing(c *gin.Context) {
	var req struct {
		BusinessID        string   `json:"business_id" binding:"required"`
		Title             string   `json:"title" binding:"required"`
		Description       string   `json:"description"`
		AddressID         string   `json:"address_id"`
		PricePerNight     float64  `json:"price_per_night" binding:"required"`
		Currency          string   `json:"currency" default:"Ksh"`
		NumberOfGuests    int32    `json:"number_of_guests"`
		NumberOfBedrooms  int32    `json:"number_of_bedrooms"`
		Type              string   `json:"type" binding:"required"` // apartment, house, etc.
		ImageUrls         []string `json:"image_urls"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)

		property, err := qtx.CreateProperty(c.Request.Context(), db.CreatePropertyParams{
			ID:               uuid.New().String(),
			BusinessID:       req.BusinessID,
			Title:            req.Title,
			Description:      sql.NullString{String: req.Description, Valid: req.Description != ""},
			AddressID:        sql.NullString{String: req.AddressID, Valid: req.AddressID != ""},
			PricePerNight:    fmt.Sprintf("%.2f", req.PricePerNight),
			Currency:         req.Currency,
			NumberOfGuests:   sql.NullInt32{Int32: req.NumberOfGuests, Valid: req.NumberOfGuests > 0},
			NumberOfBedrooms: sql.NullInt32{Int32: req.NumberOfBedrooms, Valid: req.NumberOfBedrooms > 0},
			Type:             db.PropertyTypeEnum(req.Type),
			ImageUrls:        req.ImageUrls,
		})
		if err != nil {
			return err
		}

		c.JSON(http.StatusCreated, property)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// BookProperty creates a booking for a stay
func (h *PropertyHandler) BookProperty(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var req struct {
		PropertyID   string    `json:"property_id" binding:"required"`
		CheckInDate  time.Time `json:"check_in_date" binding:"required"`
		CheckOutDate time.Time `json:"check_out_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)

		// 1. Get property details to calculate price
		prop, err := qtx.GetPropertyByID(c.Request.Context(), req.PropertyID)
		if err != nil {
			return err
		}

		// 2. Calculate total (simplified)
		days := int32(req.CheckOutDate.Sub(req.CheckInDate).Hours() / 24)
		if days <= 0 {
			return fmt.Errorf("check-out date must be after check-in date")
		}
		
		// In a real app, parse the numeric price
		// totalAmount := prop.PricePerNight * float64(days)

		// 3. Create booking
		booking, err := qtx.CreatePropertyBooking(c.Request.Context(), db.CreatePropertyBookingParams{
			ID:          uuid.New().String(),
			UserID:      userID,
			PropertyID:  prop.ID,
			CheckInDate: req.CheckInDate,
			CheckOutDate: req.CheckOutDate,
			TotalAmount: prop.PricePerNight, // Placeholder calculation
			Currency:    prop.Currency,
		})
		if err != nil {
			return err
		}

		c.JSON(http.StatusCreated, booking)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// SearchProperties allows users to find stays near a specific location
func (h *PropertyHandler) SearchProperties(c *gin.Context) {
	var params struct {
		Longitude float64 `form:"longitude" binding:"required"`
		Latitude  float64 `form:"latitude" binding:"required"`
		Radius    float64 `form:"radius,default=5000"` // default 5km
	}

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		properties, err := qtx.SearchPropertiesByLocation(c.Request.Context(), db.SearchPropertiesByLocationParams{
			StMakepoint:   params.Longitude,
			StMakepoint_2: params.Latitude,
			StDwithin:     params.Radius,
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, properties)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
