package handler

import (
	"mentalchat/internal/service"
	"mentalchat/internal/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
	storage     *storage.Storage
}

func NewUserHandler(user *service.UserService, storage *storage.Storage) *UserHandler {
	return &UserHandler{userService: user, storage: storage}
}

type UserProfileUpdateRequest struct {
	DisplayName    string `json:"display_name"`
	MentalState    string `json:"mental_state" binding:"oneof=harmony satisfied anxiety stress"`
	MarketingEmail bool   `json:"marketing_email"`
	MicrophonePerm bool   `json:"microphone_perm"`
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	user, err := h.userService.GetUserProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req UserProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.MentalState != "" {
		updates["mental_state"] = req.MentalState
	}
	updates["marketing_email"] = req.MarketingEmail
	updates["microphone_perm"] = req.MicrophonePerm

	if err := h.userService.UpdateUserProfile(userID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func (h *UserHandler) DeleteProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	if err := h.userService.DeleteUserProfile(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile deleted successfully"})
}

func (h *UserHandler) GetChatSessions(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	sessions, err := h.userService.GetUserChatSessions(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chat sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

func (h *UserHandler) ArchiveChatSession(c *gin.Context) {
	var req struct {
		SessionID uint `json:"session_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.storage.ArchiveChatSession(req.SessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive chat session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chat session archived"})
}
