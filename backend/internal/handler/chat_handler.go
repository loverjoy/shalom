package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"shalom/internal/models"
	"shalom/internal/service"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

func (h *ChatHandler) SendTextMessage(c *gin.Context) {
	var req struct {
		ChatRoomID string     `json:"chat_room_id" binding:"required"`
		Content    string     `json:"content" binding:"required"`
		ReplyTo    *uuid.UUID `json:"reply_to"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	username := c.GetString("username")

	msg, err := h.chatService.SendTextMessage(c.Request.Context(), req.ChatRoomID, userID, username, req.Content, req.ReplyTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, msg)
}

func (h *ChatHandler) SendFileMessage(c *gin.Context) {
	var req struct {
		ChatRoomID string `json:"chat_room_id" binding:"required"`
		FileURL    string `json:"file_url" binding:"required"`
		FileName   string `json:"file_name" binding:"required"`
		FileSize   int64  `json:"file_size"`
		MsgType    string `json:"msg_type"` // file, image, video
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	username := c.GetString("username")

	msgType := models.MessageTypeFile
	if req.MsgType == "image" {
		msgType = models.MessageTypeImage
	} else if req.MsgType == "video" {
		msgType = models.MessageTypeVideo
	}

	msg, err := h.chatService.SendFileMessage(c.Request.Context(), req.ChatRoomID, userID, username, req.FileURL, req.FileName, req.FileSize, msgType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, msg)
}

func (h *ChatHandler) SendVoiceMessage(c *gin.Context) {
	var req struct {
		ChatRoomID string `json:"chat_room_id" binding:"required"`
		FileURL    string `json:"file_url" binding:"required"`
		Duration   int    `json:"duration" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	username := c.GetString("username")

	msg, err := h.chatService.SendVoiceMessage(c.Request.Context(), req.ChatRoomID, userID, username, req.FileURL, req.Duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, msg)
}

func (h *ChatHandler) SendPoll(c *gin.Context) {
	var req struct {
		ChatRoomID  string   `json:"chat_room_id" binding:"required"`
		Question    string   `json:"question" binding:"required"`
		Options     []string `json:"options" binding:"required,min=2"`
		IsAnonymous bool     `json:"is_anonymous"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	username := c.GetString("username")

	msg, err := h.chatService.SendPollMessage(c.Request.Context(), req.ChatRoomID, userID, username, req.Question, req.Options, req.IsAnonymous)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, msg)
}

func (h *ChatHandler) VotePoll(c *gin.Context) {
	var req struct {
		MessageID   string `json:"message_id" binding:"required"`
		OptionIndex int    `json:"option_index" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.chatService.VotePoll(c.Request.Context(), req.MessageID, userID, req.OptionIndex); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "vote recorded"})
}

func (h *ChatHandler) AddReaction(c *gin.Context) {
	var req struct {
		MessageID string `json:"message_id" binding:"required"`
		Emoji     string `json:"emoji" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.chatService.AddReaction(c.Request.Context(), req.MessageID, userID, req.Emoji); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reaction added"})
}

func (h *ChatHandler) RemoveReaction(c *gin.Context) {
	var req struct {
		MessageID string `json:"message_id" binding:"required"`
		Emoji     string `json:"emoji" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.chatService.RemoveReaction(c.Request.Context(), req.MessageID, userID, req.Emoji); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reaction removed"})
}

func (h *ChatHandler) EditMessage(c *gin.Context) {
	var req struct {
		MessageID string `json:"message_id" binding:"required"`
		Content   string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.chatService.EditMessage(c.Request.Context(), req.MessageID, userID, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "edited"})
}

func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	var req struct {
		MessageID string `json:"message_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.chatService.DeleteMessage(c.Request.Context(), req.MessageID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	chatRoomID := c.Param("roomId")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.ParseInt(limitStr, 10, 64)

	var before *time.Time
	if beforeStr := c.Query("before"); beforeStr != "" {
		t, err := time.Parse(time.RFC3339, beforeStr)
		if err == nil {
			before = &t
		}
	}

	messages, err := h.chatService.GetMessages(c.Request.Context(), chatRoomID, limit, before)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *ChatHandler) CreateChatRoom(c *gin.Context) {
	var req struct {
		Type      string   `json:"type" binding:"required"`
		Name      string   `json:"name" binding:"required"`
		MeetingID string   `json:"meeting_id"`
		Members   []string `json:"members"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	room, err := h.chatService.CreateChatRoom(c.Request.Context(), req.Type, req.Name, req.MeetingID, userID, req.Members)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, room)
}

func (h *ChatHandler) GetChatRooms(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	rooms, err := h.chatService.GetChatRooms(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rooms)
}
