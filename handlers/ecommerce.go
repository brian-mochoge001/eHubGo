package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ehubgo/cache"
	"ehubgo/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type EcommerceHandler struct {
	Queries *db.Queries
	DB      *sql.DB
	Cache   cache.Store
	OC      *OrderCoordinator
}

func NewEcommerceHandler(queries *db.Queries, dbConn *sql.DB, c cache.Store, oc *OrderCoordinator) *EcommerceHandler {
	return &EcommerceHandler{
		Queries: queries,
		DB:      dbConn,
		Cache:   c,
		OC:      oc,
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
	ModelID            *string    `json:"model_id"`
	ImageUrls          []string   `json:"image_urls"`
	Rating             *string    `json:"rating"`
	ReviewCount        *int32     `json:"review_count"`
	IsFlashSale        *bool      `json:"is_flash_sale"`
	DiscountPercentage *string    `json:"discount_percentage"`
	CreatedAt          *time.Time `json:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at"`
	CategoryName       *string    `json:"category_name"`
	BrandName          *string    `json:"brand_name"`
	ModelName          *string    `json:"model_name"`
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
		ModelID:            NullStringToString(p.ModelID),
		ImageUrls:          p.ImageUrls,
		Rating:             NullStringToString(p.Rating),
		ReviewCount:        NullInt32ToInt32(p.ReviewCount),
		IsFlashSale:        NullBoolToBool(p.IsFlashSale),
		DiscountPercentage: NullStringToString(p.DiscountPercentage),
		CreatedAt:          NullTimeToTime(p.CreatedAt),
		UpdatedAt:          NullTimeToTime(p.UpdatedAt),
	}
}

func GetProductsRowToDTO(p db.GetProductsRow) ProductDTO {
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
		ModelID:            NullStringToString(p.ModelID),
		ImageUrls:          p.ImageUrls,
		Rating:             NullStringToString(p.Rating),
		ReviewCount:        NullInt32ToInt32(p.ReviewCount),
		IsFlashSale:        NullBoolToBool(p.IsFlashSale),
		DiscountPercentage: NullStringToString(p.DiscountPercentage),
		CreatedAt:          NullTimeToTime(p.CreatedAt),
		UpdatedAt:          NullTimeToTime(p.UpdatedAt),
		CategoryName:       NullStringToString(p.CategoryName),
		BrandName:          NullStringToString(p.BrandName),
		ModelName:          NullStringToString(p.ModelName),
	}
}

func GetProductsByBusinessRowToDTO(p db.GetProductsByBusinessRow) ProductDTO {
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
		ModelID:            NullStringToString(p.ModelID),
		ImageUrls:          p.ImageUrls,
		Rating:             NullStringToString(p.Rating),
		ReviewCount:        NullInt32ToInt32(p.ReviewCount),
		IsFlashSale:        NullBoolToBool(p.IsFlashSale),
		DiscountPercentage: NullStringToString(p.DiscountPercentage),
		CreatedAt:          NullTimeToTime(p.CreatedAt),
		UpdatedAt:          NullTimeToTime(p.UpdatedAt),
		CategoryName:       NullStringToString(p.CategoryName),
		BrandName:          NullStringToString(p.BrandName),
		ModelName:          NullStringToString(p.ModelName),
	}
}

func GetProductByIDRowToDTO(p db.GetProductByIDWithDetailsRow) ProductDTO {
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
		ModelID:            NullStringToString(p.ModelID),
		ImageUrls:          p.ImageUrls,
		Rating:             NullStringToString(p.Rating),
		ReviewCount:        NullInt32ToInt32(p.ReviewCount),
		IsFlashSale:        NullBoolToBool(p.IsFlashSale),
		DiscountPercentage: NullStringToString(p.DiscountPercentage),
		CreatedAt:          NullTimeToTime(p.CreatedAt),
		UpdatedAt:          NullTimeToTime(p.UpdatedAt),
		CategoryName:       NullStringToString(p.CategoryName),
		BrandName:          NullStringToString(p.BrandName),
		ModelName:          NullStringToString(p.ModelName),
	}
}

// @Summary List featured products
// @Description Returns best-selling and verified products with pagination and caching
// @Tags products
// @Accept json
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} ProductDTO
// @Failure 500 {object} map[string]string
// @Router /api/v1/featured-products [get]
func (h *EcommerceHandler) ListFeaturedProducts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	cacheKey := fmt.Sprintf("ecommerce:featured:%s:%s", limitStr, offsetStr)
	var dtoList []ProductDTO

	// 1. Try to get from cache
	found, err := h.Cache.GetJSON(c.Request.Context(), cacheKey, &dtoList)
	if err == nil && found {
		c.JSON(http.StatusOK, dtoList)
		return
	}

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// 2. Cache miss: Get from database
	err = WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		products, err := qtx.GetFeaturedProducts(c.Request.Context(), db.GetFeaturedProductsParams{
			Offset: int32(offset),
			Limit:  int32(limit),
		})
		if err != nil {
			return err
		}

		dtoList = make([]ProductDTO, 0)
		for _, p := range products {
			dtoList = append(dtoList, ToProductDTO(p))
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Store in cache for 15 minutes (featured products change more often than categories)
	_ = h.Cache.SetJSON(c.Request.Context(), cacheKey, dtoList, 15*time.Minute)

	c.JSON(http.StatusOK, dtoList)
}

// @Summary Search products
// @Description Uses PostgreSQL full-text search to find products by name or description
// @Tags products
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Limit"
// @Success 200 {array} ProductDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/products/search [get]
func (h *EcommerceHandler) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query 'q' is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)

	// Since sqlc queries might not have the specific search endpoint generated depending on the schema version,
	// we use a raw SQL query that leverages the to_tsvector index created in schema.sql.
	sqlQuery := `
		SELECT p.id, p.business_id, p.name, p.description, p.price, p.currency, 
			   p.stock_quantity, p.category_id, p.brand_id, p.model_id, p.image_urls, 
			   p.rating, p.review_count, p.is_flash_sale, p.discount_percentage, p.created_at, p.updated_at
		FROM products p
		WHERE to_tsvector('english', p.name || ' ' || COALESCE(p.description, '')) @@ plainto_tsquery('english', $1)
		ORDER BY ts_rank(to_tsvector('english', p.name || ' ' || COALESCE(p.description, '')), plainto_tsquery('english', $1)) DESC
		LIMIT $2
	`

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(c.Request.Context(), sqlQuery, query, limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		dtoList := make([]ProductDTO, 0)
		for rows.Next() {
			var p db.Product
			err := rows.Scan(
				&p.ID, &p.BusinessID, &p.Name, &p.Description, &p.Price, &p.Currency,
				&p.StockQuantity, &p.CategoryID, &p.BrandID, &p.ModelID, pq.Array(&p.ImageUrls),
				&p.Rating, &p.ReviewCount, &p.IsFlashSale, &p.DiscountPercentage, &p.CreatedAt, &p.UpdatedAt,
			)
			if err != nil {
				return err
			}
			dtoList = append(dtoList, ToProductDTO(p))
		}
		c.JSON(http.StatusOK, dtoList)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// ListProducts returns all products with pagination, optionally filtered by business_id
func (h *EcommerceHandler) ListProducts(c *gin.Context) {
	businessID := c.Query("business_id")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	cacheKey := fmt.Sprintf("ecommerce:products:%s:%s:%s", businessID, limitStr, offsetStr)

	var dtoList []ProductDTO
	found, err := h.Cache.GetJSON(c.Request.Context(), cacheKey, &dtoList)
	if err == nil && found {
		c.JSON(http.StatusOK, dtoList)
		return
	}

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	err = WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		dtoList = make([]ProductDTO, 0)

		if businessID != "" {
			products, err := qtx.GetProductsByBusiness(c.Request.Context(), db.GetProductsByBusinessParams{
				BusinessID: businessID,
				Limit:      int32(limit),
				Offset:     int32(offset),
			})
			if err != nil {
				return err
			}
			for _, p := range products {
				dtoList = append(dtoList, GetProductsByBusinessRowToDTO(p))
			}
		} else {
			products, err := qtx.GetProducts(c.Request.Context(), db.GetProductsParams{
				Limit:  int32(limit),
				Offset: int32(offset),
			})
			if err != nil {
				return err
			}
			for _, p := range products {
				dtoList = append(dtoList, GetProductsRowToDTO(p))
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = h.Cache.SetJSON(c.Request.Context(), cacheKey, dtoList, 10*time.Minute)
	c.JSON(http.StatusOK, dtoList)
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

		dto := GetProductByIDRowToDTO(p)
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

// CreateProduct creates a new product
func (h *EcommerceHandler) CreateProduct(c *gin.Context) {
	var req struct {
		BusinessID         string   `json:"business_id" binding:"required"`
		Name               string   `json:"name" binding:"required"`
		Description        string   `json:"description"`
		Price              string   `json:"price" binding:"required"`
		Currency           string   `json:"currency"`
		StockQuantity      int32    `json:"stock_quantity"`
		CategoryID         string   `json:"category_id"`
		BrandID            string   `json:"brand_id"`
		ModelID            string   `json:"model_id"`
		ImageUrls          []string `json:"image_urls"`
		IsFlashSale        bool     `json:"is_flash_sale"`
		DiscountPercentage string   `json:"discount_percentage"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Currency == "" {
		req.Currency = "Ksh"
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		p, err := qtx.CreateProduct(c.Request.Context(), db.CreateProductParams{
			ID:                 uuid.New().String(),
			BusinessID:         req.BusinessID,
			Name:               req.Name,
			Description:        sql.NullString{String: req.Description, Valid: req.Description != ""},
			Price:              req.Price,
			Currency:           req.Currency,
			StockQuantity:      req.StockQuantity,
			CategoryID:         sql.NullString{String: req.CategoryID, Valid: req.CategoryID != ""},
			BrandID:            sql.NullString{String: req.BrandID, Valid: req.BrandID != ""},
			ModelID:            sql.NullString{String: req.ModelID, Valid: req.ModelID != ""},
			ImageUrls:          req.ImageUrls,
			IsFlashSale:        sql.NullBool{Bool: req.IsFlashSale, Valid: true},
			DiscountPercentage: sql.NullString{String: req.DiscountPercentage, Valid: req.DiscountPercentage != ""},
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusCreated, ToProductDTO(p))
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// UpdateProduct updates an existing product
func (h *EcommerceHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name               string   `json:"name" binding:"required"`
		Description        string   `json:"description"`
		Price              string   `json:"price" binding:"required"`
		StockQuantity      int32    `json:"stock_quantity"`
		CategoryID         string   `json:"category_id"`
		BrandID            string   `json:"brand_id"`
		ModelID            string   `json:"model_id"`
		ImageUrls          []string `json:"image_urls"`
		IsFlashSale        bool     `json:"is_flash_sale"`
		DiscountPercentage string   `json:"discount_percentage"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		p, err := qtx.UpdateProduct(c.Request.Context(), db.UpdateProductParams{
			ID:                 id,
			Name:               req.Name,
			Description:        sql.NullString{String: req.Description, Valid: req.Description != ""},
			Price:              req.Price,
			StockQuantity:      req.StockQuantity,
			CategoryID:         sql.NullString{String: req.CategoryID, Valid: req.CategoryID != ""},
			BrandID:            sql.NullString{String: req.BrandID, Valid: req.BrandID != ""},
			ModelID:            sql.NullString{String: req.ModelID, Valid: req.ModelID != ""},
			ImageUrls:          req.ImageUrls,
			IsFlashSale:        sql.NullBool{Bool: req.IsFlashSale, Valid: true},
			DiscountPercentage: sql.NullString{String: req.DiscountPercentage, Valid: req.DiscountPercentage != ""},
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, ToProductDTO(p))
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// DeleteProduct removes a product from the database
func (h *EcommerceHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		err := qtx.DeleteProduct(c.Request.Context(), id)
		if err != nil {
			return err
		}

		// Invalidate cache
		_ = h.Cache.Delete(c.Request.Context(), "ecommerce:product:" + id)
		_ = h.Cache.Delete(c.Request.Context(), "ecommerce:featured:*")
		_ = h.Cache.Delete(c.Request.Context(), "ecommerce:products:*")

		c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *EcommerceHandler) DeleteProductModel(c *gin.Context) {
	id := c.Param("id")
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		return qtx.DeleteProductModel(c.Request.Context(), id)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product model deleted"})
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

// Checkout creates an order from the user's cart using OrderCoordinator
func (h *EcommerceHandler) Checkout(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	var req struct {
		ShippingAddressID string `json:"shipping_address_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Get cart items
	var cartItems []db.GetCartItemsByUserIDRow
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		items, err := qtx.GetCartItemsByUserID(c.Request.Context(), userID)
		if err != nil {
			return err
		}
		cartItems = items
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. Orchestrate Checkout
	parentOrderID, err := h.OC.OrchestrateCheckout(c.Request.Context(), userID, req.ShippingAddressID, cartItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"parent_order_id": parentOrderID, "message": "Checkout successful"})
}

// GetMiniserviceAnalytics returns platform-wide analytics for a specific miniservice type
func (h *EcommerceHandler) GetMiniserviceAnalytics(c *gin.Context) {
	serviceType := c.Query("type")
	// For now, returning mock data as the analytics schema is still under development
	c.JSON(http.StatusOK, gin.H{
		"type":            serviceType,
		"total_revenue":   1250000.0,
		"total_orders":    4500,
		"active_users":    1284,
		"active_drivers":  24,
		"active_vendors":  45,
		"completion_rate": 94.5,
	})
}

// GetStoreAnalytics returns analytics for a specific business
func (h *EcommerceHandler) GetStoreAnalytics(c *gin.Context) {
	businessID := c.Param("id")
	// For now, returning mock data
	c.JSON(http.StatusOK, gin.H{
		"business_id":     businessID,
		"total_revenue":   45000.0,
		"total_orders":    124,
		"active_users":    85, // Represents visitors for vendor
		"active_drivers":  5,  // Represents assigned drivers
		"completion_rate": 98.0,
	})
}

// Category Management

func (h *EcommerceHandler) ListCategories(c *gin.Context) {
	categories, err := h.Queries.GetCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (h *EcommerceHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		ImageUrl    string `json:"image_url"`
		ParentID    string `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.Queries.CreateCategory(c.Request.Context(), db.CreateCategoryParams{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: NullStringToSQL(req.Description),
		ImageUrl:    NullStringToSQL(req.ImageUrl),
		ParentID:    NullStringToSQL(req.ParentID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, category)
}

func (h *EcommerceHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ImageUrl    string `json:"image_url"`
		ParentID    string `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.Queries.UpdateCategory(c.Request.Context(), db.UpdateCategoryParams{
		ID:          id,
		Name:        req.Name,
		Description: NullStringToSQL(req.Description),
		ImageUrl:    NullStringToSQL(req.ImageUrl),
		ParentID:    NullStringToSQL(req.ParentID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *EcommerceHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	err := h.Queries.DeleteCategory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

// Brand Management

func (h *EcommerceHandler) ListBrands(c *gin.Context) {
	brands, err := h.Queries.GetBrands(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch brands"})
		return
	}
	c.JSON(http.StatusOK, brands)
}

func (h *EcommerceHandler) CreateBrand(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		LogoUrl    string `json:"logo_url"`
		CategoryID string `json:"category_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	brand, err := h.Queries.CreateBrand(c.Request.Context(), db.CreateBrandParams{
		ID:         uuid.New().String(),
		Name:       req.Name,
		LogoUrl:    NullStringToSQL(req.LogoUrl),
		CategoryID: NullStringToSQL(req.CategoryID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create brand"})
		return
	}

	c.JSON(http.StatusCreated, brand)
}

func (h *EcommerceHandler) UpdateBrand(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name       string `json:"name"`
		LogoUrl    string `json:"logo_url"`
		CategoryID string `json:"category_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	brand, err := h.Queries.UpdateBrand(c.Request.Context(), db.UpdateBrandParams{
		ID:         id,
		Name:       req.Name,
		LogoUrl:    NullStringToSQL(req.LogoUrl),
		CategoryID: NullStringToSQL(req.CategoryID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update brand"})
		return
	}

	c.JSON(http.StatusOK, brand)
}

func (h *EcommerceHandler) DeleteBrand(c *gin.Context) {
	id := c.Param("id")

	err := h.Queries.DeleteBrand(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete brand"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Brand deleted successfully"})
}

// Product Model Management

func (h *EcommerceHandler) ListProductModels(c *gin.Context) {
	models, err := h.Queries.ListProductModels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product models"})
		return
	}
	c.JSON(http.StatusOK, models)
}

func (h *EcommerceHandler) CreateProductModel(c *gin.Context) {
	var req struct {
		Name    string `json:"name" binding:"required"`
		BrandID string `json:"brand_id" binding:"required"`
		ImageUrl string `json:"image_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model, err := h.Queries.CreateProductModel(c.Request.Context(), db.CreateProductModelParams{
		ID:       uuid.New().String(),
		BrandID:  req.BrandID,
		Name:     req.Name,
		ImageUrl: NullStringToSQL(req.ImageUrl),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product model"})
		return
	}

	c.JSON(http.StatusCreated, model)
}

func (h *EcommerceHandler) UpdateProductModel(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name     string `json:"name"`
		BrandID  string `json:"brand_id"`
		ImageUrl string `json:"image_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model, err := h.Queries.UpdateProductModel(c.Request.Context(), db.UpdateProductModelParams{
		ID:       id,
		BrandID:  req.BrandID,
		Name:     req.Name,
		ImageUrl: NullStringToSQL(req.ImageUrl),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product model"})
		return
	}

	c.JSON(http.StatusOK, model)
}
