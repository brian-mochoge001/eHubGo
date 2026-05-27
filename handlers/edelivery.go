package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type DeliveryHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewDeliveryHandler(queries *db.Queries, dbConn *sql.DB) *DeliveryHandler {
	return &DeliveryHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *DeliveryHandler) ListDeliveryOptions(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		services, err := qtx.ListServicesByType(c.Request.Context(), "delivery")
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

// CalculateDeliveryQuote calculates price based on distance: < 25km (Internal) or >= 25km (External Courier)
func (h *DeliveryHandler) CalculateDeliveryQuote(c *gin.Context) {
	var req struct {
		Distance         float64 `json:"distance" binding:"required"`
		IsPeakHour       bool    `json:"is_peak_hour"`
		WeatherSurcharge float64 `json:"weather_surcharge"`
		PickupLat        float64 `json:"pickup_lat"`
		PickupLng        float64 `json:"pickup_lng"`
		DropoffLat       float64 `json:"dropoff_lat"`
		DropoffLng       float64 `json:"dropoff_lng"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var polylineStr string
	if req.PickupLat != 0 && req.DropoffLat != 0 {
		polylineStr = GetPointToPointRoute(c.Request.Context(), req.PickupLat, req.PickupLng, req.DropoffLat, req.DropoffLng)
	}

	if req.Distance < 25.0 {
		// Use internal system
		baseFee := 100.0
		ratePerKm := 30.0
		peakMultiplier := 1.0
		if req.IsPeakHour {
			peakMultiplier = 1.2
		}

		price := CalculatePrice(baseFee, ratePerKm, req.Distance, peakMultiplier, req.WeatherSurcharge)
		c.JSON(http.StatusOK, gin.H{
			"type":            "internal",
			"estimated_price": price,
			"currency":        "Ksh",
			"polyline":        polylineStr,
		})
	} else {
		// Long distance: Suggest courier companies
		c.JSON(http.StatusOK, gin.H{
			"type":     "courier_required",
			"message":  "Distance requires specialized courier service. Please select from available courier partners.",
			"polyline": polylineStr,
		})
	}
}

