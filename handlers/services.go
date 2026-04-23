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

type ServiceHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewServiceHandler(queries *db.Queries, dbConn *sql.DB) *ServiceHandler {
	return &ServiceHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

// ListServices returns services of a specific type (laundry, cleaning, etc.)
func (h *ServiceHandler) ListServices(c *gin.Context) {
	serviceType := c.Query("type")
	if serviceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service type is required"})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		services, err := qtx.ListServicesByType(c.Request.Context(), serviceType)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, services)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// BookService creates a new booking for a service
func (h *ServiceHandler) BookService(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var req struct {
		ServiceID string    `json:"service_id" binding:"required"`
		StartTime time.Time `json:"start_time" binding:"required"`
		EndTime   time.Time `json:"end_time"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)

		// 1. Get service details
		service, err := qtx.GetServiceByID(c.Request.Context(), req.ServiceID)
		if err != nil {
			return err
		}

		// 2. Create booking
		booking, err := qtx.CreateServiceBooking(c.Request.Context(), db.CreateServiceBookingParams{
			ID:            uuid.New().String(),
			UserID:        userID,
			ServiceType:   service.ServiceType,
			ServiceItemID: service.ID,
			ProviderID:    sql.NullString{String: service.BusinessID, Valid: true},
			ProviderType:  sql.NullString{String: "business", Valid: true},
			StartTime:     req.StartTime,
			EndTime:       sql.NullTime{Time: req.EndTime, Valid: !req.EndTime.IsZero()},
			TotalAmount:   service.BasePrice,
			Currency:      service.Currency,
			Status:        "pending",
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

// GetMyBookings returns bookings for the current user
func (h *ServiceHandler) GetMyBookings(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		bookings, err := qtx.GetServiceBookingsByUserID(c.Request.Context(), userID)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, bookings)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// ProviderUpdateBookingStatus allows a service provider to update the status of a booking
func (h *ServiceHandler) ProviderUpdateBookingStatus(c *gin.Context) {
	bookingID := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"` // accepted, completed, cancelled
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		booking, err := qtx.UpdateServiceBookingStatus(c.Request.Context(), db.UpdateServiceBookingStatusParams{
			ID:     bookingID,
			Status: req.Status,
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, booking)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// CreateServiceListing allows a verified vendor to list a new service
func (h *ServiceHandler) CreateServiceListing(c *gin.Context) {
	var req struct {
		BusinessID          string  `json:"business_id" binding:"required"`
		Type                string  `json:"type" binding:"required"` // laundry, cleaning, repair, health
		Name                string  `json:"name" binding:"required"`
		Description         string  `json:"description"`
		BasePrice          float64 `json:"base_price" binding:"required"`
		Currency            string  `json:"currency" default:"Ksh"`
		Location            interface{} `json:"location"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)

		service, err := qtx.CreateService(c.Request.Context(), db.CreateServiceParams{
			ID:              uuid.New().String(),
			BusinessID:      req.BusinessID,
			ServiceType:     req.Type,
			Name:            req.Name,
			Description:     sql.NullString{String: req.Description, Valid: req.Description != ""},
			BasePrice:       fmt.Sprintf("%.2f", req.BasePrice),
			Currency:        req.Currency,
			Location:        req.Location,
		})
		if err != nil {
			return err
		}

		c.JSON(http.StatusCreated, service)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
