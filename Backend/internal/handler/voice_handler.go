package handler

import (
	"io"
	"mentalchat/internal/model"
	"mentalchat/internal/service"
	"mentalchat/internal/storage"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type VoiceHandler struct {
	aiService *service.AIService
	storage   *storage.Storage
}

func NewVoiceHandler(ai *service.AIService, storage *storage.Storage) *VoiceHandler {
	return &VoiceHandler{
		aiService: ai,
		storage:   storage,
	}
}

func (h *VoiceHandler) TranscribeVoice(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Err(err).Msg("Failed to get file from request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get audio file"})
		return
	}
	defer file.Close()

	// Validate file type
	if header.Header.Get("Content-Type") != "audio/ogg" &&
		header.Header.Get("Content-Type") != "audio/webm" &&
		header.Header.Get("Content-Type") != "audio/wav" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audio format. Supported: ogg, webm, wav"})
		return
	}

	// Create storage directory if not exists
	storagePath := "./storage/voices"
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		log.Err(err).Msg("Failed to create storage directory")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save audio file"})
		return
	}

	// Save file
	fileName := time.Now().Format("20060102_150405") + "_" + header.Filename
	filePath := filepath.Join(storagePath, fileName)

	dst, err := os.Create(filePath)
	if err != nil {
		log.Err(err).Msg("Failed to create file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save audio file"})
		return
	}
	defer dst.Close()

	fileSize, err := io.Copy(dst, file)
	if err != nil {
		log.Err(err).Msg("Failed to save file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save audio file"})
		return
	}

	// Read file for transcription
	audioData, err := os.ReadFile(filePath)
	if err != nil {
		log.Err(err).Msg("Failed to read file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process audio file"})
		return
	}

	// Transcribe voice
	transcript, err := h.aiService.TranscribeVoice(audioData)
	if err != nil {
		log.Err(err).Msg("Failed to transcribe voice")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to transcribe voice"})
		return
	}

	// Create voice message record
	voiceMessage := &model.VoiceMessage{
		UserID:     userID,
		FilePath:   filePath,
		FileName:   fileName,
		FileSize:   fileSize,
		Duration:   0, // Calculate from audio metadata if needed
		Transcript: transcript,
	}

	if err := h.storage.CreateVoiceMessage(voiceMessage); err != nil {
		log.Err(err).Msg("Failed to save voice message to database")
	}

	c.JSON(http.StatusOK, gin.H{
		"transcript": transcript,
		"file_name":  fileName,
		"file_size":  fileSize,
	})
}

func (h *VoiceHandler) SaveMicrophonePermission(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req struct {
		Granted bool `json:"granted"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user's microphone permission
	user, err := h.storage.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	user.MicrophonePerm = req.Granted
	if err := h.storage.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Microphone permission saved",
		"granted": req.Granted,
	})
}
