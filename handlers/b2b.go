package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type B2BHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewB2BHandler(queries *db.Queries, dbConn *sql.DB) *B2BHandler {
	return &B2BHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

// GetB2BDashboard returns stats for the B2B dashboard
func (h *B2BHandler) GetB2BDashboard(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		stats, err := qtx.GetB2BDashboardStats(c.Request.Context(), userID)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, stats)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// ListB2BServices lists businesses of type B2B
func (h *B2BHandler) ListB2BServices(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		services, err := qtx.ListServicesByType(c.Request.Context(), "b2b")
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

// CreateRFQ allows a business to request a quote from a supplier
func (h *B2BHandler) CreateRFQ(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var req struct {
		BusinessID  string `json:"business_id" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Items       []struct {
			ProductID   *string `json:"product_id"`
			ItemName    string  `json:"item_name" binding:"required"`
			Quantity    int32   `json:"quantity" binding:"required"`
			Unit        string  `json:"unit"`
			TargetPrice float64 `json:"target_price"`
		} `json:"items" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)

		// 1. Create RFQ
		rfq, err := qtx.CreateRFQ(c.Request.Context(), db.CreateRFQParams{
			ID:          uuid.New().String(),
			BuyerID:     userID,
			BusinessID:  req.BusinessID,
			Title:       req.Title,
			Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
			Status:      "pending",
		})
		if err != nil {
			return err
		}

		// 2. Add Items
		for _, item := range req.Items {
			_, err = qtx.AddRFQItem(c.Request.Context(), db.AddRFQItemParams{
				ID:          uuid.New().String(),
				RfqID:       rfq.ID,
				ProductID:   sql.NullString{String: *item.ProductID, Valid: item.ProductID != nil},
				ItemName:    item.ItemName,
				Quantity:    item.Quantity,
				Unit:        sql.NullString{String: item.Unit, Valid: item.Unit != ""},
				TargetPrice: sql.NullString{String: string(rune(item.TargetPrice)), Valid: item.TargetPrice > 0}, // Placeholder for numeric conversion
			})
			if err != nil {
				return err
			}
		}

		c.JSON(http.StatusCreated, rfq)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// ListMyRFQs returns RFQs created by the user
func (h *B2BHandler) ListMyRFQs(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		rfqs, err := qtx.ListRFQsForBuyer(c.Request.Context(), userID)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, rfqs)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// SubmitQuote allows a supplier to respond to an RFQ
func (h *B2BHandler) SubmitQuote(c *gin.Context) {
	var req struct {
		RFQID       string    `json:"rfq_id" binding:"required"`
		BusinessID  string    `json:"business_id" binding:"required"`
		TotalAmount float64   `json:"total_amount" binding:"required"`
		Currency    string    `json:"currency" default:"Ksh"`
		ValidUntil  time.Time `json:"valid_until"`
		Notes       string    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)

		quote, err := qtx.CreateB2BQuote(c.Request.Context(), db.CreateB2BQuoteParams{
			ID:          uuid.New().String(),
			RfqID:       req.RFQID,
			BusinessID:  req.BusinessID,
			TotalAmount: string(rune(req.TotalAmount)), // Placeholder
			Currency:    req.Currency,
			ValidUntil:  sql.NullTime{Time: req.ValidUntil, Valid: !req.ValidUntil.IsZero()},
			Status:      "offered",
			Notes:       sql.NullString{String: req.Notes, Valid: req.Notes != ""},
		})
		if err != nil {
			return err
		}

		// Update RFQ status
		_, _ = qtx.UpdateRFQStatus(c.Request.Context(), db.UpdateRFQStatusParams{
			ID:     req.RFQID,
			Status: "responded",
		})

		c.JSON(http.StatusCreated, quote)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// ListWholesaleItems returns all available wholesale products
func (h *B2BHandler) ListWholesaleItems(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		items, err := qtx.ListWholesaleItems(c.Request.Context())
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

// GetWholesaleItem returns details for a specific wholesale item including its business info and locations
func (h *B2BHandler) GetWholesaleItem(c *gin.Context) {
	id := c.Param("id")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		
		// 1. Get item details (includes business info from join)
		item, err := qtx.GetWholesaleItemByID(c.Request.Context(), id)
		if err != nil {
			return err
		}

		// 2. Get business locations
		locations, err := qtx.ListBusinessLocations(c.Request.Context(), item.BusinessID)
		if err != nil {
			return err
		}

		// 3. Get reviews
		reviews, _ := qtx.GetReviewsByTarget(c.Request.Context(), db.GetReviewsByTargetParams{
			TargetID:   item.ID,
			TargetType: "wholesale_item",
		})

		c.JSON(http.StatusOK, gin.H{
			"item":      item,
			"locations": locations,
			"reviews":   reviews,
		})
		return nil
	})

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "wholesale item not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

