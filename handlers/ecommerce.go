package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EcommerceHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewEcommerceHandler(queries *db.Queries, dbConn *sql.DB) *EcommerceHandler {
	return &EcommerceHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

// ListFeaturedProducts returns best-selling and verified products
func (h *EcommerceHandler) ListFeaturedProducts(c *gin.Context) {
	limit := int32(10) // Default limit
	
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		products, err := qtx.GetFeaturedProducts(c.Request.Context(), limit)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, products)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetCart returns cart items for the current user
func (h *EcommerceHandler) GetCart(c *gin.Context) {
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

// AddToCart adds an item to the user's cart
func (h *EcommerceHandler) AddToCart(c *gin.Context) {
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

// UpdateCartItemQuantity updates the quantity of a cart item
func (h *EcommerceHandler) UpdateCartItemQuantity(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Quantity int32 `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		err := qtx.UpdateCartItemQuantity(c.Request.Context(), db.UpdateCartItemQuantityParams{
			ID:       id,
			Quantity: req.Quantity,
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, gin.H{"message": "quantity updated"})
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// RemoveCartItem removes an item from the cart
func (h *EcommerceHandler) RemoveCartItem(c *gin.Context) {
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

// Checkout creates an order from the user's cart
func (h *EcommerceHandler) Checkout(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	var req struct {
		ShippingAddressID string `json:"shipping_address_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)

		// 1. Get cart items
		cartItems, err := qtx.GetCartItemsByUserID(c.Request.Context(), userID)
		if err != nil {
			return err
		}

		if len(cartItems) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cart is empty"})
			return nil
		}

		// 2. ACID-Compliant Stock Reservation
		for _, item := range cartItems {
			err = qtx.LockAndDecrementStock(c.Request.Context(), db.LockAndDecrementStockParams{
				StockQuantity: item.Quantity,
				ID:            item.ItemID,
			})
			if err != nil {
				return fmt.Errorf("insufficient stock for item: %s", item.ItemID)
			}
		}

		// 3. Create Order
		orderID := uuid.New().String()
		totalAmount := 0.0 

		order, err := qtx.CreateOrder(c.Request.Context(), db.CreateOrderParams{
			ID:                orderID,
			UserID:            userID,
			TotalAmount:       fmt.Sprintf("%.2f", totalAmount),
			Currency:          "Ksh",
			Status:            "pending",
			ShippingAddressID: sql.NullString{String: req.ShippingAddressID, Valid: true},
		})
		if err != nil {
			return err
		}

		// 4. Create Order Items
		for _, item := range cartItems {
			unitPrice := "100.00" // Placeholder
			_, err = qtx.CreateOrderItem(c.Request.Context(), db.CreateOrderItemParams{
				ID:         uuid.New().String(),
				OrderID:    orderID,
				BusinessID: item.BusinessID,
				ItemID:     item.ItemID,
				ItemType:   item.ItemType,
				Quantity:   item.Quantity,
				UnitPrice:  unitPrice,
			})
			if err != nil {
				return err
			}
		}

		// 5. Clear Cart
		err = qtx.ClearCart(c.Request.Context(), userID)
		if err != nil {
			return err
		}

		c.JSON(http.StatusCreated, order)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetOrders returns orders for the current user
func (h *EcommerceHandler) GetOrders(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		orders, err := qtx.GetOrdersByUserID(c.Request.Context(), userID)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, orders)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// ListProducts returns all products
func (h *EcommerceHandler) ListProducts(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		products, err := qtx.GetProducts(c.Request.Context())
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, products)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetProductByID returns a single product with details
func (h *EcommerceHandler) GetProductByID(c *gin.Context) {
	id := c.Param("id")
	
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		product, err := qtx.GetProductByIDWithDetails(c.Request.Context(), id)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, product)
		return nil
	})

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

// ListCategories returns all categories
func (h *EcommerceHandler) ListCategories(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		categories, err := qtx.GetCategories(c.Request.Context())
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, categories)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// ListFlashSaleProducts returns flash sale products
func (h *EcommerceHandler) ListFlashSaleProducts(c *gin.Context) {
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		products, err := qtx.GetFlashSaleProducts(c.Request.Context())
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, products)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
