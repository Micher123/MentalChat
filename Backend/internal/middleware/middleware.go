package middleware

import (
	"net/http"
	"strings"
	"time"

	"mentalchat/internal/service"
	"mentalchat/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// RequestLogger middleware
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("duration", duration).
			Msg("Request processed")
	}
}

// AuthMiddleware checks if user is authenticated
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get JWT token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		// Check if it's Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token and get user ID
		jwtService := service.NewJWTService()
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			// Token expired - allow client to refresh
			if strings.Contains(err.Error(), "expired") {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Token expired",
					"code":  "TOKEN_EXPIRED",
				})
				c.Abort()
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_tier", claims.Tier)

		c.Next()
	}
}

// RateLimiter middleware
func RateLimiter(maxRequests int, windowSeconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		store := &storage.Storage{} // RateLimit methods use in-memory singleton, no DB needed

		// Check if IP is rate limited
		allowed, err := store.CheckRateLimit(ip, windowSeconds)
		if err != nil {
			log.Err(err).Str("ip", ip).Msg("Rate limit check failed")
			// Don't block on error, allow request
			c.Next()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded. Please try again later."})
			c.Abort()
			return
		}

		// Increment rate limit counter
		if err := store.IncrementRateLimit(ip); err != nil {
			log.Err(err).Str("ip", ip).Msg("Rate limit increment failed")
		}

		c.Next()
	}
}

// DDoSProtection middleware
func DDoSProtection(maxConcurrentRequests int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		store := &storage.Storage{} // DDoS methods use in-memory singleton, no DB needed

		// Decrement concurrent counter when request finishes
		defer store.DecrementDDoS(ip)

		// Check if IP is locked due to DDoS
		allowed, err := store.CheckDDoS(ip)
		if err != nil {
			log.Err(err).Str("ip", ip).Msg("DDoS check failed")
			c.Next()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please try again later."})
			c.Abort()
			return
		}

		// Increment DDoS counter
		if err := store.IncrementDDoS(ip); err != nil {
			log.Err(err).Str("ip", ip).Msg("DDoS increment failed")
		}

		// Check if threshold exceeded and lock if needed
		count, err := store.GetDDoSCount(ip)
		if err == nil && count > maxConcurrentRequests {
			if err := store.LockDDoS(ip, 5*time.Minute); err != nil {
				log.Err(err).Str("ip", ip).Msg("Failed to lock IP")
			}
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Suspicious activity detected. Access temporarily blocked."})
			c.Abort()
			return
		}

		c.Next()
	}
}
