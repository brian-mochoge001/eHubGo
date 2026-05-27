package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
)

type LiquorHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewLiquorHandler(queries *db.Queries, dbConn *sql.DB) *LiquorHandler {
	return &LiquorHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

func (h *LiquorHandler) ListLiquorItems(c *gin.Context) {
	businessID := c.Query("business_id")

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		var items []db.LiquorItem
		var err error

		if businessID != "" {
			items, err = qtx.ListLiquorItemsByBusiness(c.Request.Context(), businessID)
		} else {
			rows, err := qtx.ListLiquorItems(c.Request.Context())
			if err != nil {
				return err
			}
			c.JSON(http.StatusOK, rows)
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

func (h *LiquorHandler) SearchLiquorStores(c *gin.Context) {
	var params struct {
		Latitude  float64 `form:"latitude" binding:"required"`
		Longitude float64 `form:"longitude" binding:"required"`
		Radius    float64 `form:"radius" binding:"required"`
	}

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		stores, err := qtx.SearchStoresByLocation(c.Request.Context(), db.SearchStoresByLocationParams{
			StMakepoint:     params.Longitude,
			StMakepoint_2:   params.Latitude,
			Column3:         []string{"liquor"},
			StDwithin:       params.Radius,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusOK, []interface{}{})
				return nil
			}
			return err
		}

		var results []gin.H
		for _, store := range stores {
			storeLat := ParseCoordinate(store.Latitude)
			storeLng := ParseCoordinate(store.Longitude)

			results = append(results, gin.H{
				"business": store,
				"polyline": GetPointToPointRoute(c.Request.Context(), params.Latitude, params.Longitude, storeLat, storeLng),
			})
		}

		if len(results) == 0 {
			c.JSON(http.StatusOK, []interface{}{})
		} else {
			c.JSON(http.StatusOK, results)
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
