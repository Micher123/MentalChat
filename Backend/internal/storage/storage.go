package storage

import (
	"mentalchat/internal/config"
	"mentalchat/internal/model"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ---- In-memory rate limit / DDoS store (package-level singleton) ----

type ipWindow struct {
	count   int
	resetAt time.Time
}

type ipLock struct {
	lockedUntil time.Time
}

type rateLimitStore struct {
	mu          sync.Mutex
	windows     map[string]*ipWindow // IP → sliding window counter
	locks       map[string]*ipLock   // IP → DDoS lock
	concurrent  map[string]int       // IP → concurrent request count
	stopCleanup chan struct{}
}

var (
	globalRateLimitStore *rateLimitStore
	rateLimitOnce        sync.Once
)

func getRateLimitStore() *rateLimitStore {
	rateLimitOnce.Do(func() {
		globalRateLimitStore = &rateLimitStore{
			windows:     make(map[string]*ipWindow),
			locks:       make(map[string]*ipLock),
			concurrent:  make(map[string]int),
			stopCleanup: make(chan struct{}),
		}
		go globalRateLimitStore.cleanupLoop()
	})
	return globalRateLimitStore
}

func (rls *rateLimitStore) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rls.cleanup()
		case <-rls.stopCleanup:
			return
		}
	}
}

func (rls *rateLimitStore) cleanup() {
	rls.mu.Lock()
	defer rls.mu.Unlock()

	now := time.Now()
	for ip, w := range rls.windows {
		if now.After(w.resetAt) {
			delete(rls.windows, ip)
		}
	}
	for ip, l := range rls.locks {
		if now.After(l.lockedUntil) {
			delete(rls.locks, ip)
		}
	}
	for ip, c := range rls.concurrent {
		if c <= 0 {
			delete(rls.concurrent, ip)
		}
	}
}

// ---- Storage ----

type Storage struct {
	db *gorm.DB
}

func NewStorage(db *gorm.DB) *Storage {
	return &Storage{db: db}
}

func NewPostgresDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := buildDSN(cfg)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate models
	if err := db.AutoMigrate(
		&model.User{},
		&model.Message{},
		&model.ChatSession{},
		&model.PaymentTransaction{},
		&model.EmailLog{},
		&model.Subscription{},
		&model.VoiceMessage{},
		&model.UserCache{},
		&model.RateLimit{},
		&model.DDoSEntry{},
		&model.SyncCursor{},
		&model.UserDevice{},
	); err != nil {
		return nil, err
	}

	return db, nil
}

func buildDSN(cfg config.DatabaseConfig) string {
	host := cfg.Host
	if host == "" {
		host = "localhost"
	}
	port := cfg.Port
	if port == "" {
		port = "5432"
	}
	user := cfg.User
	if user == "" {
		user = "mentalchat"
	}
	password := cfg.Password
	if password == "" {
		password = "mentalchat"
	}
	dbname := cfg.Name
	if dbname == "" {
		dbname = "mentalchat"
	}
	sslmode := cfg.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}
	return "host=" + host +
		" user=" + user +
		" password=" + password +
		" dbname=" + dbname +
		" port=" + port +
		" sslmode=" + sslmode
}

// ---- Rate limit operations (in-memory, shared across all Storage instances) ----

func (s *Storage) CheckRateLimit(ip string, windowSeconds int) (bool, error) {
	rls := getRateLimitStore()
	rls.mu.Lock()
	defer rls.mu.Unlock()

	now := time.Now()
	w, exists := rls.windows[ip]

	const maxRequests = 100

	if !exists || now.After(w.resetAt) {
		rls.windows[ip] = &ipWindow{
			count:   0,
			resetAt: now.Add(time.Duration(windowSeconds) * time.Second),
		}
		return true, nil
	}

	return w.count < maxRequests, nil
}

func (s *Storage) IncrementRateLimit(ip string) error {
	rls := getRateLimitStore()
	rls.mu.Lock()
	defer rls.mu.Unlock()

	w, exists := rls.windows[ip]
	if !exists {
		return nil
	}
	w.count++
	return nil
}

// ---- DDoS operations (in-memory, shared across all Storage instances) ----

func (s *Storage) CheckDDoS(ip string) (bool, error) {
	rls := getRateLimitStore()
	rls.mu.Lock()
	defer rls.mu.Unlock()

	now := time.Now()
	l, exists := rls.locks[ip]
	if exists && now.Before(l.lockedUntil) {
		return false, nil
	}
	if exists {
		delete(rls.locks, ip)
	}
	return true, nil
}

func (s *Storage) IncrementDDoS(ip string) error {
	rls := getRateLimitStore()
	rls.mu.Lock()
	defer rls.mu.Unlock()

	rls.concurrent[ip]++
	return nil
}

func (s *Storage) DecrementDDoS(ip string) error {
	rls := getRateLimitStore()
	rls.mu.Lock()
	defer rls.mu.Unlock()

	rls.concurrent[ip]--
	if rls.concurrent[ip] <= 0 {
		delete(rls.concurrent, ip)
	}
	return nil
}

func (s *Storage) LockDDoS(ip string, lockDuration time.Duration) error {
	rls := getRateLimitStore()
	rls.mu.Lock()
	defer rls.mu.Unlock()

	rls.locks[ip] = &ipLock{
		lockedUntil: time.Now().Add(lockDuration),
	}
	delete(rls.concurrent, ip)
	return nil
}

func (s *Storage) GetDDoSCount(ip string) (int, error) {
	rls := getRateLimitStore()
	rls.mu.Lock()
	defer rls.mu.Unlock()

	return rls.concurrent[ip], nil
}

// ---- User operations ----

func (s *Storage) CreateUser(user *model.User) error {
	return s.db.Create(user).Error
}

func (s *Storage) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Storage) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Storage) GetUserByEmailToken(token string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("email_token = ?", token).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Storage) GetUserByFingerprint(fingerprint string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("fingerprint = ?", fingerprint).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Storage) GetAllUsersWithFingerprint() ([]model.User, error) {
	var users []model.User
	if err := s.db.Where("fingerprint IS NOT NULL AND fingerprint != ''").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Storage) UpdateUser(user *model.User) error {
	return s.db.Save(user).Error
}

func (s *Storage) DeleteUser(id uint) error {
	return s.db.Delete(&model.User{}, id).Error
}

// ---- Message operations ----

func (s *Storage) CreateMessage(message *model.Message) error {
	return s.db.Create(message).Error
}

func (s *Storage) GetMessageByID(id uint) (*model.Message, error) {
	var message model.Message
	if err := s.db.Preload("User").First(&message, id).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

func (s *Storage) GetMessageByLocalID(userID uint, localID int64) (*model.Message, error) {
	var message model.Message
	err := s.db.Where("user_id = ? AND local_id = ?", userID, localID).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (s *Storage) GetMessageByContentHash(hash string) (*model.Message, error) {
	var message model.Message
	err := s.db.Where("content_hash = ?", hash).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// CreateMessageWithSequence assigns next sequence number per user+chat_type and creates the message.
// Uses raw SQL to atomically increment without races.
func (s *Storage) CreateMessageWithSequence(message *model.Message) error {
	// Assign next sequence number
	var maxSeq int64
	s.db.Model(&model.Message{}).
		Where("user_id = ? AND chat_type = ?", message.UserID, message.ChatType).
		Select("COALESCE(MAX(sequence_number), 0)").
		Scan(&maxSeq)
	message.SequenceNumber = maxSeq + 1

	// Compute content hash if not set
	if message.ContentHash == "" {
		message.ContentHash = model.MessageHash(
			message.UserID, message.ChatType, message.Content,
			message.Role, message.Timestamp, message.LocalID,
		)
	}
	return s.db.Create(message).Error
}

// GetMessagesSince returns messages for user+chat_type with sequence > afterSeq (for pull sync)
func (s *Storage) GetMessagesSince(userID uint, chatType string, afterSeq int64, limit int) ([]*model.Message, error) {
	var messages []*model.Message
	err := s.db.Where("user_id = ? AND chat_type = ? AND sequence_number > ?", userID, chatType, afterSeq).
		Order("sequence_number ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// GetRecentMessagesForContext returns last N messages for AI context building
func (s *Storage) GetRecentMessagesForContext(userID uint, chatType string, limit int) ([]*model.Message, error) {
	var messages []*model.Message
	err := s.db.Where("user_id = ? AND chat_type = ?", userID, chatType).
		Order("sequence_number DESC").
		Limit(limit).
		Find(&messages).Error
	// Reverse to chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, err
}

// ---- SyncCursor operations ----

func (s *Storage) GetOrCreateSyncCursor(userID uint, deviceID string) (*model.SyncCursor, error) {
	var cursor model.SyncCursor
	err := s.db.Where("user_id = ? AND device_id = ?", userID, deviceID).First(&cursor).Error
	if err == gorm.ErrRecordNotFound {
		cursor = model.SyncCursor{
			UserID:   userID,
			DeviceID: deviceID,
		}
		if err := s.db.Create(&cursor).Error; err != nil {
			return nil, err
		}
		return &cursor, nil
	}
	if err != nil {
		return nil, err
	}
	return &cursor, nil
}

func (s *Storage) UpdateSyncCursor(cursor *model.SyncCursor) error {
	return s.db.Save(cursor).Error
}

func (s *Storage) GetMaxSequenceForUser(userID uint, chatType string) (int64, error) {
	var maxSeq int64
	err := s.db.Model(&model.Message{}).
		Where("user_id = ? AND chat_type = ?", userID, chatType).
		Select("COALESCE(MAX(sequence_number), 0)").
		Scan(&maxSeq).Error
	return maxSeq, err
}

func (s *Storage) GetUserMessages(userID uint, chatType string, limit, offset int) ([]*model.Message, error) {
	var messages []*model.Message
	err := s.db.Where("user_id = ? AND chat_type = ?", userID, chatType).
		Order("timestamp DESC").
		Limit(limit).Offset(offset).
		Preload("User").
		Find(&messages).Error
	return messages, err
}

func (s *Storage) SearchMessages(userID uint, chatType, searchTerm string) ([]*model.Message, error) {
	var messages []*model.Message
	err := s.db.Where("user_id = ? AND chat_type = ? AND content ILIKE ?", userID, chatType, "%"+searchTerm+"%").
		Order("timestamp DESC").
		Preload("User").
		Find(&messages).Error
	return messages, err
}

// ---- Chat session operations ----

func (s *Storage) CreateChatSession(session *model.ChatSession) error {
	return s.db.Create(session).Error
}

func (s *Storage) GetChatSessions(userID uint) ([]*model.ChatSession, error) {
	var sessions []*model.ChatSession
	err := s.db.Where("user_id = ? AND archived = false", userID).
		Order("last_message_time DESC").
		Find(&sessions).Error
	return sessions, err
}

func (s *Storage) UpdateChatSession(session *model.ChatSession) error {
	return s.db.Save(session).Error
}

func (s *Storage) ArchiveChatSession(sessionID uint) error {
	return s.db.Model(&model.ChatSession{}).Where("id = ?", sessionID).Update("archived", true).Error
}

// ---- Payment operations ----

func (s *Storage) CreatePaymentTransaction(txn *model.PaymentTransaction) error {
	return s.db.Create(txn).Error
}

func (s *Storage) GetPaymentTransactionByTransactionID(txnID string) (*model.PaymentTransaction, error) {
	var txn model.PaymentTransaction
	if err := s.db.Where("transaction_id = ?", txnID).First(&txn).Error; err != nil {
		return nil, err
	}
	return &txn, nil
}

func (s *Storage) UpdatePaymentTransaction(txn *model.PaymentTransaction) error {
	return s.db.Save(txn).Error
}

func (s *Storage) GetUserActiveSubscription(userID uint) (*model.Subscription, error) {
	var subscription model.Subscription
	err := s.db.Where("user_id = ? AND status = ?", userID, "active").
		Order("end_date DESC").
		First(&subscription).Error
	return &subscription, err
}

// ---- Email operations ----

func (s *Storage) CreateEmailLog(log *model.EmailLog) error {
	return s.db.Create(log).Error
}

func (s *Storage) UpdateEmailLog(log *model.EmailLog) error {
	return s.db.Save(log).Error
}

// ---- Delete message operations ----

// DeleteMessage removes a single message if it belongs to the user
func (s *Storage) DeleteMessage(userID uint, messageID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", messageID, userID).Delete(&model.Message{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteMessages removes multiple messages belonging to the user
func (s *Storage) DeleteMessages(userID uint, messageIDs []uint) error {
	if len(messageIDs) == 0 {
		return nil
	}
	result := s.db.Where("id IN ? AND user_id = ?", messageIDs, userID).Delete(&model.Message{})
	return result.Error
}

// ClearChatHistory removes all messages for a user's specific chat type
func (s *Storage) ClearChatHistory(userID uint, chatType string) error {
	result := s.db.Where("user_id = ? AND chat_type = ?", userID, chatType).Delete(&model.Message{})
	return result.Error
}

// ---- Subscription operations ----

func (s *Storage) CreateSubscription(subscription *model.Subscription) error {
	return s.db.Create(subscription).Error
}

func (s *Storage) UpdateSubscription(subscription *model.Subscription) error {
	return s.db.Save(subscription).Error
}

func (s *Storage) DeleteSubscription(id uint) error {
	return s.db.Delete(&model.Subscription{}, id).Error
}

// ---- Voice operations ----

func (s *Storage) CreateVoiceMessage(vm *model.VoiceMessage) error {
	return s.db.Create(vm).Error
}

func (s *Storage) GetVoiceMessageByID(id uint) (*model.VoiceMessage, error) {
	var vm model.VoiceMessage
	if err := s.db.Preload("User").Preload("Message").First(&vm, id).Error; err != nil {
		return nil, err
	}
	return &vm, nil
}

// ---- Cache operations ----

func (s *Storage) GetUserCache(userID uint, cacheKey string) (*model.UserCache, error) {
	var cache model.UserCache
	if err := s.db.Where("user_id = ? AND cache_key = ?", userID, cacheKey).First(&cache).Error; err != nil {
		return nil, err
	}
	return &cache, nil
}

func (s *Storage) SetUserCache(cache *model.UserCache) error {
	return s.db.Save(cache).Error
}

func (s *Storage) DeleteUserCache(userID uint, cacheKey string) error {
	return s.db.Where("user_id = ? AND cache_key = ?", userID, cacheKey).Delete(&model.UserCache{}).Error
}

// ---- Fingerprint operations ----

func (s *Storage) UpdateFingerprint(userID uint, fingerprint string) error {
	return s.db.Model(&model.User{}).Where("id = ?", userID).Update("fingerprint", fingerprint).Error
}

// ---- Refresh token operations ----

// ---- UserDevice operations (multi-device tracking) ----

// UpsertUserDevice creates or updates the last-active timestamp for a user device.
// Returns the total number of active devices for this user (counted after upsert).
func (s *Storage) UpsertUserDevice(userID uint, deviceID string) (int64, error) {
	now := time.Now()
	result := s.db.Where("user_id = ? AND device_id = ?", userID, deviceID).
		Assign(map[string]interface{}{"last_active_at": now}).
		FirstOrCreate(&model.UserDevice{
			UserID:       userID,
			DeviceID:     deviceID,
			LastActiveAt: now,
		})
	if result.Error != nil {
		return 0, result.Error
	}

	var count int64
	if err := s.db.Model(&model.UserDevice{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetUserDeviceCount returns the number of distinct devices registered for a user.
func (s *Storage) GetUserDeviceCount(userID uint) (int64, error) {
	var count int64
	err := s.db.Model(&model.UserDevice{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (s *Storage) UpdateRefreshToken(userID uint, token string, expiresAt *time.Time) error {
	return s.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"refresh_token":     token,
		"refresh_token_exp": expiresAt,
	}).Error
}
