package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

type ProductDTO struct {
	ID                 string     `json:"id"`
	BusinessID         string     `json:"business_id"`
	Name               string     `json:"name"`
	Description        *string    `json:"description"`
	Price              string     `json:"price"`
	Currency           string     `json:"currency"`
	StockQuantity      int32      `json:"stock_quantity"`
	CategoryID         *string    `json:"category_id"`
	BrandID            *string    `json:"brand_id"`
	Rating             *string    `json:"rating"`
	ReviewCount        *int32     `json:"review_count"`
	IsFlashSale        *bool      `json:"is_flash_sale"`
	DiscountPercentage *string    `json:"discount_percentage"`
	CreatedAt          *time.Time `json:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at"`
}

func ToProductDTO(p db.Product) ProductDTO {
	return ProductDTO{
		ID:                 p.ID,
		BusinessID:         p.BusinessID,
		Name:               p.Name,
		Description:        NullStringToString(p.Description),
		Price:              p.Price,
		Currency:           p.Currency,
		StockQuantity:      p.StockQuantity,
		CategoryID:         NullStringToString(p.CategoryID),
		BrandID:            NullStringToString(p.BrandID),
		Rating:             NullStringToString(p.Rating),
		ReviewCount:        NullInt32ToInt32(p.ReviewCount),
		IsFlashSale:        NullBoolToBool(p.IsFlashSale),
		DiscountPercentage: NullStringToString(p.DiscountPercentage),
		CreatedAt:          NullTimeToTime(p.CreatedAt),
		UpdatedAt:          NullTimeToTime(p.UpdatedAt),
	}
}

func ToProductDTOFromFeatured(p db.GetFeaturedProductsRow) ProductDTO {
	return ProductDTO{
		ID:                 p.ID,
		BusinessID:         p.BusinessID,
		Name:               p.Name,
		Description:        NullStringToString(p.Description),
		Price:              p.Price,
		Currency:           p.Currency,
		StockQuantity:      p.StockQuantity,
		CategoryID:         NullStringToString(p.CategoryID),
		BrandID:            NullStringToString(p.BrandID),
		Rating:             NullStringToString(p.Rating),
		ReviewCount:        NullInt32ToInt32(p.ReviewCount),
		IsFlashSale:        NullBoolToBool(p.IsFlashSale),
		DiscountPercentage: NullStringToString(p.DiscountPercentage),
		CreatedAt:          NullTimeToTime(p.CreatedAt),
		UpdatedAt:          NullTimeToTime(p.UpdatedAt),
	}
}

func ToProductDTOFromDetails(p db.GetProductByIDWithDetailsRow) ProductDTO {
	return ProductDTO{
		ID:                 p.ID,
		BusinessID:         p.BusinessID,
		Name:               p.Name,
		Description:        NullStringToString(p.Description),
		Price:              p.Price,
		Currency:           p.Currency,
		StockQuantity:      p.StockQuantity,
		CategoryID:         NullStringToString(p.CategoryID),
		BrandID:            NullStringToString(p.BrandID),
		Rating:             NullStringToString(p.Rating),
		ReviewCount:        NullInt32ToInt32(p.ReviewCount),
		IsFlashSale:        NullBoolToBool(p.IsFlashSale),
		DiscountPercentage: NullStringToString(p.DiscountPercentage),
		CreatedAt:          NullTimeToTime(p.CreatedAt),
		UpdatedAt:          NullTimeToTime(p.UpdatedAt),
	}
}

func ToProductDTOFromGetProducts(p db.GetProductsRow) ProductDTO {
	return ProductDTO{
		ID:                 p.ID,
		BusinessID:         p.BusinessID,
		Name:               p.Name,
		Description:        NullStringToString(p.Description),
		Price:              p.Price,
		Currency:           p.Currency,
		StockQuantity:      p.StockQuantity,
		CategoryID:         NullStringToString(p.CategoryID),
		BrandID:            NullStringToString(p.BrandID),
		Rating:             NullStringToString(p.Rating),
		ReviewCount:        NullInt32ToInt32(p.ReviewCount),
		IsFlashSale:        NullBoolToBool(p.IsFlashSale),
		DiscountPercentage: NullStringToString(p.DiscountPercentage),
		CreatedAt:          NullTimeToTime(p.CreatedAt),
		UpdatedAt:          NullTimeToTime(p.UpdatedAt),
	}
}

// ListFeaturedProducts returns best-selling and verified products with pagination
func (h *EcommerceHandler) ListFeaturedProducts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		products, err := qtx.GetFeaturedProducts(c.Request.Context(), db.GetFeaturedProductsParams{
			OffsetCount: int32(offset),
			LimitCount:  int32(limit),
		})
		if err != nil {
			return err
		}

		dtoList := make([]ProductDTO, 0)
		for _, p := range products {
			dtoList = append(dtoList, ToProductDTOFromFeatured(p))
		}
		c.JSON(http.StatusOK, dtoList)
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

type CategoryDTO struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	ImageUrl    *string    `json:"image_url"`
	ParentID    *string    `json:"parent_id"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

func ToCategoryDTO(cat db.Category) CategoryDTO {
	return CategoryDTO{
		ID:          cat.ID,
		Name:        cat.Name,
		Description: NullStringToString(cat.Description),
		ImageUrl:    NullStringToString(cat.ImageUrl),
		ParentID:    NullStringToString(cat.ParentID),
		CreatedAt:   NullTimeToTime(cat.CreatedAt),
		UpdatedAt:   NullTimeToTime(cat.UpdatedAt),
	}
}

// ListProducts returns all products with pagination
func (h *EcommerceHandler) ListProducts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		products, err := qtx.GetProducts(c.Request.Context(), db.GetProductsParams{
			OffsetCount: int32(offset),
			LimitCount:  int32(limit),
		})
		if err != nil {
			return err
		}

		dtoList := make([]ProductDTO, 0)
		for _, p := range products {
			dtoList = append(dtoList, ToProductDTOFromGetProducts(p))
		}
		c.JSON(http.StatusOK, dtoList)
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
		p, err := qtx.GetProductByIDWithDetails(c.Request.Context(), id)
		if err != nil {
			return err
		}

		dto := ToProductDTOFromDetails(p)
		c.JSON(http.StatusOK, dto)
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

		dtoList := make([]CategoryDTO, 0)
		for _, cat := range categories {
			dtoList = append(dtoList, ToCategoryDTO(cat))
		}
		c.JSON(http.StatusOK, dtoList)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// ListFlashSaleProducts returns flash sale products with pagination
func (h *EcommerceHandler) ListFlashSaleProducts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		products, err := qtx.GetFlashSaleProducts(c.Request.Context(), db.GetFlashSaleProductsParams{
			OffsetCount: int32(offset),
			LimitCount:  int32(limit),
		})
		if err != nil {
			return err
		}

		dtoList := make([]ProductDTO, 0)
		for _, p := range products {
			dtoList = append(dtoList, ToProductDTO(p))
		}
		c.JSON(http.StatusOK, dtoList)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
