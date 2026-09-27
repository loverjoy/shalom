package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"shalom/internal/models"
	"shalom/internal/service"
)

type BandwidthHandler struct {
	bandwidth *service.BandwidthOptimizer
}

func NewBandwidthHandler(bandwidth *service.BandwidthOptimizer) *BandwidthHandler {
	return &BandwidthHandler{bandwidth: bandwidth}
}

func (h *BandwidthHandler) GetProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	profile, err := h.bandwidth.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *BandwidthHandler) SwitchMode(c *gin.Context) {
	var req struct {
		Mode string `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	mode := models.BandwidthMode(req.Mode)
	profile := h.bandwidth.SwitchMode(c.Request.Context(), userID, mode)
	c.JSON(http.StatusOK, profile)
}

func (h *BandwidthHandler) ReportNetwork(c *gin.Context) {
	var req struct {
		BandwidthKbps int `json:"bandwidth_kbps" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	condition := h.bandwidth.DetectNetworkCondition(c.Request.Context(), userID, req.BandwidthKbps)
	profile := h.bandwidth.GetProfile(c.Request.Context(), userID)

	c.JSON(http.StatusOK, gin.H{
		"condition": condition,
		"profile":   profile,
	})
}

func (h *BandwidthHandler) GetPresets(c *gin.Context) {
	c.JSON(http.StatusOK, service.QualityPresets)
}
