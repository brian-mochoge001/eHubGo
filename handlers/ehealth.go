package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
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
			// Updated to use the new join query that includes business name
			rows, err := qtx.ListPharmacyItems(c.Request.Context())
			if err != nil {
				return err
			}
			// Map to pharmacy items for simple list
			for _, r := range rows {
				items = append(items, db.PharmacyItem{
					ID: r.ID,
					BusinessID: r.BusinessID,
					Name: r.Name,
					Description: r.Description,
					Price: r.Price,
					Currency: r.Currency,
					ImageUrl: r.ImageUrl,
					RequiresPrescription: r.RequiresPrescription,
					StockQuantity: r.StockQuantity,
					Category: r.Category,
					IsAvailable: r.IsAvailable,
					CreatedAt: r.CreatedAt,
					UpdatedAt: r.UpdatedAt,
				})
			}
			c.JSON(http.StatusOK, rows) // Sending full row with business name
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
