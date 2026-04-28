package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HealthHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewHealthHandler(queries *db.Queries, dbConn *sql.DB) *HealthHandler {
	return &HealthHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *HealthHandler) ListHealthServices(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		services, err := qtx.ListServicesByType(c.Request.Context(), "health")
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

func (h *HealthHandler) ListPharmacyItems(c *gin.Context) {
	businessID := c.Query("business_id")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		var items []db.PharmacyItem
		var err error

		if businessID != "" {
			items, err = qtx.ListPharmacyItemsByBusiness(c.Request.Context(), businessID)
		} else {
			items, err = qtx.ListPharmacyItems(c.Request.Context())
		}

		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, items)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *HealthHandler) ListDoctors(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		doctors, err := qtx.ListDoctors(c.Request.Context())
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, doctors)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *HealthHandler) BookAppointment(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	var req struct {
		DoctorID        string    `json:"doctor_id" binding:"required"`
		AppointmentTime time.Time `json:"appointment_time" binding:"required"`
		Notes           string    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		appointment, err := qtx.CreateAppointment(c.Request.Context(), db.CreateAppointmentParams{
			ID:              uuid.New().String(),
			PatientID:       userID,
			DoctorID:        req.DoctorID,
			AppointmentTime: req.AppointmentTime,
			Notes:           sql.NullString{String: req.Notes, Valid: req.Notes != ""},
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusCreated, appointment)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
