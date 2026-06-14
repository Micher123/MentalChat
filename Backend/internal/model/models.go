package model

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Email           string         `gorm:"unique;not null;size:255" json:"email"`
	Password        string         `gorm:"not null;size:255" json:"-"`
	DisplayName     string         `gorm:"size:255" json:"display_name"`
	MentalState     string         `gorm:"size:255" json:"mental_state"`
	MarketingEmail  bool           `gorm:"default:false" json:"marketing_email"`
	MicrophonePerm  bool           `gorm:"default:false" json:"microphone_perm"`
	Role            string         `gorm:"default:user;size:50" json:"role"`
	Tier            string         `gorm:"default:free;size:20" json:"tier"`
	TrialStart      *time.Time     `json:"trial_start"`
	TrialEnd        *time.Time     `json:"trial_end"`
	SubscriptionID  *string        `gorm:"size:255" json:"subscription_id"`
	Verified        bool           `gorm:"default:false" json:"verified"`
	VerifiedAt      *time.Time     `json:"verified_at"`
	EmailToken      string         `gorm:"size:255" json:"-"`
	EmailTokenExp   *time.Time     `json:"-"`
	PrivacyPolicy   bool           `gorm:"default:false" json:"-"`
	Fingerprint     string         `gorm:"size:255" json:"fingerprint"`
	FingerprintData string         `gorm:"type:text" json:"fingerprint_data"`
	LastLoginAt     *time.Time     `json:"last_login_at"`
	LastLoginIP     string         `gorm:"size:45" json:"last_login_ip"`
	RefreshToken    string         `gorm:"size:500" json:"-"`
	RefreshTokenExp *time.Time     `json:"-"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type Message struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	UserID         uint      `gorm:"index:idx_messages_user_local,unique" json:"user_id"`
	User           User      `gorm:"foreignKey:UserID" json:"user"`
	ChatType       string    `gorm:"index:idx_messages_chat_type_seq;index:idx_messages_user_ts" json:"chat_type"`
	Content        string    `json:"content"`
	ContentHash    string    `gorm:"size:64;index" json:"content_hash"` // SHA-256 for dedup & integrity
	IsFromAI       bool      `json:"is_from_ai"`
	Role           string    `json:"role"`
	LocalID        int64     `gorm:"index:idx_messages_user_local,unique" json:"local_id"`    // Client-side ID for dedup
	DeviceID       string    `gorm:"size:64;index:idx_messages_user_ts" json:"device_id"`     // Originating device
	SequenceNumber int64     `gorm:"index:idx_messages_chat_type_seq" json:"sequence_number"` // Monotonic per chat_type
	Timestamp      time.Time `gorm:"index:idx_messages_user_ts" json:"timestamp"`
	CreatedAt      time.Time `json:"created_at"`
}

type ChatSession struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	UserID          uint      `json:"user_id"`
	User            User      `gorm:"foreignKey:UserID" json:"user"`
	ChatType        string    `json:"chat_type"`
	Title           string    `json:"title"`
	LastMessage     string    `json:"last_message"`
	LastMessageTime time.Time `json:"last_message_time"`
	Archived        bool      `json:"archived"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PaymentTransaction struct {
	ID            uint       `gorm:"primarykey" json:"id"`
	UserID        uint       `json:"user_id"`
	User          User       `gorm:"foreignKey:UserID" json:"user"`
	TransactionID string     `json:"transaction_id"`
	Tier          string     `json:"tier"`
	Amount        int        `json:"amount"`
	Status        string     `json:"status"`
	PaymentType   string     `json:"payment_type"`
	PaymentMethod string     `json:"payment_method"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at"`
}

type EmailLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	Email     string    `json:"email"`
	Subject   string    `json:"subject"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Error     string    `json:"error"`
	CreatedAt time.Time `json:"created_at"`
}

type Subscription struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	UserID     uint      `json:"user_id"`
	User       User      `gorm:"foreignKey:UserID" json:"user"`
	Tier       string    `json:"tier"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	AutoRenew  bool      `json:"auto_renew"`
	Status     string    `json:"status"`
	Provider   string    `json:"provider"`
	ProviderID string    `json:"provider_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type VoiceMessage struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	UserID     uint      `json:"user_id"`
	User       User      `gorm:"foreignKey:UserID" json:"user"`
	MessageID  uint      `json:"message_id"`
	Message    Message   `gorm:"foreignKey:MessageID" json:"message"`
	FilePath   string    `json:"file_path"`
	FileName   string    `json:"file_name"`
	FileSize   int64     `json:"file_size"`
	Duration   float64   `json:"duration"`
	Transcript string    `json:"transcript"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserCache struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	CacheKey  string    `json:"cache_key"`
	CacheData string    `json:"cache_data"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MessageHash computes SHA-256 hash of message content for dedup/integrity
func MessageHash(userID uint, chatType, content, role string, timestamp time.Time, localID int64) string {
	data := fmt.Sprintf("%d|%s|%s|%s|%d|%d", userID, chatType, content, role, timestamp.UnixMilli(), localID)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SyncCursor tracks per-device sync state
type SyncCursor struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	UserID        uint      `gorm:"uniqueIndex:idx_sync_cursor_user_device" json:"user_id"`
	DeviceID      string    `gorm:"uniqueIndex:idx_sync_cursor_user_device;size:64" json:"device_id"`
	LastSequence  int64     `json:"last_sequence"` // Last seq received from this device
	LastSyncedAt  time.Time `json:"last_synced_at"`
	LastServerSeq int64     `json:"last_server_seq"` // Last seq sent to this device
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ContextEntry — compact context entry for AI
type ContextEntry struct {
	Role       string  `json:"role"`       // "user" | "ai"
	Content    string  `json:"content"`    // Truncated/summarized content
	Hash       string  `json:"hash"`       // ContentHash for dedup
	Importance float64 `json:"importance"` // 0..1 relevance score
}

// UserDevice tracks authenticated devices for multi-device sync detection
type UserDevice struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	UserID       uint      `gorm:"uniqueIndex:idx_user_device_user_dev" json:"user_id"`
	DeviceID     string    `gorm:"uniqueIndex:idx_user_device_user_dev;size:64" json:"device_id"`
	LastActiveAt time.Time `json:"last_active_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// FileAttachment stores uploaded image metadata.
type FileAttachment struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ChatType  string    `gorm:"size:50" json:"chat_type"`
	FilePath  string    `gorm:"size:500" json:"file_path"`
	FileName  string    `gorm:"size:255" json:"file_name"`
	FileSize  int64     `json:"file_size"`
	MimeType  string    `gorm:"size:100" json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

// AllowedImageMimeTypes defines accepted image formats.
var AllowedImageMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

type RateLimit struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	IP          string    `json:"ip"`
	Count       int       `json:"count"`
	WindowStart time.Time `json:"window_start"`
}

type DDoSEntry struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	IP          string    `json:"ip"`
	Count       int       `json:"count"`
	LastRequest time.Time `json:"last_request"`
	Locked      bool      `json:"locked"`
	CreatedAt   time.Time `json:"created_at"`
}
