package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"shalom/internal/service"
)

type MeetingHandler struct {
	meetingService *service.MeetingService
}

func NewMeetingHandler(meetingService *service.MeetingService) *MeetingHandler {
	return &MeetingHandler{meetingService: meetingService}
}

func (h *MeetingHandler) CreateMeeting(c *gin.Context) {
	var req service.CreateMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	meeting, err := h.meetingService.CreateMeeting(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, meeting)
}

func (h *MeetingHandler) StartMeeting(c *gin.Context) {
	meetingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meeting id"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	resp, err := h.meetingService.StartMeeting(c.Request.Context(), meetingID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *MeetingHandler) JoinMeeting(c *gin.Context) {
	var req service.JoinMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	resp, err := h.meetingService.JoinMeeting(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *MeetingHandler) EndMeeting(c *gin.Context) {
	meetingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meeting id"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.meetingService.EndMeeting(c.Request.Context(), meetingID, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "meeting ended"})
}

func (h *MeetingHandler) GetMeeting(c *gin.Context) {
	meetingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meeting id"})
		return
	}

	meeting, err := h.meetingService.GetMeeting(c.Request.Context(), meetingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "meeting not found"})
		return
	}

	c.JSON(http.StatusOK, meeting)
}

func (h *MeetingHandler) GetParticipants(c *gin.Context) {
	meetingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meeting id"})
		return
	}

	participants, err := h.meetingService.GetParticipants(c.Request.Context(), meetingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, participants)
}

func (h *MeetingHandler) UpdateParticipant(c *gin.Context) {
	meetingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meeting id"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		IsMuted      *bool `json:"is_muted"`
		IsVideoOn    *bool `json:"is_video_on"`
		IsScreenShare *bool `json:"is_screen_share"`
		HandRaised   *bool `json:"hand_raised"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.meetingService.UpdateParticipant(c.Request.Context(), meetingID, userID, req.IsMuted, req.IsVideoOn, req.IsScreenShare, req.HandRaised); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *MeetingHandler) KickParticipant(c *gin.Context) {
	meetingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meeting id"})
		return
	}

	targetID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	hostID := c.MustGet("user_id").(uuid.UUID)
	if err := h.meetingService.KickParticipant(c.Request.Context(), meetingID, hostID, targetID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "participant kicked"})
}
