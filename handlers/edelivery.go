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
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
		})
	} else {
		// Long distance: Suggest courier companies
		c.JSON(http.StatusOK, gin.H{
			"type":    "courier_required",
			"message": "Distance requires specialized courier service. Please select from available courier partners.",
		})
	}
}
