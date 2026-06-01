package handlers

import (
	"database/sql"
	"net/http"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ComplianceHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewComplianceHandler(queries *db.Queries, dbConn *sql.DB) *ComplianceHandler {
	return &ComplianceHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

// ReviewDocument updates document verification status
func (h *ComplianceHandler) ReviewDocument(c *gin.Context) {
	docID := c.Param("doc_id")
	var req struct {
		Status string `json:"status" binding:"required"`
		Notes  string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		doc, err := qtx.UpdateDocumentStatus(c.Request.Context(), db.UpdateDocumentStatusParams{
			ID:          docID,
			Status:      req.Status,
			ReviewNotes: sql.NullString{String: req.Notes, Valid: req.Notes != ""},
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, doc)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// UpdateApplicationStatus manages the state machine for vendor onboarding
func (h *ComplianceHandler) UpdateApplicationStatus(c *gin.Context) {
	businessID := c.Param("id")
	actorID := c.MustGet("user_id").(string)

	var req struct {
		Status db.BusinessVerificationStatus `json:"status" binding:"required"`
		Reason string                     `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)

		// Get current status for the log
		business, err := qtx.GetBusinessByID(c.Request.Context(), businessID)
		if err != nil {
			return err
		}

		// Update status
		updatedBusiness, err := qtx.UpdateBusinessStatus(c.Request.Context(), db.UpdateBusinessStatusParams{
			ID:                 businessID,
			VerificationStatus: db.NullBusinessVerificationStatus{
				BusinessVerificationStatus: req.Status,
				Valid:                      true,
			},
		})
		if err != nil {
			return err
		}

		// Log the change
		_, err = qtx.CreateVerificationLog(c.Request.Context(), db.CreateVerificationLogParams{
			ID:         uuid.New().String(),
			BusinessID: businessID,
			ActorID:    actorID,
			OldStatus: db.NullBusinessVerificationStatus{
				BusinessVerificationStatus: business.VerificationStatus.BusinessVerificationStatus,
				Valid:                      true,
			},
			NewStatus: req.Status,
			Reason:    sql.NullString{String: req.Reason, Valid: req.Reason != ""},
		})
		if err != nil {
			return err
		}

		c.JSON(http.StatusOK, updatedBusiness)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
