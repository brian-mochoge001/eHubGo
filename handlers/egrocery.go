package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"ehubgo/cache"
	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GroceryHandler struct {
	Queries *db.Queries
	DB      *sql.DB
	Cache   cache.Store
}

func NewGroceryHandler(queries *db.Queries, dbConn *sql.DB, c cache.Store) *GroceryHandler {
	return &GroceryHandler{
		Queries: queries,
		DB:      dbConn,
		Cache:   c,
	}
}

func (h *GroceryHandler) ListGroceryItems(c *gin.Context) {
	businessID := c.Query("business_id")
	cacheKey := "grocery:items:" + businessID

	var items interface{}
	found, err := h.Cache.GetJSON(c.Request.Context(), cacheKey, &items)
	if err == nil && found {
		c.JSON(http.StatusOK, items)
		return
	}

	err = WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		var err error
		if businessID != "" {
			items, err = qtx.ListGroceryItemsByBusiness(c.Request.Context(), businessID)
		} else {
			items, err = qtx.ListGroceryItems(c.Request.Context())
		}
		return err
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = h.Cache.SetJSON(c.Request.Context(), cacheKey, items, 30*time.Minute)
	c.JSON(http.StatusOK, items)
}

func (h *GroceryHandler) CreateGroceryItem(c *gin.Context) {
	var req struct {
		BusinessID    string `json:"business_id" binding:"required"`
		Name          string `json:"name" binding:"required"`
		Description   string `json:"description"`
		Price         string `json:"price" binding:"required"`
		Currency      string `json:"currency"`
		ImageUrl      string `json:"image_url"`
		Unit          string `json:"unit"`
		StockQuantity int32  `json:"stock_quantity"`
		Category      string `json:"category"`
		IsAvailable   bool   `json:"is_available"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		item, err := qtx.CreateGroceryItem(c.Request.Context(), db.CreateGroceryItemParams{
			ID:            uuid.New().String(),
			BusinessID:    req.BusinessID,
			Name:          req.Name,
			Description:   sql.NullString{String: req.Description, Valid: req.Description != ""},
			Price:         req.Price,
			Currency:      req.Currency,
			ImageUrl:      sql.NullString{String: req.ImageUrl, Valid: req.ImageUrl != ""},
			Unit:          sql.NullString{String: req.Unit, Valid: req.Unit != ""},
			StockQuantity: req.StockQuantity,
			Category:      sql.NullString{String: req.Category, Valid: req.Category != ""},
			IsAvailable:   sql.NullBool{Bool: req.IsAvailable, Valid: true},
		})
		if err != nil {
			return err
		}

		// Invalidate cache
		_ = h.Cache.Delete(c.Request.Context(), "grocery:items:" + req.BusinessID)
		_ = h.Cache.Delete(c.Request.Context(), "grocery:items:") // Also invalidate general list

		c.JSON(http.StatusCreated, item)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *GroceryHandler) SearchGroceryStores(c *gin.Context) {
	var params struct {
		Latitude  float64 `form:"latitude" binding:"required"`
		Longitude float64 `form:"longitude" binding:"required"`
		Radius    float64 `form:"radius" binding:"required"`
	}

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		stores, err := qtx.SearchStoresByLocation(c.Request.Context(), db.SearchStoresByLocationParams{
			StMakepoint:   params.Longitude,
			StMakepoint_2: params.Latitude,
			Column3:       []string{"grocery"},
			StDwithin:     params.Radius,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusOK, []interface{}{})
				return nil
			}
			return err
		}

		var results []gin.H
		for _, store := range stores {
			storeLat := ParseCoordinate(store.Latitude)
			storeLng := ParseCoordinate(store.Longitude)

			results = append(results, gin.H{
				"business": store,
				"polyline": GetPointToPointRoute(c.Request.Context(), params.Latitude, params.Longitude, storeLat, storeLng),
			})
		}

		if len(results) == 0 {
			c.JSON(http.StatusOK, []interface{}{})
		} else {
			c.JSON(http.StatusOK, results)
		}
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

	if req.Distance <= 1.0 {
		c.JSON(http.StatusOK, gin.H{"estimated_price": 0.0, "currency": "Ksh", "type": "free"})
		return
	}

	price := CalculatePrice(50.0, 20.0, req.Distance, 1.0, 0.0)
	c.JSON(http.StatusOK, gin.H{
		"estimated_price": price,
		"currency":        "Ksh",
		"type":            "platform_delivery",
	})
}
