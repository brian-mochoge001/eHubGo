package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	IsFeatured         *bool      `json:"is_featured"`
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
		IsFeatured:         NullBoolToBool(p.IsFeatured),
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
		IsFeatured:         NullBoolToBool(p.IsFeatured),
		IsFlashSale:        NullBoolToBool(p.IsFlashSale),
		DiscountPercentage: NullStringToString(p.DiscountPercentage),
		CreatedAt:          NullTimeToTime(p.CreatedAt),
		UpdatedAt:          NullTimeToTime(p.UpdatedAt),
		CategoryName:       NullStringToString(p.CategoryName),
		BrandName:          NullStringToString(p.BrandName),
		ModelName:          NullStringToString(p.ModelName),
	}
}

func GetFeaturedProductsRowToDTO(p db.GetFeaturedProductsRow) ProductDTO {
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
		IsFeatured:         NullBoolToBool(p.IsFeatured),
		IsFlashSale:        NullBoolToBool(p.IsFlashSale),
		DiscountPercentage: NullStringToString(p.DiscountPercentage),
		CreatedAt:          NullTimeToTime(p.CreatedAt),
		UpdatedAt:          NullTimeToTime(p.UpdatedAt),
		CategoryName:       NullStringToString(p.CategoryName),
		BrandName:          NullStringToString(p.BrandName),
		ModelName:          NullStringToString(p.ModelName),
	}
}

func GetFlashSaleProductsRowToDTO(p db.GetFlashSaleProductsRow) ProductDTO {
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
		IsFeatured:         NullBoolToBool(p.IsFeatured),
		IsFlashSale:        NullBoolToBool(p.IsFlashSale),
		DiscountPercentage: NullStringToString(p.DiscountPercentage),
		CreatedAt:          NullTimeToTime(p.CreatedAt),
		UpdatedAt:          NullTimeToTime(p.UpdatedAt),
		CategoryName:       NullStringToString(p.CategoryName),
		BrandName:          NullStringToString(p.BrandName),
		ModelName:          NullStringToString(p.ModelName),
	}
}

func GetStandardProductsRowToDTO(p db.GetStandardProductsRow) ProductDTO {
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
		IsFeatured:         NullBoolToBool(p.IsFeatured),
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
		IsFeatured:         NullBoolToBool(p.IsFeatured),
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
		IsFeatured:         NullBoolToBool(p.IsFeatured),
		IsFlashSale:        NullBoolToBool(p.IsFlashSale),
		DiscountPercentage: NullStringToString(p.DiscountPercentage),
		CreatedAt:          NullTimeToTime(p.CreatedAt),
		UpdatedAt:          NullTimeToTime(p.UpdatedAt),
		CategoryName:       NullStringToString(p.CategoryName),
		BrandName:          NullStringToString(p.BrandName),
		ModelName:          NullStringToString(p.ModelName),
	}
}

// CategoryDTO
type CategoryDTO struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	ImageUrl    *string    `json:"image_url"`
	ParentID    *string    `json:"parent_id"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

func ToCategoryDTO(c db.Category) CategoryDTO {
	return CategoryDTO{
		ID:          c.ID,
		Name:        c.Name,
		Description: NullStringToString(c.Description),
		ImageUrl:    NullStringToString(c.ImageUrl),
		ParentID:    NullStringToString(c.ParentID),
		CreatedAt:   NullTimeToTime(c.CreatedAt),
		UpdatedAt:   NullTimeToTime(c.UpdatedAt),
	}
}

// BrandDTO
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

func ToBrandDTOFromRow(b db.GetBrandsRow) BrandDTO {
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

// ProductModelDTO
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

func ToProductModelDTOFromRow(m db.ListProductModelsRow) ProductModelDTO {
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

// ListFeaturedProducts
func (h *EcommerceHandler) ListFeaturedProducts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	cacheKey := fmt.Sprintf("ecommerce:featured:%s:%s", limitStr, offsetStr)
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
		products, err := qtx.GetFeaturedProducts(c.Request.Context(), db.GetFeaturedProductsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return err
		}

		dtoList = make([]ProductDTO, 0)
		for _, p := range products {
			dtoList = append(dtoList, GetFeaturedProductsRowToDTO(p))
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = h.Cache.SetJSON(c.Request.Context(), cacheKey, dtoList, 15*time.Minute)
	c.JSON(http.StatusOK, dtoList)
}

// SearchProducts
func (h *EcommerceHandler) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query 'q' is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)

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
				&p.Rating, &p.ReviewCount, &p.DiscountPercentage, &p.CreatedAt, &p.UpdatedAt,
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

// ListProducts
func (h *EcommerceHandler) ListProducts(c *gin.Context) {
    fmt.Println("[DEBUG] Entering ListProducts")
	businessID := c.Query("business_id")
	isFeaturedStr := c.Query("is_featured")
    isFlashSaleStr := c.Query("is_flash_sale")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	cacheKey := fmt.Sprintf("ecommerce:products:%s:%s:%s:%s:%s", businessID, isFeaturedStr, isFlashSaleStr, limitStr, offsetStr)

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
		} else if isFeaturedStr == "true" {
			products, err := qtx.GetFeaturedProducts(c.Request.Context(), db.GetFeaturedProductsParams{
				Limit:  int32(limit),
				Offset: int32(offset),
			})
			if err != nil {
				return err
			}
			for _, p := range products {
				dtoList = append(dtoList, GetFeaturedProductsRowToDTO(p))
			}
		} else if isFlashSaleStr == "true" {
			products, err := qtx.GetFlashSaleProducts(c.Request.Context(), db.GetFlashSaleProductsParams{
				Limit:  int32(limit),
				Offset: int32(offset),
			})
			if err != nil {
				return err
			}
			for _, p := range products {
				dtoList = append(dtoList, GetFlashSaleProductsRowToDTO(p))
			}
		} else {
			products, err := qtx.GetStandardProducts(c.Request.Context(), db.GetStandardProductsParams{
				Limit:  int32(limit),
				Offset: int32(offset),
			})
			if err != nil {
				return err
			}
			for _, p := range products {
				dtoList = append(dtoList, GetStandardProductsRowToDTO(p))
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("[ERROR] ListProducts failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = h.Cache.SetJSON(c.Request.Context(), cacheKey, dtoList, 10*time.Minute)
	c.JSON(http.StatusOK, dtoList)
}

// GetProductByID
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

// CreateProduct
func (h *EcommerceHandler) CreateProduct(c *gin.Context) {
	var req struct {
		BusinessID         string   `json:"business_id"` 
		Name               string   `json:"name" binding:"required"`
		Description        string   `json:"description"`
		Price              string   `json:"price" binding:"required"`
		Currency           string   `json:"currency"`
		StockQuantity      int32    `json:"stock_quantity"`
		CategoryID         string   `json:"category_id"`
		BrandID            string   `json:"brand_id"`
		ModelID            string   `json:"model_id"`
		ImageUrls          []string `json:"image_urls"`
		IsFeatured         bool     `json:"is_featured"`
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

	var businessID string
	if req.BusinessID != "" {
		businessID = req.BusinessID
	} else {
		rolesVal, _ := c.Get("user_roles")
		roles := rolesVal.(string)
		if !strings.Contains(roles, "admin") && !strings.Contains(roles, "staff") && !strings.Contains(roles, "executive_admin") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins/staff can create in-house products without a business"})
			return
		}
		businessID = "in-house" 
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		p, err := qtx.CreateProduct(c.Request.Context(), db.CreateProductParams{
			ID:                 uuid.New().String(),
			BusinessID:         businessID,
			Name:               req.Name,
			Description:        StringToNullString(req.Description),
			Price:              req.Price,
			Currency:           req.Currency,
			StockQuantity:      req.StockQuantity,
			CategoryID:         StringToNullString(req.CategoryID),
			BrandID:            StringToNullString(req.BrandID),
			ModelID:            StringToNullString(req.ModelID),
			ImageUrls:          req.ImageUrls,
			IsFeatured:         sql.NullBool{Bool: req.IsFeatured, Valid: true},
            IsFlashSale:        sql.NullBool{Bool: req.IsFlashSale, Valid: true},
			DiscountPercentage: StringToNullString(req.DiscountPercentage),
		})
		if err != nil {
            fmt.Printf("[ERROR] CreateProduct query failed: %v\n", err)
			return err
		}
		c.JSON(http.StatusCreated, ToProductDTO(p))
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
	}
}

// UpdateProduct
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
		IsFeatured         bool     `json:"is_featured"`
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
			Description:        StringToNullString(req.Description),
			Price:              req.Price,
			StockQuantity:      req.StockQuantity,
			CategoryID:         StringToNullString(req.CategoryID),
			BrandID:            StringToNullString(req.BrandID),
			ModelID:            StringToNullString(req.ModelID),
			ImageUrls:          req.ImageUrls,
			IsFeatured:         sql.NullBool{Bool: req.IsFeatured, Valid: true},
            IsFlashSale:        sql.NullBool{Bool: req.IsFlashSale, Valid: true},
			DiscountPercentage: StringToNullString(req.DiscountPercentage),
		})
		if err != nil {
			return err
		}

        // CRITICAL: Invalidate all ecommerce cache to ensure visibility changes take effect
        _ = h.Cache.Delete(c.Request.Context(), "ecommerce:product:" + id)
        _ = h.Cache.Delete(c.Request.Context(), "ecommerce:featured:*")
        _ = h.Cache.Delete(c.Request.Context(), "ecommerce:products:*")

		c.JSON(http.StatusOK, ToProductDTO(p))
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// DeleteProduct
func (h *EcommerceHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		err := qtx.DeleteProduct(c.Request.Context(), id)
		if err != nil {
			return err
		}

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

// UpdateCartItemQuantity
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

// GetOrders
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

// Checkout
func (h *EcommerceHandler) Checkout(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	var req struct {
		ShippingAddressID string `json:"shipping_address_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	parentOrderID, err := h.OC.OrchestrateCheckout(c.Request.Context(), userID, req.ShippingAddressID, cartItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"parent_order_id": parentOrderID, "message": "Checkout successful"})
}

// GetMiniserviceAnalytics
func (h *EcommerceHandler) GetMiniserviceAnalytics(c *gin.Context) {
	serviceType := c.Query("type")
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

// GetStoreAnalytics
func (h *EcommerceHandler) GetStoreAnalytics(c *gin.Context) {
	businessID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"business_id":     businessID,
		"total_revenue":   45000.0,
		"total_orders":    124,
		"active_users":    85, 
		"active_drivers":  5,  
		"completion_rate": 98.0,
	})
}

// SearchAndFilterProducts searches and filters products
func (h *EcommerceHandler) SearchAndFilterProducts(c *gin.Context) {
	categoryID := c.Query("category_id")
	brandID := c.Query("brand_id")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")
    isFeaturedStr := c.Query("is_featured")
	attrsJSON := c.Query("attributes")

	sqlQuery := `
		SELECT p.id, p.business_id, p.name, p.description, p.price, p.currency, p.stock_quantity, 
		       p.category_id, p.brand_id, p.model_id, p.image_urls, p.rating, p.review_count, 
		       p.is_featured, p.discount_percentage, p.created_at, p.updated_at,
               c.name as category_name, b.name as brand_name, m.name as model_name
		FROM products p
        LEFT JOIN categories c ON p.category_id = c.id
        LEFT JOIN brands b ON p.brand_id = b.id
        LEFT JOIN product_models m ON p.model_id = m.id
		WHERE 1=1
	`
	var args []interface{}
	argCount := 1

	if categoryID != "" {
		sqlQuery += fmt.Sprintf(" AND p.category_id = $%d", argCount)
		args = append(args, categoryID)
		argCount++
	}
	if brandID != "" {
		sqlQuery += fmt.Sprintf(" AND p.brand_id = $%d", argCount)
		args = append(args, brandID)
		argCount++
	}
	if minPrice != "" {
		sqlQuery += fmt.Sprintf(" AND p.price::numeric >= $%d", argCount)
		args = append(args, minPrice)
		argCount++
	}
	if maxPrice != "" {
		sqlQuery += fmt.Sprintf(" AND p.price::numeric <= $%d", argCount)
		args = append(args, maxPrice)
		argCount++
	}
	if attrsJSON != "" {
		sqlQuery += fmt.Sprintf(" AND p.attribute_data @> $%d::jsonb", argCount)
		args = append(args, attrsJSON)
		argCount++
	}

    if isFeaturedStr == "true" {
        sqlQuery += " AND p.is_featured = TRUE"
    } else if isFeaturedStr == "false" || isFeaturedStr == "" {
        sqlQuery += " AND p.is_featured = FALSE"
    }

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(c.Request.Context(), sqlQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		dtoList := make([]ProductDTO, 0)
		for rows.Next() {
			var p db.GetProductsRow
			err := rows.Scan(
				&p.ID, &p.BusinessID, &p.Name, &p.Description, &p.Price, &p.Currency,
				&p.StockQuantity, &p.CategoryID, &p.BrandID, &p.ModelID, pq.Array(&p.ImageUrls),
				&p.Rating, &p.ReviewCount, &p.IsFeatured, &p.DiscountPercentage, &p.CreatedAt, &p.UpdatedAt,
			&p.CategoryName, &p.BrandName, &p.ModelName,
			)
			if err != nil {
				return err
			}
			dtoList = append(dtoList, GetProductsRowToDTO(p))
		}
		c.JSON(http.StatusOK, dtoList)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// Category Management

func (h *EcommerceHandler) ListCategories(c *gin.Context) {
	categories, err := h.Queries.GetCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}
	
	dtos := make([]CategoryDTO, 0)
	for _, cat := range categories {
		dtos = append(dtos, ToCategoryDTO(cat))
	}
	c.JSON(http.StatusOK, dtos)
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
		Description: StringToNullString(req.Description),
		ImageUrl:    StringToNullString(req.ImageUrl),
		ParentID:    StringToNullString(req.ParentID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, ToCategoryDTO(category))
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
		Description: StringToNullString(req.Description),
		ImageUrl:    StringToNullString(req.ImageUrl),
		ParentID:    StringToNullString(req.ParentID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, ToCategoryDTO(category))
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

	dtos := make([]BrandDTO, 0)
	for _, b := range brands {
		dtos = append(dtos, ToBrandDTOFromRow(b))
	}
	c.JSON(http.StatusOK, dtos)
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
		LogoUrl:    StringToNullString(req.LogoUrl),
		CategoryID: StringToNullString(req.CategoryID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create brand"})
		return
	}

	c.JSON(http.StatusCreated, ToBrandDTO(brand))
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
		LogoUrl:    StringToNullString(req.LogoUrl),
		CategoryID: StringToNullString(req.CategoryID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update brand"})
		return
	}

	c.JSON(http.StatusOK, ToBrandDTO(brand))
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

	dtos := make([]ProductModelDTO, 0)
	for _, m := range models {
		dtos = append(dtos, ToProductModelDTOFromRow(m))
	}
	c.JSON(http.StatusOK, dtos)
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
		ImageUrl: StringToNullString(req.ImageUrl),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product model"})
		return
	}

	c.JSON(http.StatusCreated, ToProductModelDTO(model))
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
		ImageUrl: StringToNullString(req.ImageUrl),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product model"})
		return
	}

	c.JSON(http.StatusOK, ToProductModelDTO(model))
}

func (h *EcommerceHandler) DeleteProductModel(c *gin.Context) {
	id := c.Param("id")

	err := h.Queries.DeleteProductModel(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product model"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product model deleted"})
}
