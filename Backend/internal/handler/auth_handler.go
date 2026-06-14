package handler

import (
	"fmt"
	"mentalchat/internal/service"
	"mentalchat/internal/storage"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type AuthHandler struct {
	authService    *service.AuthService
	userService    *service.UserService
	paymentService *service.PaymentService
	storage        *storage.Storage
	emailService   *service.EmailService
}

func NewAuthHandler(auth *service.AuthService, user *service.UserService, payment *service.PaymentService, storage *storage.Storage, email *service.EmailService) *AuthHandler {
	return &AuthHandler{
		authService:    auth,
		userService:    user,
		paymentService: payment,
		storage:        storage,
		emailService:   email,
	}
}

type RegisterRequest struct {
	Email           string                 `json:"email" binding:"required,email"`
	Password        string                 `json:"password" binding:"required,min=6"`
	DisplayName     string                 `json:"display_name" binding:"required"`
	MentalState     string                 `json:"mental_state"`
	MarketingEmail  bool                   `json:"marketing_email"`
	PrivacyPolicy   bool                   `json:"privacy_policy" binding:"required"`
	Fingerprint     string                 `json:"fingerprint"`
	FingerprintData map[string]interface{} `json:"fingerprint_data"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Token string `json:"token" binding:"required"`
}

type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type PasswordResetConfirmRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !req.PrivacyPolicy {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Privacy policy must be accepted"})
		return
	}

	// Преобразуем fingerprintData в map
	fpData := make(map[string]interface{})
	if len(req.FingerprintData) > 0 {
		fpData = req.FingerprintData
	}

	// Создаем input для сервиса
	input := service.RegisterInput{
		Email:           req.Email,
		Password:        req.Password,
		DisplayName:     req.DisplayName,
		Fingerprint:     req.Fingerprint,
		FingerprintData: "", // Будет обработано в сервисе
	}

	user, err := h.authService.RegisterUser(input, fpData)
	if err != nil {
		log.Err(err).Str("email", req.Email).Msg("Registration failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send verification email
	verificationURL := fmt.Sprintf("%s/verify-email?email=%s&token=%s",
		h.authService.GetFrontendURL(), req.Email, user.EmailToken)

	if err := h.emailService.SendVerificationEmail(req.Email, verificationURL); err != nil {
		log.Err(err).Str("email", req.Email).Msg("Failed to send verification email")
		// Don't fail registration if email sending fails
	}

	// Generate token
	token, err := h.authService.GenerateJWT(user.ID, user.Email, user.Tier)
	if err != nil {
		log.Err(err).Msg("Failed to generate JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	log.Info().
		Str("email", req.Email).
		Str("fingerprint", req.Fingerprint).
		Msg("User registered successfully")

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful! Please check your email to verify your account.",
		"token":   token,
		"user": map[string]interface{}{
			"id":             user.ID,
			"email":          user.Email,
			"display_name":   user.DisplayName,
			"email_verified": false,
			"tier":           user.Tier,
			"trial_end":      user.TrialEnd,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		log.Err(err).Str("email", req.Email).Msg("Login failed")
		if err == service.ErrUserNotFound || err == service.ErrInvalidPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		}
		return
	}

	// Check if email is verified (отключено для тестов)
	// if !user.Verified {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "Please verify your email first"})
	// 	return
	// }

	// Generate access token
	token, err := h.authService.GenerateJWT(user.ID, user.Email, user.Tier)
	if err != nil {
		log.Err(err).Msg("Failed to generate JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Generate refresh token
	refreshToken, err := h.authService.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		log.Err(err).Msg("Failed to generate refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// Store refresh token in DB
	exp := time.Now().Add(24 * time.Hour * 30)
	if err := h.storage.UpdateRefreshToken(user.ID, refreshToken, &exp); err != nil {
		log.Err(err).Msg("Failed to store refresh token")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"token":         token,
		"refresh_token": refreshToken,
		"user": map[string]interface{}{
			"id":              user.ID,
			"email":           user.Email,
			"display_name":    user.DisplayName,
			"mental_state":    user.MentalState,
			"tier":            user.Tier,
			"verified":        user.Verified,
			"marketing_email": user.MarketingEmail,
		},
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.VerifyEmail(req.Email, req.Token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.RequestPasswordReset(req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset email sent"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (h *AuthHandler) GetTrialInfo(c *gin.Context) {
	tiers := []gin.H{
		{
			"tier":        "free",
			"name":        "FREE",
			"description": "Бесплатный тариф",
			"features":    []string{"Базовые AI функции", "4 чата", "Текстовый ввод"},
			"price":       "0 ₽",
		},
		{
			"tier":        "pro",
			"name":        "PRO",
			"description": "Улучшенная версия",
			"features":    []string{"Улучшенная AI модель", "Неограниченные чаты", "Голосовой ввод", "Приоритетная поддержка"},
			"price":       "499 ₽/мес",
			"trial_days":  3,
		},
		{
			"tier":        "ultra",
			"name":        "ULTRA",
			"description": "Максимальная версия",
			"features":    []string{"Максимальная AI модель", "Неограниченные чаты", "Голосовой ввод", "Персонализированные советы", "Приоритетная поддержка"},
			"price":       "999 ₽/мес",
			"trial_days":  1,
		},
	}

	c.JSON(http.StatusOK, gin.H{"tiers": tiers})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate refresh token
	claims, err := h.authService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	// Check if refresh token matches DB
	user, err := h.storage.GetUserByID(claims.UserID)
	if err != nil || user.RefreshToken != req.RefreshToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Check if refresh token is expired
	if user.RefreshTokenExp != nil && time.Now().After(*user.RefreshTokenExp) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
		return
	}

	// Generate new access token
	newToken, err := h.authService.GenerateJWT(user.ID, user.Email, user.Tier)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": newToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	// Clear refresh token in DB
	if err := h.storage.UpdateRefreshToken(userID, "", nil); err != nil {
		log.Err(err).Msg("Failed to clear refresh token")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) InitiatePayment(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var req struct {
		Tier        string `json:"tier" binding:"required,oneof=pro ultra"`
		PaymentType string `json:"payment_type" binding:"required,oneof=monthly yearly"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	paymentURL, err := h.paymentService.InitiatePaymentFlow(userID, req.Tier, req.PaymentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_url": paymentURL,
		"tier":        req.Tier,
	})
}
