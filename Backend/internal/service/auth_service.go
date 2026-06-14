package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mentalchat/internal/config"
	"mentalchat/internal/model"
	"mentalchat/internal/storage"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	storage            *storage.Storage
	jwtService         *JWTService
	emailService       *EmailService
	fingerprintService *FingerprintService
	cfg                *config.AppConfig
}

type RegisterInput struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	DisplayName     string `json:"display_name" binding:"required"`
	Fingerprint     string `json:"fingerprint"`
	FingerprintData string `json:"fingerprint_data"`
}

func NewAuthService(storage *storage.Storage, jwtService *JWTService, emailService *EmailService, fingerprintService *FingerprintService, cfg *config.AppConfig) *AuthService {
	return &AuthService{
		storage:            storage,
		jwtService:         jwtService,
		emailService:       emailService,
		fingerprintService: fingerprintService,
		cfg:                cfg,
	}
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (s *AuthService) VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (s *AuthService) GenerateEmailToken() (string, error) {
	token := generateRandomString(32)
	return token, nil
}

func (s *AuthService) GenerateJWT(userID uint, email, tier string) (string, error) {
	return s.jwtService.GenerateAccessToken(userID, email, tier)
}

func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	return s.jwtService.ValidateAccessToken(tokenString)
}

func (s *AuthService) GenerateRefreshToken(userID uint, email string) (string, error) {
	return s.jwtService.GenerateRefreshToken(userID, email)
}

func (s *AuthService) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	return s.jwtService.ValidateRefreshToken(tokenString)
}

func (s *AuthService) Login(email, password string) (*model.User, error) {
	user, err := s.storage.GetUserByEmail(email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if err := s.VerifyPassword(user.Password, password); err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}

func (s *AuthService) RegisterUser(input RegisterInput, fingerprintData map[string]interface{}) (*model.User, error) {
	// Проверка на существование пользователя (основная проверка)
	if _, err := s.storage.GetUserByEmail(input.Email); err == nil {
		return nil, ErrUserAlreadyExists
	}

	// Fingerprint проверка только для мониторинга (не блокируем)
	if input.Fingerprint != "" {
		// Проверяем схожесть fingerprint с существующими пользователями
		suspiciousUsers, err := s.checkFingerprintSimilarity(input.Fingerprint, fingerprintData)
		if err == nil && len(suspiciousUsers) > 0 {
			// Только логируем, не блокируем регистрацию
			log.Debug(). // Changed from Warn to Debug
					Str("email", input.Email).
					Int("suspicious_users", len(suspiciousUsers)).
					Msg("Fingerprint check completed - registration allowed")
		}
	}

	// Hash password
	passwordHash, err := s.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Generate email token
	emailToken, err := s.GenerateEmailToken()
	if err != nil {
		return nil, err
	}

	// Create user
	emailTokenExp := time.Now().Add(24 * time.Hour)

	// Trial period только для PRO и ULTRA будет активирован при выборе тарифа
	var trialEnd *time.Time = nil

	user := &model.User{
		Email:          input.Email,
		Password:       passwordHash,
		DisplayName:    input.DisplayName,
		MentalState:    "",
		Tier:           "free",
		Verified:       false,
		EmailToken:     emailToken,
		EmailTokenExp:  &emailTokenExp,
		MarketingEmail: false,
		PrivacyPolicy:  true,
		TrialStart:     nil,
		TrialEnd:       trialEnd,
		Fingerprint:    input.Fingerprint,
	}

	// Сохраняем fingerprint данные если есть
	if len(fingerprintData) > 0 {
		fpJSON, _ := json.Marshal(fingerprintData)
		user.FingerprintData = string(fpJSON)
	}

	if err := s.storage.CreateUser(user); err != nil {
		return nil, err
	}

	log.Info().
		Str("email", input.Email).
		Str("fingerprint", input.Fingerprint).
		Msg("User registered successfully")

	return user, nil
}

// checkFingerprintSimilarity проверяет схожесть fingerprint с существующими пользователями
func (s *AuthService) checkFingerprintSimilarity(fingerprint string, data map[string]interface{}) ([]model.User, error) {
	if s.fingerprintService == nil {
		return nil, nil
	}

	// Получаем всех пользователей с fingerprint
	users, err := s.storage.GetAllUsersWithFingerprint()
	if err != nil {
		return nil, err
	}

	var suspicious []model.User
	for _, user := range users {
		if user.Fingerprint == "" {
			continue
		}

		// Если fingerprint совпадает точно
		if user.Fingerprint == fingerprint {
			suspicious = append(suspicious, user)
			continue
		}

		// Если fingerprint похож (более 80% схожесть)
		// Здесь можно добавить более сложную логику сравнения
	}

	return suspicious, nil
}

func (s *AuthService) VerifyEmail(email, token string) error {
	user, err := s.storage.GetUserByEmail(email)
	if err != nil {
		return ErrUserNotFound
	}

	if user.EmailToken != token {
		return ErrInvalidToken
	}

	if time.Now().After(*user.EmailTokenExp) {
		return ErrTokenExpired
	}

	user.Verified = true
	user.EmailToken = ""
	user.EmailTokenExp = nil

	return s.storage.UpdateUser(user)
}

func (s *AuthService) RequestPasswordReset(email string) error {
	user, err := s.storage.GetUserByEmail(email)
	if err != nil {
		// Don't reveal if user exists
		return nil
	}

	// Generate reset token
	resetToken, err := s.GenerateEmailToken()
	if err != nil {
		return err
	}

	resetTokenExp := time.Now().Add(1 * time.Hour)
	user.EmailToken = resetToken
	user.EmailTokenExp = &resetTokenExp

	if err := s.storage.UpdateUser(user); err != nil {
		return err
	}

	// Send reset email
	resetURL := fmt.Sprintf("%s/reset-password?email=%s&token=%s",
		s.emailService.GetFrontendURL(), email, resetToken)

	if err := s.emailService.SendPasswordResetEmail(email, resetURL); err != nil {
		log.Err(err).Str("email", email).Msg("Failed to send password reset email")
		// Don't fail the request if email sending fails for security reasons
	}

	return nil
}

func (s *AuthService) GetFrontendURL() string {
	return s.emailService.GetFrontendURL()
}

func (s *AuthService) ResetPassword(token, newPassword string) error {
	user, err := s.storage.GetUserByEmailToken(token)
	if err != nil {
		return ErrInvalidToken
	}

	if user.EmailToken != token {
		return ErrInvalidToken
	}

	if user.EmailTokenExp != nil && time.Now().After(*user.EmailTokenExp) {
		return ErrTokenExpired
	}

	// Hash new password
	passwordHash, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = passwordHash
	user.EmailToken = ""
	user.EmailTokenExp = nil

	return s.storage.UpdateUser(user)
}

// Error variables
var (
	ErrUserAlreadyExists = &Error{CodeValue: "user_already_exists", Message: "User already exists"}
	ErrUserNotFound      = &Error{CodeValue: "user_not_found", Message: "User not found"}
	ErrInvalidToken      = &Error{CodeValue: "invalid_token", Message: "Invalid token"}
	ErrTokenExpired      = &Error{CodeValue: "token_expired", Message: "Token expired"}
	ErrInvalidPassword   = &Error{CodeValue: "invalid_password", Message: "Invalid password"}
)

type Error struct {
	CodeValue string
	Message   string
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Code() string {
	return e.CodeValue
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// Fallback — should never happen, but avoid panic
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}
