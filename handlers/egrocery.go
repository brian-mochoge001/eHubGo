package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type GroceryHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewGroceryHandler(queries *db.Queries, dbConn *sql.DB) *GroceryHandler {
	return &GroceryHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *GroceryHandler) SearchGroceryStores(c *gin.Context) {
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
		stores, err := qtx.SearchGroceryStoresByLocation(c.Request.Context(), db.SearchGroceryStoresByLocationParams{
			StMakepoint:   params.Longitude,
			StMakepoint_2: params.Latitude,
			StDwithin:     params.Radius,
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, stores)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *GroceryHandler) CalculateGroceryDeliveryQuote(c *gin.Context) {
	var req struct {
		Distance   float64 `json:"distance" binding:"required"`
		BusinessID string  `json:"business_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simplification: In a real app, we check if the business offers their own delivery
	// Here we default to our pricing system for distance > 1km
	if req.Distance <= 1.0 {
		c.JSON(http.StatusOK, gin.H{"estimated_price": 0.0, "currency": "Ksh", "type": "free"})
		return
	}

	// Use our pricing system
	price := CalculatePrice(50.0, 20.0, req.Distance, 1.0, 0.0)
	c.JSON(http.StatusOK, gin.H{
		"estimated_price": price,
		"currency":        "Ksh",
		"type":            "platform_delivery",
	})
}

func (h *GroceryHandler) ListGroceryItems(c *gin.Context) {
	businessID := c.Query("business_id")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		var items []db.GroceryItem
		var err error

		if businessID != "" {
			items, err = qtx.ListGroceryItemsByBusiness(c.Request.Context(), businessID)
		} else {
			rows, err := qtx.ListGroceryItems(c.Request.Context())
			if err != nil {
				return err
			}
			c.JSON(http.StatusOK, rows)
			return nil
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
