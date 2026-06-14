package main

import (
	"mentalchat/internal/config"
	"mentalchat/internal/handler"
	"mentalchat/internal/middleware"
	"mentalchat/internal/routes"
	"mentalchat/internal/service"
	"mentalchat/internal/storage"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize configuration
	cfg := config.LoadConfig()

	// Initialize logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	if !cfg.Server.Debug {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		gin.SetMode(gin.DebugMode)
	}

	// Initialize database
	db, err := storage.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Initialize storage
	storage := storage.NewStorage(db)

	// Initialize services
	promptService := service.NewPromptService()
	aiService := service.NewAIService(&cfg.AI, promptService)
	emailService := service.NewEmailService(&cfg.Email)
	fingerprintService := service.NewFingerprintService(&cfg.Security)
	jwtService := service.NewJWTService()
	authService := service.NewAuthService(storage, jwtService, emailService, fingerprintService, &cfg.App)
	paymentService := service.NewPaymentService(cfg.Payment, storage)
	userService := service.NewUserService(storage, emailService, &cfg.App)

	// Initialize new chat services
	chatIndexService := service.NewChatIndexService(storage)
	contextBuilder := service.NewContextBuilder(storage, service.DefaultContextConfig())
	messageQueueService := service.NewMessageQueueService(storage, chatIndexService, contextBuilder, aiService)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, userService, paymentService, storage, emailService)
	chatHandler := handler.NewChatHandler(aiService, chatIndexService, contextBuilder, storage, messageQueueService)
	userHandler := handler.NewUserHandler(userService, storage)
	voiceHandler := handler.NewVoiceHandler(aiService, storage)
	configHandler := handler.NewConfigHandler(storage)
	fileHandler := handler.NewFileHandler(aiService, chatIndexService, contextBuilder, storage, messageQueueService)

	// Initialize router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.RequestLogger())
	router.Use(middleware.RateLimiter(cfg.Security.RateLimitRequests, cfg.Security.RateLimitWindowSeconds))
	if cfg.Security.DdosProtectionEnabled {
		router.Use(middleware.DDoSProtection(cfg.Security.MaxConcurrentRequestsPerIP))
	}

	// Setup routes
	routes.SetupPublicRoutes(router, authHandler)
	routes.SetupAuthRoutes(router, authHandler, chatHandler, userHandler, voiceHandler, configHandler, fileHandler)

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Info().Str("address", addr).Msg("Starting server")
	if err := router.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
