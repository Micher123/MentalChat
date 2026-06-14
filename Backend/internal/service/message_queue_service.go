package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"mentalchat/internal/model"
	"mentalchat/internal/storage"

	"github.com/rs/zerolog/log"
)

// pendingBatch holds accumulated user messages while AI is processing.
// Messages arriving during an active AI call are joined and sent as one prompt
// once the current call completes.
type pendingBatch struct {
	userID   uint
	chatType string
	deviceID string
	messages []string
}

// MessageQueueService ensures sequential per-chat processing:
//   - First message → processed immediately (no delay).
//   - Messages arriving while AI is busy → accumulated and sent as one batch
//     when AI finishes the current request.
//   - Multiple rapid messages before AI starts → idle-timer debounce (10s) to
//     avoid sending half-finished thoughts.
type MessageQueueService struct {
	store          *storage.Storage
	chatIndex      *ChatIndexService
	contextBuilder *ContextBuilder
	aiService      *AIService

	// idleTimeout is how long we wait after the last pre-processing message
	// before kicking off the AI call. Only used for the very first batch
	// when new messages keep arriving before we've started processing.
	idleTimeout time.Duration // default: 10s

	mu sync.Mutex
	// key "userID:chatType"
	pending    map[string]*pendingBatch // messages queued while AI runs
	processing map[string]bool          // AI is currently handling this chat
	results    map[string]chan queueResult
}

type queueResult struct {
	message *model.Message
	userMsg *model.Message
	err     error
}

func NewMessageQueueService(
	store *storage.Storage,
	chatIndex *ChatIndexService,
	ctxBuilder *ContextBuilder,
	ai *AIService,
) *MessageQueueService {
	return &MessageQueueService{
		store:          store,
		chatIndex:      chatIndex,
		contextBuilder: ctxBuilder,
		aiService:      ai,
		idleTimeout:    10 * time.Second,
		pending:        make(map[string]*pendingBatch),
		processing:     make(map[string]bool),
		results:        make(map[string]chan queueResult),
	}
}

// Enqueue adds a message and returns the AI reply.
//
// Rules:
//  1. If AI is NOT processing this chat → start processing immediately
//     (the message is sent to AI right now).
//  2. If AI IS already processing → queue the message. When AI finishes,
//     all queued messages are joined (newline-separated) and sent as one
//     new AI prompt. The caller blocks until its batch is processed.
func (mqs *MessageQueueService) Enqueue(userID uint, chatType, content, deviceID string, userTier string) (*model.Message, *model.Message, error) {
	key := fmt.Sprintf("%d:%s", userID, chatType)

	mqs.mu.Lock()

	// If already processing — queue this message and wait for the next round.
	if mqs.processing[key] {
		batch, exists := mqs.pending[key]
		if !exists {
			batch = &pendingBatch{
				userID:   userID,
				chatType: chatType,
				deviceID: deviceID,
				messages: []string{content},
			}
			mqs.pending[key] = batch
		} else {
			batch.messages = append(batch.messages, content)
		}

		log.Debug().
			Uint("user_id", userID).
			Str("chat_type", chatType).
			Int("queue_size", len(batch.messages)).
			Msg("Message queued while AI is processing")

		// Reuse or create result channel.
		ch, chExists := mqs.results[key]
		if !chExists {
			ch = make(chan queueResult, 1)
			mqs.results[key] = ch
		}
		mqs.mu.Unlock()

		// Wait for the first caller's AI result, but do NOT return the AI message
		// to the client (otherwise every queued caller sends the same AI reply).
		<-ch
		// Return nil for message so the handler sends a "queued" response.
		return nil, nil, nil
	}

	// Not processing — start processing immediately.
	mqs.processing[key] = true
	ch := make(chan queueResult, 1)
	mqs.results[key] = ch
	mqs.mu.Unlock()

	// Process this message immediately (no timer for first message).
	result := mqs.processBatch(key, []string{content}, userID, chatType, deviceID, userTier)

	// Notify the caller.
	ch <- result

	// Now check if more messages arrived while we were processing.
	mqs.mu.Lock()
	delete(mqs.processing, key)
	delete(mqs.results, key)

	pending, hasPending := mqs.pending[key]
	delete(mqs.pending, key)
	mqs.mu.Unlock()

	if hasPending && len(pending.messages) > 0 {
		// Save pending messages as user messages only (do NOT call AI again).
		// They piggyback on the AI response from the first batch.
		joined := strings.Join(pending.messages, "\n")
		now := time.Now()
		log.Debug().
			Uint("user_id", userID).
			Str("chat_type", chatType).
			Int("pending_count", len(pending.messages)).
			Msg("Saving queued messages as user-only (no separate AI call)")

		if mqs.chatIndex != nil {
			backendLocalID := -now.UnixNano()
			_, _, err := mqs.chatIndex.IndexAndStoreMessage(userID, chatType, joined, "user", false, backendLocalID, deviceID)
			if err != nil {
				log.Err(err).Msg("Failed to save queued user message")
			}
		} else {
			userMsg := &model.Message{
				UserID:      userID,
				ChatType:    chatType,
				Content:     joined,
				ContentHash: hashContent(joined),
				IsFromAI:    false,
				Role:        "user",
				DeviceID:    deviceID,
				Timestamp:   now,
				CreatedAt:   now,
			}
			if err := mqs.store.CreateMessage(userMsg); err != nil {
				log.Err(err).Msg("Failed to save queued user message")
			}
		}
	}

	return result.message, result.userMsg, result.err
}

func (mqs *MessageQueueService) processBatch(key string, messages []string, userID uint, chatType, deviceID, userTier string) queueResult {
	joined := strings.Join(messages, "\n")
	now := time.Now()

	// Save user message (with dedup via IndexAndStoreMessage).
	// Use negative local_id for backend-generated messages to avoid
	// unique constraint conflicts (idx_messages_user_local on user_id+local_id).
	var userMsg *model.Message
	if mqs.chatIndex != nil {
		var stored bool
		var err error
		backendLocalID := -now.UnixNano()
		userMsg, stored, err = mqs.chatIndex.IndexAndStoreMessage(userID, chatType, joined, "user", false, backendLocalID, deviceID)
		if err != nil {
			log.Err(err).Msg("Failed to index & store user message")
			return queueResult{err: fmt.Errorf("failed to save user message: %w", err)}
		}
		if !stored {
			log.Debug().Msg("User message already indexed (dedup), reusing existing")
		}
	} else {
		userMsg = &model.Message{
			UserID:      userID,
			ChatType:    chatType,
			Content:     joined,
			ContentHash: hashContent(joined),
			IsFromAI:    false,
			Role:        "user",
			DeviceID:    deviceID,
			Timestamp:   now,
			CreatedAt:   now,
		}
		if err := mqs.store.CreateMessage(userMsg); err != nil {
			log.Err(err).Msg("Failed to save user message")
			return queueResult{err: fmt.Errorf("failed to save user message: %w", err)}
		}
	}

	// Check if the message is relevant to the service themes.
	// Skip filter for image uploads — user explicitly chose to share an image,
	// which is always relevant.
	skipFilter := strings.Contains(joined, "[Пользователь прикрепил изображение:")
	var aiResponse string
	if !skipFilter && !mqs.aiService.CheckContext(joined) {
		log.Info().
			Uint("user_id", userID).
			Str("chat_type", chatType).
			Str("content", joined).
			Msg("Message filtered: not relevant to service themes")
		aiResponse = "Я не знаю, как помочь тебе с этим вопросом."
	} else {
		// Build context from recent chat history.
		contextHistory, _, err := mqs.contextBuilder.BuildContext(userID, chatType, joined)
		if err != nil {
			log.Err(err).Msg("Failed to build context")
			// Non-fatal — continue without context.
			contextHistory = ""
		}

		// Call AI.
		var aiErr error
		aiResponse, aiErr = mqs.aiService.GetAIResponseWithContext(joined, chatType, userTier, contextHistory)
		if aiErr != nil {
			log.Err(aiErr).Msg("AI response failed")
			return queueResult{err: fmt.Errorf("AI response failed: %w", aiErr)}
		}
	}

	// Save AI response and index for search (one call, no duplicate).
	var aiMsg *model.Message
	if mqs.chatIndex != nil {
		backendLocalID := -now.UnixNano() - 1 // ensure different from user message local_id
		var stored bool
		var err error
		aiMsg, stored, err = mqs.chatIndex.IndexAndStoreMessage(userID, chatType, aiResponse, "ai", true, backendLocalID, deviceID)
		if err != nil {
			log.Err(err).Msg("Failed to index & store AI message")
			return queueResult{err: fmt.Errorf("failed to save AI message: %w", err)}
		}
		if !stored {
			log.Debug().Msg("AI message already indexed (dedup), reusing existing")
		}
	} else {
		// Fallback: direct storage without index.
		aiMsg = &model.Message{
			UserID:      userID,
			ChatType:    chatType,
			Content:     aiResponse,
			ContentHash: hashContent(aiResponse),
			IsFromAI:    true,
			Role:        "ai",
			DeviceID:    deviceID,
			Timestamp:   now,
			CreatedAt:   now,
		}
		storeErr := mqs.store.CreateMessage(aiMsg)
		if storeErr != nil {
			log.Err(storeErr).Msg("Failed to save AI message")
			return queueResult{err: fmt.Errorf("failed to save AI message: %w", storeErr)}
		}
	}

	log.Info().
		Uint("user_id", userID).
		Str("chat_type", chatType).
		Int("response_len", len(aiResponse)).
		Msg("AI response processed")

	return queueResult{message: aiMsg, userMsg: userMsg}
}

// hashContent computes SHA-256 hex digest of content.
func hashContent(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}
