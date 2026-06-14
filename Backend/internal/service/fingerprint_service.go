package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mentalchat/internal/config"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type FingerprintService struct {
	cfg *config.SecurityConfig
}

type FingerprintData struct {
	// Browser info
	UserAgent     string `json:"ua"`
	AcceptLanguage string `json:"al"`
	AcceptEncoding string `json:"ae"`
	
	// Screen info
	ScreenWidth  int `json:"sw"`
	ScreenHeight int `json:"sh"`
	ColorDepth   int `json:"cd"`
	PixelRatio   int `json:"pr"`
	
	// Timezone
	Timezone     string `json:"tz"`
	TimezoneOffset int  `json:"to"`
	
	// Platform
	Platform      string `json:"pf"`
	HardwareConcurrency int `json:"hc"`
	DeviceMemory  int    `json:"dm"`
	TouchPoints   int    `json:"tp"`
	
	// Canvas fingerprint (hash)
	CanvasHash    string `json:"ch"`
	
	// WebGL fingerprint (hash)
	WebGLHash     string `json:"wh"`
	
	// Audio fingerprint (hash)
	AudioHash     string `json:"ah"`
	
	// Fonts (hash of installed fonts list)
	FontsHash     string `json:"fh"`
	
	// Plugins (hash of plugins list)
	PluginsHash   string `json:"ph"`
	
	// WebRTC (IP addresses)
	LocalIPs      []string `json:"li"`
	
	// Additional browser features
	DoNotTrack    string `json:"dnt"`
	CookieEnabled bool   `json:"ce"`
	Language      string `json:"lg"`
}

func NewFingerprintService(cfg *config.SecurityConfig) *FingerprintService {
	return &FingerprintService{cfg: cfg}
}

// GenerateFingerprint создает уникальный отпечаток из данных
func (s *FingerprintService) GenerateFingerprint(data *FingerprintData) string {
	// Собираем все данные в одну строку
	components := []string{
		data.UserAgent,
		data.AcceptLanguage,
		data.AcceptEncoding,
		fmt.Sprintf("%dx%d", data.ScreenWidth, data.ScreenHeight),
		fmt.Sprintf("%d-%d-%d", data.ColorDepth, data.PixelRatio, data.DeviceMemory),
		data.Timezone,
		fmt.Sprintf("%d-%d-%d", data.TimezoneOffset, data.HardwareConcurrency, data.TouchPoints),
		data.Platform,
		data.CanvasHash,
		data.WebGLHash,
		data.AudioHash,
		data.FontsHash,
		data.PluginsHash,
		strings.Join(data.LocalIPs, ","),
		data.DoNotTrack,
		fmt.Sprintf("%v", data.CookieEnabled),
		data.Language,
	}

	// Объединяем компоненты
	rawData := strings.Join(components, "|")

	// Создаем SHA-256 хеш
	hash := sha256.Sum256([]byte(rawData))

	// Кодируем в base64
	fingerprint := base64.StdEncoding.EncodeToString(hash[:])

	log.Debug().
		Str("fingerprint", fingerprint).
		Int("components", len(components)).
		Msg("Generated fingerprint")

	return fingerprint
}

// ValidateFingerprint проверяет валидность fingerprint данных
func (s *FingerprintService) ValidateFingerprint(data *FingerprintData) error {
	if data.UserAgent == "" {
		return fmt.Errorf("user agent is required")
	}

	if data.ScreenWidth <= 0 || data.ScreenHeight <= 0 {
		return fmt.Errorf("invalid screen dimensions")
	}

	if data.Timezone == "" {
		return fmt.Errorf("timezone is required")
	}

	return nil
}

// ExtractClientInfo извлекает информацию о клиенте из HTTP запроса
func (s *FingerprintService) ExtractClientInfo(r *http.Request, fpData *FingerprintData) {
	// User Agent
	if fpData.UserAgent == "" {
		fpData.UserAgent = r.UserAgent()
	}

	// Accept-Language
	if fpData.AcceptLanguage == "" {
		fpData.AcceptLanguage = r.Header.Get("Accept-Language")
	}

	// Accept-Encoding
	if fpData.AcceptEncoding == "" {
		fpData.AcceptEncoding = r.Header.Get("Accept-Encoding")
	}

	// IP address
	clientIP := s.getClientIP(r)
	log.Debug().Str("client_ip", clientIP).Msg("Extracted client IP")

	// Language
	if fpData.Language == "" {
		if langs := r.Header.Values("Accept-Language"); len(langs) > 0 {
			parts := strings.Split(langs[0], ",")
			if len(parts) > 0 {
				fpData.Language = strings.Split(parts[0], ";")[0]
			}
		}
	}
}

// getClientIP получает реальный IP клиента
func (s *FingerprintService) getClientIP(r *http.Request) string {
	// Проверка X-Forwarded-For (за прокси)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Проверка X-Real-IP
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Проверка CF-Connecting-IP (Cloudflare)
	cfip := r.Header.Get("CF-Connecting-IP")
	if cfip != "" {
		return cfip
	}

	// Fallback на RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// CalculateSimilarity вычисляет схожесть двух fingerprint
func (s *FingerprintService) CalculateSimilarity(fp1, fp2 *FingerprintData) float64 {
	matchingComponents := 0
	totalComponents := 0

	// User Agent (exact match)
	if fp1.UserAgent == fp2.UserAgent {
		matchingComponents++
	}
	totalComponents++

	// Accept Language (partial match)
	if strings.HasPrefix(fp1.AcceptLanguage, fp2.AcceptLanguage[:min(5, len(fp2.AcceptLanguage))]) {
		matchingComponents++
	}
	totalComponents++

	// Screen dimensions (exact match)
	if fp1.ScreenWidth == fp2.ScreenWidth && fp1.ScreenHeight == fp2.ScreenHeight {
		matchingComponents++
	}
	totalComponents++

	// Timezone (exact match)
	if fp1.Timezone == fp2.Timezone {
		matchingComponents++
	}
	totalComponents++

	// Platform (exact match)
	if fp1.Platform == fp2.Platform {
		matchingComponents++
	}
	totalComponents++

	// Canvas hash (exact match - most reliable)
	if fp1.CanvasHash == fp2.CanvasHash && fp1.CanvasHash != "" {
		matchingComponents++
	}
	totalComponents++

	// WebGL hash (exact match - very reliable)
	if fp1.WebGLHash == fp2.WebGLHash && fp1.WebGLHash != "" {
		matchingComponents++
	}
	totalComponents++

	// Fonts hash (exact match)
	if fp1.FontsHash == fp2.FontsHash && fp1.FontsHash != "" {
		matchingComponents++
	}
	totalComponents++

	similarity := float64(matchingComponents) / float64(totalComponents)
	
	log.Debug().
		Float64("similarity", similarity).
		Int("matching", matchingComponents).
		Int("total", totalComponents).
		Msg("Calculated fingerprint similarity")

	return similarity
}

// IsSuspiciousFingerprint проверяет, является ли fingerprint подозрительным
func (s *FingerprintService) IsSuspiciousFingerprint(data *FingerprintData) bool {
	// Проверка на пустые данные
	if data.UserAgent == "" {
		log.Warn().Msg("Suspicious fingerprint: empty user agent")
		return true
	}

	// Проверка на ботов/скрипты
	suspiciousAgents := []string{
		"bot", "crawler", "spider", "scraper",
		"curl", "wget", "python", "java",
	}

	userAgentLower := strings.ToLower(data.UserAgent)
	for _, suspicious := range suspiciousAgents {
		if strings.Contains(userAgentLower, suspicious) {
			log.Warn().Str("user_agent", data.UserAgent).Msg("Suspicious fingerprint: bot detected")
			return true
		}
	}

	// Проверка на нереальные размеры экрана
	if data.ScreenWidth < 320 || data.ScreenWidth > 5120 ||
		data.ScreenHeight < 200 || data.ScreenHeight > 4320 {
		log.Warn().Int("width", data.ScreenWidth).Int("height", data.ScreenHeight).
			Msg("Suspicious fingerprint: invalid screen dimensions")
		return true
	}

	// Проверка на нереальное количество ядер CPU
	if data.HardwareConcurrency < 1 || data.HardwareConcurrency > 128 {
		log.Warn().Int("cores", data.HardwareConcurrency).
			Msg("Suspicious fingerprint: invalid CPU cores")
		return true
	}

	// Проверка на отключенные cookies (подозрительно для обычного пользователя)
	if !data.CookieEnabled {
		log.Warn().Msg("Suspicious fingerprint: cookies disabled")
		return true
	}

	return false
}

// GenerateFingerprintID создает короткий ID для fingerprint
func (s *FingerprintService) GenerateFingerprintID(fingerprint string) string {
	// Берем первые 16 символов base64 хеша
	if len(fingerprint) >= 16 {
		return fingerprint[:16]
	}
	return fingerprint
}

// GetFingerprintExpiry возвращает время жизни fingerprint
func (s *FingerprintService) GetFingerprintExpiry() time.Duration {
	// Fingerprint действует 30 дней
	return 30 * 24 * time.Hour
}

// CreateFingerprintRecord создает запись fingerprint для хранения в БД
func (s *FingerprintService) CreateFingerprintRecord(fingerprint string, userID uint, fpData *FingerprintData) ([]byte, error) {
	record := map[string]interface{}{
		"fingerprint":     fingerprint,
		"user_id":         userID,
		"created_at":      time.Now().Unix(),
		"user_agent":      fpData.UserAgent,
		"platform":        fpData.Platform,
		"screen":          fmt.Sprintf("%dx%d", fpData.ScreenWidth, fpData.ScreenHeight),
		"canvas_hash":     fpData.CanvasHash,
		"webgl_hash":      fpData.WebGLHash,
		"fonts_hash":      fpData.FontsHash,
		"ip_hash":         s.hashIP(s.getClientIP(&http.Request{})),
	}

	return json.Marshal(record)
}

// hashIP создает хеш IP адреса для безопасного хранения
func (s *FingerprintService) hashIP(ip string) string {
	hash := sha256.Sum256([]byte(ip))
	return base64.StdEncoding.EncodeToString(hash[:8])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
