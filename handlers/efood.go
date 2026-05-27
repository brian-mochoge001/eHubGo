package handlers

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"ehubgo/db"

	"github.com/gin-gonic/gin"
)

type FoodHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewFoodHandler(queries *db.Queries, dbConn *sql.DB) *FoodHandler {
	return &FoodHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *FoodHandler) ListAllFoodItems(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		items, err := qtx.ListAllFoodItems(c.Request.Context())
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

func (h *FoodHandler) GetFoodItem(c *gin.Context) {
	id := c.Param("id")
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		item, err := qtx.GetFoodItemByID(c.Request.Context(), id)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, item)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Food item not found"})
	}
}

func (h *FoodHandler) ListFoodByCategory(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		h.ListAllFoodItems(c)
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		items, err := qtx.ListFoodItemsByCategory(c.Request.Context(), category)
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

func (h *FoodHandler) GetNearbyMotorbikeDrivers(c *gin.Context) {
	var params struct {
		Longitude float64 `form:"longitude" binding:"required"`
		Latitude  float64 `form:"latitude" binding:"required"`
		Radius    float64 `form:"radius,default=5000"` // default 5km
		Limit     int32   `form:"limit,default=10"`
	}

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		drivers, err := qtx.GetNearbyMotorbikeDrivers(c.Request.Context(), db.GetNearbyMotorbikeDriversParams{
			StMakepoint:   params.Longitude,
			StMakepoint_2: params.Latitude,
			StDwithin:     params.Radius,
			Limit:         params.Limit,
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

func parseDistanceMeters(distance interface{}) float64 {
	switch d := distance.(type) {
	case float64:
		return d
	case float32:
		return float64(d)
	case int:
		return float64(d)
	case int64:
		return float64(d)
	case int32:
		return float64(d)
	case string:
		if f, err := strconv.ParseFloat(d, 64); err == nil {
			return f
		}
	case []byte:
		if f, err := strconv.ParseFloat(string(d), 64); err == nil {
			return f
		}
	}
	return 0
}

func (h *FoodHandler) EstimateDelivery(c *gin.Context) {
	var req struct {
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
		Radius    float64 `json:"radius"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Radius <= 0 {
		req.Radius = 60000 // 60km max food delivery coverage
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		drivers, err := qtx.GetNearbyMotorbikeDrivers(c.Request.Context(), db.GetNearbyMotorbikeDriversParams{
			StMakepoint:   req.Longitude,
			StMakepoint_2: req.Latitude,
			StDwithin:     req.Radius,
			Limit:         10,
		})
		if err != nil {
			return err
		}

		nearestDistance := math.Inf(1)
		var nearestDriverLat, nearestDriverLng float64
		hasNearest := false

		for _, driver := range drivers {
			distance := parseDistanceMeters(driver.Distance)
			if distance > 0 && distance < nearestDistance {
				nearestDistance = distance
				nearestDriverLat = ParseCoordinate(driver.Latitude)
				nearestDriverLng = ParseCoordinate(driver.Longitude)
				hasNearest = true
			}
		}

		if math.IsInf(nearestDistance, 1) {
			nearestDistance = req.Radius
		}

		estimatedMinutes := int(math.Ceil((nearestDistance / 1000.0) / 30.0 * 60.0))
		if estimatedMinutes < 10 {
			estimatedMinutes = 10
		}

		var polylineStr string
		if hasNearest {
			polylineStr = GetPointToPointRoute(c.Request.Context(), nearestDriverLat, nearestDriverLng, req.Latitude, req.Longitude)
		}

		available := len(drivers) > 0 && estimatedMinutes <= 120
		c.JSON(http.StatusOK, gin.H{
			"available":                 available,
			"driver_count":              len(drivers),
			"nearest_driver_distance_m": nearestDistance,
			"estimated_minutes":         estimatedMinutes,
			"polyline":                  polylineStr,
			"max_minutes_allowed":       120,
			"note":                      "Only motorbike drivers are eligible for food delivery.",
		})

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *FoodHandler) AssignDelivery(c *gin.Context) {
	var req struct {
		OrderID     string  `json:"order_id" binding:"required"`
		DriverID    string  `json:"driver_id" binding:"required"`
		DeliveryFee float64 `json:"delivery_fee" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		order, err := qtx.AssignDriverToOrder(c.Request.Context(), db.AssignDriverToOrderParams{
			ID:          req.OrderID,
			DriverID:    sql.NullString{String: req.DriverID, Valid: true},
			DeliveryFee: fmt.Sprintf("%.2f", req.DeliveryFee),
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, order)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
