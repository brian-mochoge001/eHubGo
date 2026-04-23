package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CartHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewCartHandler(queries *db.Queries, dbConn *sql.DB) *CartHandler {
	return &CartHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		items, err := qtx.GetCartItemsByUserID(c.Request.Context(), userID)
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

func (h *CartHandler) AddToCart(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	var req struct {
		BusinessID string `json:"business_id" binding:"required"`
		ItemID     string `json:"item_id" binding:"required"`
		ItemType   string `json:"item_type" binding:"required"`
		Quantity   int32  `json:"quantity" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		params := db.AddItemToCartParams{
			ID:         uuid.New().String(),
			UserID:     userID,
			BusinessID: req.BusinessID,
			ItemID:     req.ItemID,
			ItemType:   req.ItemType,
			Quantity:   req.Quantity,
		}

		item, err := qtx.AddItemToCart(c.Request.Context(), params)
		if err != nil {
			return err
		}
		c.JSON(http.StatusCreated, item)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *CartHandler) RemoveCartItem(c *gin.Context) {
	id := c.Param("id")
	
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		err := qtx.RemoveCartItem(c.Request.Context(), id)
		if err != nil {
			return err
		}
		c.JSON(http.StatusNoContent, nil)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
