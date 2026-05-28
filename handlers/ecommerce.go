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
}

func NewEcommerceHandler(queries *db.Queries, dbConn *sql.DB, c cache.Store) *EcommerceHandler {
	return &EcommerceHandler{
		Queries: queries,
		DB:      dbConn,
		Cache:   c,
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
		c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// --- Categories ---

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

// ListCategories returns all categories, with Redis caching
func (h *EcommerceHandler) ListCategories(c *gin.Context) {
	cacheKey := "ecommerce:categories"
	var dtoList []CategoryDTO

	// 1. Try to get from cache
	found, err := h.Cache.GetJSON(c.Request.Context(), cacheKey, &dtoList)
	if err == nil && found {
		c.JSON(http.StatusOK, dtoList)
		return
	}

	// 2. Cache miss: Get from database
	err = WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		categories, err := qtx.GetCategories(c.Request.Context())
		if err != nil {
			return err
		}

		dtoList = make([]CategoryDTO, 0)
		for _, cat := range categories {
			dtoList = append(dtoList, ToCategoryDTO(cat))
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Store in cache for 1 hour
	_ = h.Cache.SetJSON(c.Request.Context(), cacheKey, dtoList, 1*time.Hour)

	c.JSON(http.StatusOK, dtoList)
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
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		cat, err := qtx.CreateCategory(c.Request.Context(), db.CreateCategoryParams{
			ID:          uuid.New().String(),
			Name:        req.Name,
			Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
			ImageUrl:    sql.NullString{String: req.ImageUrl, Valid: req.ImageUrl != ""},
			ParentID:    sql.NullString{String: req.ParentID, Valid: req.ParentID != ""},
		})
		if err != nil {
			return err
		}

		// Invalidate cache
		_ = h.Cache.Delete(c.Request.Context(), "ecommerce:categories")

		c.JSON(http.StatusCreated, ToCategoryDTO(cat))
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *EcommerceHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
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
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		cat, err := qtx.UpdateCategory(c.Request.Context(), db.UpdateCategoryParams{
			ID:          id,
			Name:        req.Name,
			Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
			ImageUrl:    sql.NullString{String: req.ImageUrl, Valid: req.ImageUrl != ""},
			ParentID:    sql.NullString{String: req.ParentID, Valid: req.ParentID != ""},
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, ToCategoryDTO(cat))
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *EcommerceHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		return qtx.DeleteCategory(c.Request.Context(), id)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "category deleted"})
}

// --- Brands ---

type BrandDTO struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	LogoUrl      *string    `json:"logo_url"`
	CategoryID   *string    `json:"category_id"`
	CategoryName *string    `json:"category_name"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

func ToBrandDTO(b db.Brand) BrandDTO {
	return BrandDTO{
		ID:         b.ID,
		Name:       b.Name,
		LogoUrl:    NullStringToString(b.LogoUrl),
		CategoryID: NullStringToString(b.CategoryID),
		CreatedAt:  NullTimeToTime(b.CreatedAt),
		UpdatedAt:  NullTimeToTime(b.UpdatedAt),
	}
}

func GetBrandsRowToDTO(b db.GetBrandsRow) BrandDTO {
	return BrandDTO{
		ID:           b.ID,
		Name:         b.Name,
		LogoUrl:      NullStringToString(b.LogoUrl),
		CategoryID:   NullStringToString(b.CategoryID),
		CategoryName: NullStringToString(b.CategoryName),
		CreatedAt:    NullTimeToTime(b.CreatedAt),
		UpdatedAt:    NullTimeToTime(b.UpdatedAt),
	}
}

func (h *EcommerceHandler) ListBrands(c *gin.Context) {
	catID := c.Query("category_id")
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		if catID != "" {
			brands, err := qtx.ListBrandsByCategory(c.Request.Context(), sql.NullString{String: catID, Valid: true})
			if err != nil {
				return err
			}
			dtoList := make([]BrandDTO, 0)
			for _, b := range brands {
				dtoList = append(dtoList, ToBrandDTO(b))
			}
			c.JSON(http.StatusOK, dtoList)
		} else {
			brands, err := qtx.GetBrands(c.Request.Context())
			if err != nil {
				return err
			}
			dtoList := make([]BrandDTO, 0)
			for _, b := range brands {
				dtoList = append(dtoList, GetBrandsRowToDTO(b))
			}
			c.JSON(http.StatusOK, dtoList)
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
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
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		brand, err := qtx.CreateBrand(c.Request.Context(), db.CreateBrandParams{
			ID:         uuid.New().String(),
			Name:       req.Name,
			LogoUrl:    sql.NullString{String: req.LogoUrl, Valid: req.LogoUrl != ""},
			CategoryID: sql.NullString{String: req.CategoryID, Valid: req.CategoryID != ""},
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusCreated, ToBrandDTO(brand))
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *EcommerceHandler) UpdateBrand(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name       string `json:"name" binding:"required"`
		LogoUrl    string `json:"logo_url"`
		CategoryID string `json:"category_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		brand, err := qtx.UpdateBrand(c.Request.Context(), db.UpdateBrandParams{
			ID:         id,
			Name:       req.Name,
			LogoUrl:    sql.NullString{String: req.LogoUrl, Valid: req.LogoUrl != ""},
			CategoryID: sql.NullString{String: req.CategoryID, Valid: req.CategoryID != ""},
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, ToBrandDTO(brand))
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *EcommerceHandler) DeleteBrand(c *gin.Context) {
	id := c.Param("id")
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		return qtx.DeleteBrand(c.Request.Context(), id)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "brand deleted"})
}

// --- Product Models ---

type ProductModelDTO struct {
	ID        string     `json:"id"`
	BrandID   string     `json:"brand_id"`
	BrandName *string    `json:"brand_name"`
	Name      string     `json:"name"`
	ImageUrl  *string    `json:"image_url"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func ToProductModelDTO(m db.ProductModel) ProductModelDTO {
	return ProductModelDTO{
		ID:        m.ID,
		BrandID:   m.BrandID,
		Name:      m.Name,
		ImageUrl:  NullStringToString(m.ImageUrl),
		CreatedAt: NullTimeToTime(m.CreatedAt),
		UpdatedAt: NullTimeToTime(m.UpdatedAt),
	}
}

func ListProductModelsRowToDTO(m db.ListProductModelsRow) ProductModelDTO {
	return ProductModelDTO{
		ID:        m.ID,
		BrandID:   m.BrandID,
		BrandName: &m.BrandName,
		Name:      m.Name,
		ImageUrl:  NullStringToString(m.ImageUrl),
		CreatedAt: NullTimeToTime(m.CreatedAt),
		UpdatedAt: NullTimeToTime(m.UpdatedAt),
	}
}

func (h *EcommerceHandler) ListProductModels(c *gin.Context) {
	brandID := c.Query("brand_id")
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		if brandID != "" {
			models, err := qtx.ListModelsByBrand(c.Request.Context(), brandID)
			if err != nil {
				return err
			}
			dtoList := make([]ProductModelDTO, 0)
			for _, m := range models {
				dtoList = append(dtoList, ToProductModelDTO(m))
			}
			c.JSON(http.StatusOK, dtoList)
		} else {
			models, err := qtx.ListProductModels(c.Request.Context())
			if err != nil {
				return err
			}
			dtoList := make([]ProductModelDTO, 0)
			for _, m := range models {
				dtoList = append(dtoList, ListProductModelsRowToDTO(m))
			}
			c.JSON(http.StatusOK, dtoList)
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *EcommerceHandler) CreateProductModel(c *gin.Context) {
	var req struct {
		BrandID  string `json:"brand_id" binding:"required"`
		Name     string `json:"name" binding:"required"`
		ImageUrl string `json:"image_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		m, err := qtx.CreateProductModel(c.Request.Context(), db.CreateProductModelParams{
			ID:       uuid.New().String(),
			BrandID:  req.BrandID,
			Name:     req.Name,
			ImageUrl: sql.NullString{String: req.ImageUrl, Valid: req.ImageUrl != ""},
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusCreated, ToProductModelDTO(m))
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *EcommerceHandler) UpdateProductModel(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		BrandID  string `json:"brand_id" binding:"required"`
		Name     string `json:"name" binding:"required"`
		ImageUrl string `json:"image_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		m, err := qtx.UpdateProductModel(c.Request.Context(), db.UpdateProductModelParams{
			ID:       id,
			BrandID:  req.BrandID,
			Name:     req.Name,
			ImageUrl: sql.NullString{String: req.ImageUrl, Valid: req.ImageUrl != ""},
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, ToProductModelDTO(m))
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

// --- Existing methods ---

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

	var notifyDriverUserID string
	var notifyDriverUserIDs []string
	var notifyPayload interface{}

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

		// 3. Calculate Total and Create Order
		orderID := uuid.New().String()
		totalAmount := 0.0
		for _, item := range cartItems {
			price, _ := strconv.ParseFloat(item.Price, 64)
			totalAmount += price * float64(item.Quantity)
		}

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
			_, err = qtx.CreateOrderItem(c.Request.Context(), db.CreateOrderItemParams{
				ID:         uuid.New().String(),
				OrderID:    orderID,
				BusinessID: item.BusinessID,
				ItemID:     item.ItemID,
				ItemType:   item.ItemType,
				Quantity:   item.Quantity,
				UnitPrice:  item.Price,
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

		// Determine if this order requires delivery (food/grocery/liquor)
		isDelivery := false
		var pickupLng, pickupLat float64
		for _, item := range cartItems {
			var miniservice sql.NullString
			var lng, lat sql.NullFloat64
			if err := tx.QueryRowContext(c.Request.Context(), `SELECT b.miniservice_type, a.longitude, a.latitude FROM businesses b JOIN addresses a ON a.id = b.address_id WHERE b.id = $1 LIMIT 1`, item.BusinessID).Scan(&miniservice, &lng, &lat); err == nil {
				if miniservice.Valid && (miniservice.String == "food" || miniservice.String == "grocery" || miniservice.String == "liquor") {
					isDelivery = true
					if lng.Valid && lat.Valid {
						pickupLng = lng.Float64
						pickupLat = lat.Float64
					}
					break
				}
			}
		}

		if isDelivery {
			// Try to auto-assign the nearest motorbike driver
			drivers, dErr := qtx.GetNearbyMotorbikeDrivers(c.Request.Context(), db.GetNearbyMotorbikeDriversParams{
				StMakepoint:   pickupLng,
				StMakepoint_2: pickupLat,
				StDwithin:     5000,
				Limit:         10,
			})
			if dErr == nil && len(drivers) > 0 {
				// assign the nearest
				drv := drivers[0]
				_, aErr := qtx.AssignDriverToOrder(c.Request.Context(), db.AssignDriverToOrderParams{
					ID:          orderID,
					DriverID:    sql.NullString{String: drv.ID, Valid: true},
					DeliveryFee: fmt.Sprintf("%.2f", 0.00),
				})
				if aErr == nil {
					// notify the assigned driver after tx
					notifyDriverUserID = drv.UserID
					notifyPayload = gin.H{"type": "delivery_assigned", "order": order}
				}
			} else {
				// No auto-assign; mark as requested and broadcast to nearby motorbike drivers
				if _, uErr := tx.ExecContext(c.Request.Context(), `UPDATE orders SET status='requested', updated_at = CURRENT_TIMESTAMP WHERE id = $1`, orderID); uErr == nil {
					// gather nearby drivers to notify
					if dlist, dlErr := qtx.GetNearbyMotorbikeDrivers(c.Request.Context(), db.GetNearbyMotorbikeDriversParams{
						StMakepoint:   pickupLng,
						StMakepoint_2: pickupLat,
						StDwithin:     5000,
						Limit:         20,
					}); dlErr == nil {
						for _, d := range dlist {
							notifyDriverUserIDs = append(notifyDriverUserIDs, d.UserID)
						}
						notifyPayload = gin.H{"type": "delivery_request", "order": order}
					}
				}
			}
		}

		c.JSON(http.StatusCreated, order)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// After transaction commit, notify drivers (best-effort)
	if DriverHub != nil {
		if notifyDriverUserID != "" {
			DriverHub.SendToDriver(notifyDriverUserID, notifyPayload)
		}
		if len(notifyDriverUserIDs) > 0 {
			DriverHub.BroadcastToDrivers(notifyDriverUserIDs, notifyPayload)
		}
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
