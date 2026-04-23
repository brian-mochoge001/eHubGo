package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type SpecializedHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewSpecializedHandler(queries *db.Queries, dbConn *sql.DB) *SpecializedHandler {
	return &SpecializedHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

// eGrocery
func (h *SpecializedHandler) ListGroceryItems(c *gin.Context) {
	businessID := c.Query("business_id")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		var items []db.GroceryItem
		var err error

		if businessID != "" {
			items, err = qtx.ListGroceryItemsByBusiness(c.Request.Context(), businessID)
		} else {
			items, err = qtx.ListGroceryItems(c.Request.Context())
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

// eLiquor
func (h *SpecializedHandler) ListLiquorItems(c *gin.Context) {
	businessID := c.Query("business_id")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		var items []db.LiquorItem
		var err error

		if businessID != "" {
			items, err = qtx.ListLiquorItemsByBusiness(c.Request.Context(), businessID)
		} else {
			items, err = qtx.ListLiquorItems(c.Request.Context())
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

// eHealth / Pharmacy
func (h *SpecializedHandler) ListPharmacyItems(c *gin.Context) {
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

// eFood (Specific list all food items with restaurant name)
func (h *SpecializedHandler) ListAllFoodItems(c *gin.Context) {
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
