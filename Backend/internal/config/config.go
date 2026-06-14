package config

import (
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	AI       AIConfig       `mapstructure:"ai"`
	Payment  PaymentConfig  `mapstructure:"payment"`
	Email    EmailConfig    `mapstructure:"email"`
	Security SecurityConfig `mapstructure:"security"`
	Storage  StorageConfig  `mapstructure:"storage"`
	App      AppConfig      `mapstructure:"app"`
	Sync     SyncConfig     `mapstructure:"sync"`
}

type ServerConfig struct {
	Host  string `mapstructure:"host"`
	Port  string `mapstructure:"port"`
	Debug bool   `mapstructure:"debug"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type AIConfig struct {
	ChadAPIURL      string                `mapstructure:"chad_api_url"`
	ChadAPIKey      string                `mapstructure:"chad_api_key"`
	Models          AIModelConfig         `mapstructure:"models"`
	YandexSpeechKit YandexSpeechKitConfig `mapstructure:"yandex_speechkit"`
}

type YandexSpeechKitConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	FolderID    string `mapstructure:"folder_id"`
	IAMToken    string `mapstructure:"iam_token"`
	ServiceKey  string `mapstructure:"service_key"`
	APIEndpoint string `mapstructure:"api_endpoint"`
}

type AIModelConfig struct {
	Free  string `mapstructure:"free"`
	Pro   string `mapstructure:"pro"`
	Ultra string `mapstructure:"ultra"`
}

type PaymentConfig struct {
	Provider string `mapstructure:"provider"`
	YooMoney YooMoneyConfig
	Prices   PriceConfig
}

type YooMoneyConfig struct {
	ShopID string `mapstructure:"shop_id"`
	Secret string `mapstructure:"secret"`
	Scid   string `mapstructure:"scid"`
}

type PriceConfig struct {
	ProMonthly     int `mapstructure:"pro_monthly"`
	ProYearly      int `mapstructure:"pro_yearly"`
	UltraMonthly   int `mapstructure:"ultra_monthly"`
	UltraYearly    int `mapstructure:"ultra_yearly"`
	ProTrialDays   int `mapstructure:"pro_trial_days"`
	UltraTrialDays int `mapstructure:"ultra_trial_days"`
}

type EmailConfig struct {
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     string `mapstructure:"smtp_port"`
	SMTPUser     string `mapstructure:"smtp_user"`
	SMTPPassword string `mapstructure:"smtp_password"`
	FromEmail    string `mapstructure:"from_email"`
	FromName     string `mapstructure:"from_name"`
}

type SecurityConfig struct {
	JWTSecret                  string `mapstructure:"jwt_secret"`
	JWTExpirationHours         int    `mapstructure:"jwt_expiration_hours"`
	RateLimitRequests          int    `mapstructure:"rate_limit_requests"`
	RateLimitWindowSeconds     int    `mapstructure:"rate_limit_window_seconds"`
	DdosProtectionEnabled      bool   `mapstructure:"ddos_protection_enabled"`
	MaxConcurrentRequestsPerIP int    `mapstructure:"max_concurrent_requests_per_ip"`
}

type StorageConfig struct {
	VoiceStoragePath  string `mapstructure:"voice_storage_path"`
	AvatarStoragePath string `mapstructure:"avatar_storage_path"`
	MaxFileSizeMB     int    `mapstructure:"max_file_size_mb"`
}

type AppConfig struct {
	FrontendURL  string `mapstructure:"frontend_url"`
	AppName      string `mapstructure:"app_name"`
	SupportEmail string `mapstructure:"support_email"`
}

type SyncConfig struct {
	IntervalSeconds int `mapstructure:"interval_seconds"`
}

var (
	instance *Config
	once     sync.Once
)

func LoadConfig() *Config {
	once.Do(func() {
		viper.SetConfigName(".env")
		viper.SetConfigType("json")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/mentalchat/")

		// Явно читаем файл
		if err := viper.ReadInConfig(); err != nil {
			panic("Config file error: " + err.Error())
		}

		viper.SetDefault("server.host", "localhost")
		viper.SetDefault("server.port", "8080")
		viper.SetDefault("server.debug", false)

		viper.SetDefault("database.host", "localhost")
		viper.SetDefault("database.port", "5432")
		viper.SetDefault("database.name", "mentalchat")
		viper.SetDefault("database.user", "mentalchat")
		viper.SetDefault("database.password", "mentalchat")
		viper.SetDefault("database.ssl_mode", "disable")

		viper.SetDefault("ai.chad_api_url", "https://ask.chadgpt.ru/api/public")
		viper.SetDefault("ai.chad_api_key", "")
		viper.SetDefault("ai.models.free", "gpt-4o-mini")
		viper.SetDefault("ai.models.pro", "gpt-4o")
		viper.SetDefault("ai.models.ultra", "gpt-4-turbo")
		viper.SetDefault("ai.yandex_speechkit.enabled", false)
		viper.SetDefault("ai.yandex_speechkit.folder_id", "")
		viper.SetDefault("ai.yandex_speechkit.iam_token", "")
		viper.SetDefault("ai.yandex_speechkit.service_key", "")
		viper.SetDefault("ai.yandex_speechkit.api_endpoint", "https://stt.api.cloud.yandex.net")

		viper.SetDefault("payment.provider", "yoomoney")
		viper.SetDefault("payment.prices.pro_monthly", 499)
		viper.SetDefault("payment.prices.pro_yearly", 4990)
		viper.SetDefault("payment.prices.ultra_monthly", 999)
		viper.SetDefault("payment.prices.ultra_yearly", 9990)
		viper.SetDefault("payment.prices.pro_trial_days", 3)
		viper.SetDefault("payment.prices.ultra_trial_days", 1)

		viper.SetDefault("email.smtp_host", "localhost")
		viper.SetDefault("email.smtp_port", "587")
		viper.SetDefault("email.from_email", "noreply@localhost")
		viper.SetDefault("email.from_name", "MentalChat")

		viper.SetDefault("security.jwt_expiration_hours", 168)
		viper.SetDefault("security.rate_limit_requests", 100)
		viper.SetDefault("security.rate_limit_window_seconds", 60)
		viper.SetDefault("security.ddos_protection_enabled", true)
		viper.SetDefault("security.max_concurrent_requests_per_ip", 50)

		viper.SetDefault("storage.voice_storage_path", "./storage/voices")
		viper.SetDefault("storage.avatar_storage_path", "./storage/avatars")
		viper.SetDefault("storage.max_file_size_mb", 10)

		viper.SetDefault("app.frontend_url", "http://localhost:3000")
		viper.SetDefault("app.app_name", "MentalChat")
		viper.SetDefault("app.support_email", "support@localhost")

		viper.SetDefault("sync.interval_seconds", 300)

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file not found, use defaults
			} else {
				panic(err)
			}
		}

		instance = &Config{}
		if err := viper.Unmarshal(instance); err != nil {
			panic(err)
		}
	})

	return instance
}

func GetJWTSecret() string {
	return LoadConfig().Security.JWTSecret
}

func GetJWTExpirationHours() int {
	return LoadConfig().Security.JWTExpirationHours
}

func GetFrontendURL() string {
	return LoadConfig().App.FrontendURL
}

func GetAppSupportEmail() string {
	return LoadConfig().App.SupportEmail
}

func GetAIModel(tier string) string {
	cfg := LoadConfig()
	switch tier {
	case "free":
		return cfg.AI.Models.Free
	case "pro":
		return cfg.AI.Models.Pro
	case "ultra":
		return cfg.AI.Models.Ultra
	default:
		return cfg.AI.Models.Free
	}
}

func GetPaymentConfig() PaymentConfig {
	return LoadConfig().Payment
}
