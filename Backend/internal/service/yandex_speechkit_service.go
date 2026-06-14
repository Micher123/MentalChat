package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mentalchat/internal/config"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type YandexSpeechKitService struct {
	cfg        *config.YandexSpeechKitConfig
	httpClient *http.Client
	tokenExp   time.Time
}

type YandexSTTRequest struct {
	Audio           AudioSpec       `json:"audio"`
	AudioSpecs      []AudioFormat   `json:"audioSpecs"`
	RecognitionSpec RecognitionSpec `json:"recognitionSpec"`
}

type AudioSpec struct {
	AudioContent string `json:"audioContent,omitempty"`
	AudioURI     string `json:"audioUri,omitempty"`
}

type AudioFormat struct {
	SampleRateHertz int64  `json:"sampleRateHertz"`
	Encoding        string `json:"encoding"`
}

type RecognitionSpec struct {
	LanguageCode          string `json:"languageCode"`
	ProfanityFilter       bool   `json:"profanityFilter"`
	Model                 string `json:"model"`
	AudioFormatType       string `json:"audioFormatType"`
	EnableWordTimeMapping bool   `json:"enableWordTimeMapping"`
	SingleUtterance       bool   `json:"singleUtterance"`
	MaxAlternatives       int    `json:"maxAlternatives"`
}

type YandexSTTResponse struct {
	Chunks []Chunk `json:"chunks"`
}

type Chunk struct {
	Alternative Alternative `json:"alternative"`
	StartTime   string      `json:"startTime"`
	EndTime     string      `json:"endTime"`
}

type Alternative struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Words      []Word  `json:"words,omitempty"`
}

type Word struct {
	Text       string  `json:"text"`
	StartTime  string  `json:"startTime"`
	EndTime    string  `json:"endTime"`
	Confidence float64 `json:"confidence"`
}

type IAMTokenResponse struct {
	IAMToken  string `json:"iamToken"`
	ExpiresIn string `json:"expiresIn"`
}

func NewYandexSpeechKitService(cfg *config.YandexSpeechKitConfig) *YandexSpeechKitService {
	return &YandexSpeechKitService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetIAMToken получает IAM токен для аутентификации
func (s *YandexSpeechKitService) GetIAMToken() (string, error) {
	// Если токен уже есть и не истек, используем его
	if s.cfg.IAMToken != "" && time.Now().Before(s.tokenExp) {
		return s.cfg.IAMToken, nil
	}

	// Если есть service key, получаем токен через него
	if s.cfg.ServiceKey != "" {
		return s.getIAMTokenFromServiceKey()
	}

	// Если токена нет, возвращаем ошибку
	if s.cfg.IAMToken == "" {
		return "", fmt.Errorf("IAM token not configured")
	}

	return s.cfg.IAMToken, nil
}

func (s *YandexSpeechKitService) getIAMTokenFromServiceKey() (string, error) {
	url := "https://iam.api.cloud.yandex.net/iam/v1/tokens"

	requestBody := map[string]string{
		"yandexPassportOauthToken": s.cfg.ServiceKey,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var tokenResp IAMTokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Устанавливаем время истечения (обычно 1 час)
	expiresIn, err := time.ParseDuration(fmt.Sprintf("%ds", 3600))
	if err == nil {
		s.tokenExp = time.Now().Add(expiresIn)
	}

	s.cfg.IAMToken = tokenResp.IAMToken

	return tokenResp.IAMToken, nil
}

// Transcribe отправляет аудио на транскрибацию в Yandex SpeechKit
func (s *YandexSpeechKitService) Transcribe(audioData []byte, language string) (string, error) {
	if !s.cfg.Enabled {
		return "", fmt.Errorf("Yandex SpeechKit is not enabled")
	}

	// Получаем IAM токен
	iamToken, err := s.GetIAMToken()
	if err != nil {
		return "", fmt.Errorf("failed to get IAM token: %w", err)
	}

	// Кодируем аудио в base64
	audioBase64 := base64.StdEncoding.EncodeToString(audioData)

	// Формируем запрос
	requestBody := YandexSTTRequest{
		Audio: AudioSpec{
			AudioContent: audioBase64,
		},
		AudioSpecs: []AudioFormat{
			{
				SampleRateHertz: 48000,
				Encoding:        "OGGOPUS",
			},
		},
		RecognitionSpec: RecognitionSpec{
			LanguageCode:          language,
			ProfanityFilter:       false,
			Model:                 "phonecall",
			AudioFormatType:       "LONG_AUDIO_UNSPECIFIED",
			EnableWordTimeMapping: false,
			SingleUtterance:       true,
			MaxAlternatives:       1,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Создаем запрос к API
	url := fmt.Sprintf("%s/speech/v1/stt:recognize", s.cfg.APIEndpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", iamToken))
	req.Header.Set("X-Data-Logging-Enabled", "true")

	// Отправляем запрос
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Логируем ответ для отладки
	log.Debug().Str("response", string(respBody)).Msg("Yandex SpeechKit response")

	// Проверяем статус код
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Парсим ответ
	var result YandexSTTResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Собираем текст из всех чанков
	var fullText string
	for _, chunk := range result.Chunks {
		if chunk.Alternative.Text != "" {
			fullText += chunk.Alternative.Text + " "
		}
	}

	return fullText, nil
}

// TranscribeFromURL транскрибирует аудио по URL
func (s *YandexSpeechKitService) TranscribeFromURL(audioURL string, language string) (string, error) {
	if !s.cfg.Enabled {
		return "", fmt.Errorf("Yandex SpeechKit is not enabled")
	}

	iamToken, err := s.GetIAMToken()
	if err != nil {
		return "", fmt.Errorf("failed to get IAM token: %w", err)
	}

	requestBody := YandexSTTRequest{
		Audio: AudioSpec{
			AudioURI: audioURL,
		},
		AudioSpecs: []AudioFormat{
			{
				SampleRateHertz: 48000,
				Encoding:        "OGGOPUS",
			},
		},
		RecognitionSpec: RecognitionSpec{
			LanguageCode:    language,
			ProfanityFilter: false,
			Model:           "phonecall",
			SingleUtterance: true,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/speech/v1/stt:recognize", s.cfg.APIEndpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", iamToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result YandexSTTResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	var fullText string
	for _, chunk := range result.Chunks {
		if chunk.Alternative.Text != "" {
			fullText += chunk.Alternative.Text + " "
		}
	}

	return fullText, nil
}

// RefreshToken обновляет IAM токен
func (s *YandexSpeechKitService) RefreshToken() error {
	if s.cfg.ServiceKey != "" {
		_, err := s.getIAMTokenFromServiceKey()
		return err
	}
	return nil
}

// StartAutoTokenRefresh запускает автоматическое обновление токена
func (s *YandexSpeechKitService) StartAutoTokenRefresh(ctx context.Context) {
	ticker := time.NewTicker(50 * time.Minute) // Обновляем за 10 минут до истечения
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.RefreshToken(); err != nil {
				log.Err(err).Msg("Failed to refresh IAM token")
			} else {
				log.Info().Msg("IAM token refreshed successfully")
			}
		}
	}
}
