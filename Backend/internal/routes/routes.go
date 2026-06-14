package routes

import (
	"mentalchat/internal/handler"
	"mentalchat/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupPublicRoutes(router *gin.Engine, authHandler *handler.AuthHandler) {
	v1 := router.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/verify-email", authHandler.VerifyEmail)
			auth.POST("/request-password-reset", authHandler.RequestPasswordReset)
			auth.POST("/reset-password", authHandler.ResetPassword)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Public configuration
		v1.GET("/config/trial", authHandler.GetTrialInfo)
	}
}

func SetupAuthRoutes(router *gin.Engine, authHandler *handler.AuthHandler, chatHandler *handler.ChatHandler, userHandler *handler.UserHandler, voiceHandler *handler.VoiceHandler, configHandler *handler.ConfigHandler) {
	v1 := router.Group("/api/v1")
	{
		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// Auth routes
			protected.POST("/auth/logout", authHandler.Logout)

			// User routes
			user := protected.Group("/user")
			{
				user.GET("/profile", userHandler.GetProfile)
				user.PUT("/profile", userHandler.UpdateProfile)
				user.DELETE("/profile", userHandler.DeleteProfile)
				user.GET("/chat-sessions", userHandler.GetChatSessions)
				user.POST("/chat-sessions/archive", userHandler.ArchiveChatSession)
				user.POST("/microphone-permission", voiceHandler.SaveMicrophonePermission)
			}

			// Chat routes
			chat := protected.Group("/chat")
			{
				chat.POST("/", chatHandler.Chat)
				chat.POST("/history", chatHandler.GetHistory)
				chat.POST("/search", chatHandler.SearchMessages)
				chat.POST("/sync", chatHandler.SyncMessages)
				chat.POST("/pull", chatHandler.PullMessages)
			}

			// Payment routes
			payment := protected.Group("/payment")
			{
				payment.POST("/initiate", authHandler.InitiatePayment)
			}

			// Voice routes
			voice := protected.Group("/voice")
			{
				voice.POST("/transcribe", voiceHandler.TranscribeVoice)
			}

			// Config routes (protected — requires auth)
			protected.GET("/config/sync", configHandler.GetSyncConfig)
		}
	}
}
