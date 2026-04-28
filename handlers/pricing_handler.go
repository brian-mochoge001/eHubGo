package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type PricingHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewPricingHandler(queries *db.Queries, dbConn *sql.DB) *PricingHandler {
	return &PricingHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *PricingHandler) GetPriceEstimate(c *gin.Context) {
	var req struct {
		ServiceType      string  `json:"service_type" binding:"required"` // 'taxi' or 'delivery'
		Distance         float64 `json:"distance" binding:"required"`
		IsPeakHour       bool    `json:"is_peak_hour"`
		WeatherSurcharge float64 `json:"weather_surcharge"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default parameters
	baseFee := 50.0
	ratePerKm := 20.0
	peakMultiplier := 1.0
	if req.IsPeakHour {
		peakMultiplier = 1.5
	}

	price := CalculatePrice(baseFee, ratePerKm, req.Distance, peakMultiplier, req.WeatherSurcharge)

	c.JSON(http.StatusOK, gin.H{
		"estimated_price": price,
		"currency":        "Ksh",
	})
}
