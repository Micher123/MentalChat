package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mentalchat/internal/config"
	"mentalchat/internal/model"
	"mentalchat/internal/service"
	"mentalchat/internal/storage"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type FileHandler struct {
	aiService        *service.AIService
	chatIndexService *service.ChatIndexService
	contextBuilder   *service.ContextBuilder
	storage          *storage.Storage
	messageQueue     *service.MessageQueueService
}

func NewFileHandler(
	ai *service.AIService,
	chatIndex *service.ChatIndexService,
	ctxBuilder *service.ContextBuilder,
	store *storage.Storage,
	msgQueue *service.MessageQueueService,
) *FileHandler {
	return &FileHandler{
		aiService:        ai,
		chatIndexService: chatIndex,
		contextBuilder:   ctxBuilder,
		storage:          store,
		messageQueue:     msgQueue,
	}
}

// UploadWithMessage handles multipart file upload together with an optional message.
// POST /api/v1/chat/upload
// Form fields: file (required), chat_type (required), content (optional)
func (h *FileHandler) UploadWithMessage(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	chatType := strings.TrimSpace(c.PostForm("chat_type"))
	if chatType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chat_type is required"})
		return
	}
	validTypes := map[string]bool{
		"psychologist":   true,
		"tarot":          true,
		"sexologist":     true,
		"fortune_teller": true,
	}
	if !validTypes[chatType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_type"})
		return
	}

	content := strings.TrimSpace(c.PostForm("content"))

	// --- File handling ---
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// Validate mime type
	mimeType := header.Header.Get("Content-Type")
	if !model.AllowedImageMimeTypes[mimeType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("unsupported file type: %s. Allowed: jpeg, png, webp, gif", mimeType),
		})
		return
	}

	// Validate file size (use MaxUploadFileSizeMB if set, fallback to MaxFileSizeMB)
	cfg := config.LoadConfig()
	maxMB := cfg.Storage.MaxUploadFileSizeMB
	if maxMB <= 0 {
		maxMB = cfg.Storage.MaxFileSizeMB
	}
	if maxMB <= 0 {
		maxMB = 20
	}
	maxSize := int64(maxMB) * 1024 * 1024
	if header.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("file too large: %d bytes (max %d MB)", header.Size, maxMB),
		})
		return
	}

	// Create uploads directory
	uploadsPath := cfg.Storage.UploadsPath
	if uploadsPath == "" {
		uploadsPath = "./storage/uploads"
	}
	if err := os.MkdirAll(filepath.Join(uploadsPath, chatType), 0755); err != nil {
		log.Err(err).Msg("Failed to create uploads directory")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store file"})
		return
	}

	// Generate unique filename
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		switch mimeType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		case "image/gif":
			ext = ".gif"
		default:
			ext = ".bin"
		}
	}
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		log.Err(err).Msg("Failed to generate random filename")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store file"})
		return
	}
	savedName := hex.EncodeToString(randBytes) + ext
	savedPath := filepath.Join(uploadsPath, chatType, savedName)

	// Save file to disk
	dst, err := os.Create(savedPath)
	if err != nil {
		log.Err(err).Msg("Failed to create file on disk")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store file"})
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(savedPath)
		log.Err(err).Msg("Failed to write file to disk")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store file"})
		return
	}

	// Save FileAttachment record
	attachment := &model.FileAttachment{
		UserID:   userID,
		ChatType: chatType,
		FilePath: savedPath,
		FileName: header.Filename,
		FileSize: written,
		MimeType: mimeType,
	}
	if err := h.storage.CreateFileAttachment(attachment); err != nil {
		os.Remove(savedPath)
		log.Err(err).Msg("Failed to save file attachment record")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store file"})
		return
	}

	// Build augmented content for AI
	// AI cannot see images directly — append descriptive note
	imageNote := fmt.Sprintf("[Пользователь прикрепил изображение: %s (%s, %d байт)]",
		header.Filename, mimeType, written)
	augmentedContent := content
	if augmentedContent == "" {
		augmentedContent = imageNote
	} else {
		augmentedContent = content + "\n" + imageNote
	}

	// Determine device ID
	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		deviceID = "unknown"
	}

	// Get user subscription tier
	user, err := h.storage.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	tier := user.Tier

	// Track device (non-blocking)
	go func() {
		if _, err := h.storage.UpsertUserDevice(userID, deviceID); err != nil {
			log.Err(err).Msg("Failed to track device")
		}
	}()

	// Process through message queue with augmented content
	aiMsg, userMsg, err := h.messageQueue.Enqueue(userID, chatType, augmentedContent, deviceID, tier)
	if err != nil {
		log.Err(err).Msg("Message queue processing failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process message"})
		return
	}

	// If queued
	if aiMsg == nil {
		c.JSON(http.StatusOK, gin.H{
			"chat_type":      chatType,
			"queued":         true,
			"file_id":        attachment.ID,
			"file_name":      header.Filename,
			"file_mime_type": mimeType,
		})
		return
	}

	resp := gin.H{
		"message":        aiMsg.Content,
		"chat_type":      chatType,
		"is_from_ai":     true,
		"file_id":        attachment.ID,
		"file_name":      header.Filename,
		"file_mime_type": mimeType,
	}

	if userMsg != nil {
		resp["user_seq"] = userMsg.SequenceNumber
		resp["user_hash"] = userMsg.ContentHash
	}
	if aiMsg != nil {
		resp["ai_seq"] = aiMsg.SequenceNumber
		resp["ai_hash"] = aiMsg.ContentHash
	}

	c.JSON(http.StatusOK, resp)
}
