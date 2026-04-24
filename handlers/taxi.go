package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaxiHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewTaxiHandler(queries *db.Queries, dbConn *sql.DB) *TaxiHandler {
	return &TaxiHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

// UpdateLocation allows drivers to ping their GPS coordinates
func (h *TaxiHandler) UpdateLocation(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var req struct {
		Longitude float64 `json:"longitude" binding:"required"`
		Latitude  float64 `json:"latitude" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		driver, err := qtx.UpdateDriverLocation(c.Request.Context(), db.UpdateDriverLocationParams{
			UserID: userID,
			Lng:    req.Longitude,
			Lat:    req.Latitude,
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, driver)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// UpdateStatus allows drivers to go online/offline
func (h *TaxiHandler) UpdateStatus(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var req struct {
		Status string `json:"status" binding:"required"` // online, offline, busy
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		driver, err := qtx.UpdateDriverStatus(c.Request.Context(), db.UpdateDriverStatusParams{
			UserID: userID,
			Status: req.Status,
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, driver)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetNearbyDrivers allows users to see available taxis on the map
func (h *TaxiHandler) GetNearbyDrivers(c *gin.Context) {
	var params struct {
		Longitude float64 `form:"longitude" binding:"required"`
		Latitude  float64 `form:"latitude" binding:"required"`
		Radius    float64 `form:"radius,default=2000"` // default 2km
		Limit     int32   `form:"limit,default=10"`
	}

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		// Check for correct query name and parameters
		drivers, err := qtx.GetNearbyDrivers(c.Request.Context(), db.GetNearbyDriversParams{
			Lng:        params.Longitude,
			Lat:        params.Latitude,
			Radius:     params.Radius,
			LimitCount: params.Limit,
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, drivers)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// RequestRide creates a new taxi trip request
func (h *TaxiHandler) RequestRide(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var req struct {
		PickupLng  float64 `json:"pickup_lng" binding:"required"`
		PickupLat  float64 `json:"pickup_lat" binding:"required"`
		DropoffLng float64 `json:"dropoff_lng" binding:"required"`
		DropoffLat float64 `json:"dropoff_lat" binding:"required"`
		Amount     float64 `json:"amount" binding:"required"`
		Currency   string  `json:"currency" default:"Ksh"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		trip, err := qtx.CreateTaxiTrip(c.Request.Context(), db.CreateTaxiTripParams{
			ID:         uuid.New().String(),
			UserID:     userID,
			PickupLng:  req.PickupLng,
			PickupLat:  req.PickupLat,
			DropoffLng: req.DropoffLng,
			DropoffLat: req.DropoffLat,
			TotalAmount: fmt.Sprintf("%.2f", req.Amount),
			Currency:   req.Currency,
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusCreated, trip)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
