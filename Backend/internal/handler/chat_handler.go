package handler

import (
	"mentalchat/internal/model"
	"mentalchat/internal/service"
	"mentalchat/internal/storage"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ChatHandler struct {
	aiService        *service.AIService
	chatIndexService *service.ChatIndexService
	contextBuilder   *service.ContextBuilder
	storage          *storage.Storage
	messageQueue     *service.MessageQueueService
}

func NewChatHandler(ai *service.AIService, chatIndex *service.ChatIndexService, ctxBuilder *service.ContextBuilder, storage *storage.Storage, msgQueue *service.MessageQueueService) *ChatHandler {
	return &ChatHandler{
		aiService:        ai,
		chatIndexService: chatIndex,
		contextBuilder:   ctxBuilder,
		storage:          storage,
		messageQueue:     msgQueue,
	}
}

type ChatRequest struct {
	ChatType string `json:"chat_type" binding:"required,oneof=psychologist tarot sexologist fortune_teller"`
	Content  string `json:"content" binding:"required"`
}

func (h *ChatHandler) Chat(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user subscription tier
	user, err := h.storage.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	tier := user.Tier

	// Determine device ID from header or fingerprint
	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		deviceID = "unknown"
	}

	// Track device for multi-device detection (non-blocking)
	go func() {
		if _, err := h.storage.UpsertUserDevice(userID, deviceID); err != nil {
			log.Err(err).Msg("Failed to track device")
		}
	}()

	// Use message queue service for debounced, sequential processing
	aiMsg, userMsg, err := h.messageQueue.Enqueue(userID, req.ChatType, req.Content, deviceID, tier)
	if err != nil {
		log.Err(err).Msg("Message queue processing failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process message"})
		return
	}

	// Build response with sequence numbers for client sync
	resp := gin.H{
		"message":    aiMsg.Content,
		"chat_type":  req.ChatType,
		"is_from_ai": true,
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

type HistoryRequest struct {
	ChatType string `json:"chat_type" binding:"required"`
	Limit    int    `json:"limit" binding:"min=1,max=100"`
	Offset   int    `json:"offset" binding:"min=0"`
}

func (h *ChatHandler) GetHistory(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req HistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit == 0 {
		req.Limit = 50
	}

	messages, err := h.storage.GetUserMessages(userID, req.ChatType, req.Limit, req.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	// Get max sequence for sync cursor
	maxSeq, _ := h.storage.GetMaxSequenceForUser(userID, req.ChatType)

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"total":    len(messages),
		"max_seq":  maxSeq,
	})
}

// ---- Pull-based incremental sync ----

type PullRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
	LastSeq  int64  `json:"last_seq"`
}

func (h *ChatHandler) PullMessages(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req PullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deviceID := strings.TrimSpace(req.DeviceID)
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}

	// Track device for multi-device detection
	go func() {
		if _, err := h.storage.UpsertUserDevice(userID, deviceID); err != nil {
			log.Err(err).Msg("Failed to track device on pull")
		}
	}()

	// Get or create sync cursor for this device
	cursor, err := h.storage.GetOrCreateSyncCursor(userID, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sync cursor"})
		return
	}

	// Determine last known sequence: use client-provided or server-stored
	lastKnownSeq := req.LastSeq
	if cursor.LastServerSeq > lastKnownSeq {
		lastKnownSeq = cursor.LastServerSeq
	}

	// Fetch new messages across all chat types since last known sequence
	newMessages, err := h.storage.GetMessagesSince(userID, "", lastKnownSeq, 200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch new messages"})
		return
	}

	// Update cursor
	if len(newMessages) > 0 {
		lastMsg := newMessages[len(newMessages)-1]
		cursor.LastServerSeq = lastMsg.SequenceNumber
		cursor.LastSyncedAt = c.GetTime("now")
		if err := h.storage.UpdateSyncCursor(cursor); err != nil {
			log.Err(err).Msg("Failed to update sync cursor")
		}
	}

	// Get max seq across all chat types for this user
	maxSeq, _ := h.storage.GetMaxSequenceForUser(userID, "")

	c.JSON(http.StatusOK, gin.H{
		"messages": newMessages,
		"count":    len(newMessages),
		"last_seq": cursor.LastServerSeq,
		"max_seq":  maxSeq,
		"has_more": len(newMessages) == 200,
	})
}

// ---- Push-based sync (original, improved with hashing) ----

type SyncMessage struct {
	LocalID  int64  `json:"local_id"`
	ChatID   string `json:"chat_id" binding:"required"`
	Content  string `json:"content" binding:"required"`
	IsFromAI bool   `json:"is_from_ai"`
	Role     string `json:"role" binding:"required,oneof=user ai"`
	Hash     string `json:"hash"`      // Client-computed hash for integrity check
	DeviceID string `json:"device_id"` // Originating device
}

type SyncRequest struct {
	DeviceID string        `json:"device_id"`
	Messages []SyncMessage `json:"messages" binding:"required,min=1,max=100"`
}

func (h *ChatHandler) SyncMessages(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deviceID := strings.TrimSpace(req.DeviceID)
	if deviceID == "" {
		deviceID = c.GetHeader("X-Device-ID")
	}
	if deviceID == "" {
		deviceID = "unknown"
	}

	// Track device for multi-device detection
	go func() {
		if _, err := h.storage.UpsertUserDevice(userID, deviceID); err != nil {
			log.Err(err).Msg("Failed to track device on sync")
		}
	}()

	synced := 0
	failed := 0
	conflicts := 0
	syncedIDs := make([]int64, 0)

	for _, msg := range req.Messages {
		// Validate client hash if provided
		if msg.Hash != "" {
			expectedHash := model.MessageHash(userID, msg.ChatID, msg.Content, msg.Role, c.GetTime("now"), msg.LocalID)
			if msg.Hash != expectedHash {
				log.Warn().
					Int64("local_id", msg.LocalID).
					Str("client_hash", msg.Hash).
					Str("server_hash", expectedHash).
					Msg("Hash mismatch — possible tampering or corruption")
				conflicts++
				continue // Skip tampered messages
			}
		}

		// Index and store with dedup
		_, isNew, err := h.chatIndexService.IndexAndStoreMessage(
			userID, msg.ChatID, msg.Content, msg.Role,
			msg.IsFromAI, msg.LocalID, deviceID,
		)
		if err != nil {
			log.Err(err).
				Int64("local_id", msg.LocalID).
				Msg("Failed to index synced message")
			failed++
			continue
		}

		if isNew {
			syncedIDs = append(syncedIDs, msg.LocalID)
		}
		synced++
	}

	// Update sync cursor
	cursor, err := h.storage.GetOrCreateSyncCursor(userID, deviceID)
	if err == nil {
		cursor.LastSyncedAt = c.GetTime("now")
		if synced > 0 {
			maxSeq, _ := h.storage.GetMaxSequenceForUser(userID, "")
			cursor.LastServerSeq = maxSeq
		}
		h.storage.UpdateSyncCursor(cursor)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Sync completed",
		"synced":     synced,
		"failed":     failed,
		"conflicts":  conflicts,
		"synced_ids": syncedIDs,
	})
}

// ---- Search ----

type SearchRequest struct {
	ChatType   string `json:"chat_type" binding:"required"`
	SearchTerm string `json:"search_term" binding:"required"`
}

func (h *ChatHandler) SearchMessages(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	messages, err := h.storage.SearchMessages(userID, req.ChatType, req.SearchTerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"total":    len(messages),
	})
}
