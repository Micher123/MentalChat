package service

import (
	"mentalchat/internal/model"
	"mentalchat/internal/storage"
	"time"

	"github.com/rs/zerolog/log"
)

// ChatIndexService handles message hashing, indexing, and sequence management
type ChatIndexService struct {
	store *storage.Storage
}

func NewChatIndexService(store *storage.Storage) *ChatIndexService {
	return &ChatIndexService{store: store}
}

// IndexAndStoreMessage computes hash, assigns sequence, stores with dedup check.
// Returns the stored message (or existing one if duplicate) and whether it was actually new.
func (s *ChatIndexService) IndexAndStoreMessage(userID uint, chatType, content, role string, isFromAI bool, localID int64, deviceID string) (*model.Message, bool, error) {
	// 1. Check dedup by LocalID (per user)
	if localID > 0 {
		existing, err := s.store.GetMessageByLocalID(userID, localID)
		if err == nil && existing != nil && existing.ID > 0 {
			log.Debug().
				Uint("user_id", userID).
				Int64("local_id", localID).
				Msg("Message already indexed (local_id dedup)")
			return existing, false, nil
		}
	}

	// 2. Compute content hash
	now := time.Now()
	contentHash := model.MessageHash(userID, chatType, content, role, now, localID)

	// 3. Check dedup by ContentHash
	existingByHash, err := s.store.GetMessageByContentHash(contentHash)
	if err == nil && existingByHash != nil && existingByHash.ID > 0 {
		log.Debug().
			Str("hash", contentHash).
			Msg("Message already indexed (content hash dedup)")
		return existingByHash, false, nil
	}

	// 4. Create with sequence
	msg := &model.Message{
		UserID:      userID,
		ChatType:    chatType,
		Content:     content,
		ContentHash: contentHash,
		IsFromAI:    isFromAI,
		Role:        role,
		LocalID:     localID,
		DeviceID:    deviceID,
		Timestamp:   now,
		CreatedAt:   now,
	}

	if err := s.store.CreateMessageWithSequence(msg); err != nil {
		return nil, false, err
	}

	log.Info().
		Uint("user_id", userID).
		Str("chat_type", chatType).
		Int64("seq", msg.SequenceNumber).
		Str("hash", contentHash[:16]).
		Msg("Message indexed")

	return msg, true, nil
}
