package service

import (
	"errors"
	"mentalchat/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secretKey  string
	issuer     string
	expiration int64
}

type JWTClaims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	Tier      string `json:"tier"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func NewJWTService() *JWTService {
	cfg := config.LoadConfig()
	return &JWTService{
		secretKey:  cfg.Security.JWTSecret,
		issuer:     cfg.App.AppName,
		expiration: int64(cfg.Security.JWTExpirationHours) * 60 * 60,
	}
}

func (s *JWTService) GenerateAccessToken(userID uint, email, tier string) (string, error) {
	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		Tier:      tier,
		TokenType: "access_token",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.expiration) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

func (s *JWTService) GenerateRefreshToken(userID uint, email string) (string, error) {
	claims := RefreshClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 30)), // 30 дней
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

func (s *JWTService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Check if token is an access token
		if claims.TokenType != "access_token" {
			return nil, errors.New("invalid token type")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *JWTService) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *JWTService) GenerateNewAccessTokenFromRefresh(refreshToken string) (string, error) {
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Generate new access token
	return s.GenerateAccessToken(claims.UserID, claims.Email, "free") // Tier will be fetched from DB
}
