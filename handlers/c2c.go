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

type C2CHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewC2CHandler(queries *db.Queries, dbConn *sql.DB) *C2CHandler {
	return &C2CHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

type C2CListingDTO struct {
	ID           string     `json:"id"`
	SellerID     string     `json:"seller_id"`
	SellerUserID string     `json:"seller_user_id"`
	SellerAvatar *string    `json:"seller_avatar_url"`
	Title        string     `json:"title"`
	Description  *string    `json:"description"`
	Price        string     `json:"price"`
	Currency     string     `json:"currency"`
	ImageUrls    []string   `json:"image_urls"`
	IsNegotiable bool       `json:"is_negotiable"`
	Location     *string    `json:"location"`
	Condition    *string    `json:"condition"`
	Status       string     `json:"status"`
	CreatedAt    *time.Time `json:"created_at"`
}

type C2CSellerDTO struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Bio       *string   `json:"bio"`
	AvatarUrl *string   `json:"avatar_url"`
	CreatedAt *time.Time `json:"created_at"`
}

// ToC2CListingDTO maps raw db results to JSON-friendly DTO
func ToC2CListingDTO(l db.C2cListing, s db.C2cSeller) C2CListingDTO {
	return C2CListingDTO{
		ID:           l.ID,
		SellerID:     l.SellerID,
		SellerUserID: s.UserID,
		SellerAvatar: NullStringToStringPtr(s.AvatarUrl),
		Title:        l.Title,
		Description:  NullStringToStringPtr(l.Description),
		Price:        l.Price,
		Currency:     l.Currency,
		ImageUrls:    l.ImageUrls,
		IsNegotiable: l.IsNegotiable.Bool,
		Location:     NullStringToStringPtr(l.Location),
		Condition:    NullStringToStringPtr(l.Condition),
		Status:       l.Status,
		CreatedAt:    NullTimeToTimePtr(l.CreatedAt),
	}
}

// ListC2CListings returns available C2C listings
func (h *C2CHandler) ListC2CListings(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)

		// Use the generated ListC2CListings function from sqlc
		rows, err := qtx.ListC2CListings(c.Request.Context())
		if err != nil {
			return err
		}

		// Apply simple manual limit/offset
		start := offset
		if start > len(rows) {
			start = len(rows)
		}
		end := start + limit
		if end > len(rows) {
			end = len(rows)
		}

		pagedRows := rows[start:end]

		dtoList := make([]C2CListingDTO, 0, len(pagedRows))
		for _, r := range pagedRows {
			dtoList = append(dtoList, ToC2CListingDTO(db.C2cListing{
				ID:           r.ID,
				SellerID:     r.SellerID,
				Title:        r.Title,
				Description:  r.Description,
				Price:        r.Price,
				Currency:     r.Currency,
				ImageUrls:    r.ImageUrls,
				IsNegotiable: r.IsNegotiable,
				Location:     r.Location,
				Condition:    r.Condition,
				Status:       r.Status,
				CreatedAt:    r.CreatedAt,
				UpdatedAt:    r.UpdatedAt,
			}, db.C2cSeller{
				UserID:    r.SellerUserID,
				AvatarUrl: r.SellerAvatarUrl,
			}))
		}
		c.JSON(http.StatusOK, dtoList)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetC2CListing returns a single C2C listing
func (h *C2CHandler) GetC2CListing(c *gin.Context) {
	id := c.Param("id")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		listing, err := qtx.GetC2CListingByID(c.Request.Context(), id)
		if err != nil {
			return err
		}

		var seller db.C2cSeller
		err = tx.QueryRowContext(c.Request.Context(), "SELECT id, user_id, bio, avatar_url, created_at FROM c2c_sellers WHERE id = $1", listing.SellerID).Scan(
			&seller.ID, &seller.UserID, &seller.Bio, &seller.AvatarUrl, &seller.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("seller not found: %w", err)
		}

		dto := ToC2CListingDTO(listing, seller)
		c.JSON(http.StatusOK, dto)
		return nil
	})

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "C2C listing not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

// CreateC2CListing creates a new C2C listing
func (h *C2CHandler) CreateC2CListing(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	var req struct {
		Title        string   `json:"title" binding:"required"`
		Description  *string  `json:"description"`
		Price        string   `json:"price" binding:"required"`
		Currency     string   `json:"currency" binding:"required"`
		ImageUrls    []string `json:"image_urls"`
		IsNegotiable bool     `json:"is_negotiable"`
		Location     *string  `json:"location"`
		Condition    *string  `json:"condition"`
		Status       string   `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)

		// 1. Ensure seller exists, create if not
		seller, err := qtx.GetC2CSellerByUserID(c.Request.Context(), userID)
		if err != nil {
			if err == sql.ErrNoRows {
				sellerID := uuid.New().String()
				newSeller, createErr := qtx.CreateC2CSeller(c.Request.Context(), db.CreateC2CSellerParams{
					ID:     sellerID,
					UserID: userID,
				})
				if createErr != nil {
					return fmt.Errorf("failed to create seller: %w", createErr)
				}
				seller = newSeller // Use the newly created seller
			} else {
				return fmt.Errorf("failed to get seller: %w", err)
			}
		}

		// 2. Create listing
		listingID := uuid.New().String()
		listing, err := qtx.CreateC2CListing(c.Request.Context(), db.CreateC2CListingParams{
			ID:           listingID,
			SellerID:     seller.ID,
			Title:        req.Title,
			Description:  sql.NullString{String: *req.Description, Valid: req.Description != nil},
			Price:        req.Price,
			Currency:     req.Currency,
			ImageUrls:    req.ImageUrls,
			IsNegotiable: sql.NullBool{Bool: req.IsNegotiable, Valid: true},
			Location:     sql.NullString{String: *req.Location, Valid: req.Location != nil},
			Condition:    sql.NullString{String: *req.Condition, Valid: req.Condition != nil},
			Status:       req.Status,
		})
		if err != nil {
			return err
		}

		c.JSON(http.StatusCreated, listing)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
