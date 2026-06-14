package service

import (
	"mentalchat/internal/config"
	"mentalchat/internal/model"
	"mentalchat/internal/storage"
	"time"
)

type UserService struct {
	storage *storage.Storage
	email   *EmailService
	cfg     *config.AppConfig
}

func NewUserService(storage *storage.Storage, email *EmailService, cfg *config.AppConfig) *UserService {
	return &UserService{
		storage: storage,
		email:   email,
		cfg:     cfg,
	}
}

func (s *UserService) GetUserProfile(userID uint) (*model.User, error) {
	return s.storage.GetUserByID(userID)
}

func (s *UserService) UpdateUserProfile(userID uint, updates map[string]interface{}) error {
	user, err := s.storage.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Update allowed fields
	if displayName, ok := updates["display_name"].(string); ok {
		user.DisplayName = displayName
	}
	if mentalState, ok := updates["mental_state"].(string); ok {
		user.MentalState = mentalState
	}
	if marketingEmail, ok := updates["marketing_email"].(bool); ok {
		user.MarketingEmail = marketingEmail
	}
	if microphonePerm, ok := updates["microphone_perm"].(bool); ok {
		user.MicrophonePerm = microphonePerm
	}

	return s.storage.UpdateUser(user)
}

func (s *UserService) DeleteUserProfile(userID uint) error {
	return s.storage.DeleteUser(userID)
}

func (s *UserService) GetUserSubscription(userID uint) (*model.Subscription, error) {
	return s.storage.GetUserActiveSubscription(userID)
}

func (s *UserService) GetUserChatSessions(userID uint) ([]*model.ChatSession, error) {
	return s.storage.GetChatSessions(userID)
}

func (s *UserService) GetUserMessages(userID uint, chatType string, limit, offset int) ([]*model.Message, error) {
	return s.storage.GetUserMessages(userID, chatType, limit, offset)
}

func (s *UserService) SearchUserMessages(userID uint, chatType, searchTerm string) ([]*model.Message, error) {
	return s.storage.SearchMessages(userID, chatType, searchTerm)
}

func (s *UserService) GetUserCache(userID uint, cacheKey string) (*model.UserCache, error) {
	return s.storage.GetUserCache(userID, cacheKey)
}

func (s *UserService) SetUserCache(userID uint, cacheKey string, cacheData string, ttl int) error {
	cache := &model.UserCache{
		UserID:    userID,
		CacheKey:  cacheKey,
		CacheData: cacheData,
		ExpiresAt: model.UserCache{}.CreatedAt.Add(time.Duration(ttl) * time.Second),
	}
	return s.storage.SetUserCache(cache)
}

func (s *UserService) DeleteUserCache(userID uint, cacheKey string) error {
	return s.storage.DeleteUserCache(userID, cacheKey)
}

func (s *UserService) UpgradeSubscription(userID uint, tier string) error {
	user, err := s.storage.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Set subscription tier
	user.Tier = tier

	// Set trial end date if free tier
	if tier == "free" {
		return s.storage.UpdateUser(user)
	}

	// For paid tiers, create subscription
	subscription := &model.Subscription{
		UserID:     userID,
		Tier:       tier,
		StartDate:  time.Now(),
		EndDate:    time.Now().Add(30 * 24 * time.Hour), // 1 month
		AutoRenew:  true,
		Status:     "active",
		Provider:   "yoomoney",
		ProviderID: "",
	}

	if err := s.storage.CreateSubscription(subscription); err != nil {
		return err
	}

	return s.storage.UpdateUser(user)
}
