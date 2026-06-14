package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"mentalchat/internal/config"

	"github.com/rs/zerolog/log"
)

// --- ChadGPT API request/response structures ---
// ChadGPT (ask.chadgpt.ru) expects a simple JSON body:
//
//	{"message": "...", "api_key": "chad-..."}
//
// and returns:
//
//	{"is_success": true, "response": "...", ...}
type chadRequest struct {
	Message string `json:"message"`
	APIKey  string `json:"api_key"`
}

type chadResponse struct {
	IsSuccess       bool   `json:"is_success"`
	Response        string `json:"response"`
	Content         string `json:"content"`
	Text            string `json:"text"`
	Message         string `json:"message"`
	UsedSparksCount int    `json:"used_sparks_count"`
	UsedTokensCount int    `json:"used_tokens_count"`
	ErrorCode       string `json:"error_code,omitempty"`
	ErrorMessage    string `json:"error_message,omitempty"`
}

type AIService struct {
	cfg             *config.AIConfig
	httpClient      *http.Client
	yandexSpeechKit *YandexSpeechKitService
	promptService   *PromptService
}

func NewAIService(cfg *config.AIConfig, promptSvc *PromptService) *AIService {
	aiService := &AIService{
		cfg:           cfg,
		promptService: promptSvc,
		httpClient: &http.Client{
			Timeout: 180 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 30 * time.Second,
				DisableCompression:  true,
			},
		},
	}

	// Инициализируем Yandex SpeechKit если включен
	if cfg.YandexSpeechKit.Enabled {
		aiService.yandexSpeechKit = NewYandexSpeechKitService(&cfg.YandexSpeechKit)

		// Запускаем авто-обновление токена
		go aiService.yandexSpeechKit.StartAutoTokenRefresh(context.Background())

		log.Info().Msg("Yandex SpeechKit initialized successfully")
	}

	return aiService
}

// CheckContext sends a lightweight prompt to AI asking whether the user message
// is relevant to the service themes (psychologist, tarot, sexologist, fortune_teller).
// Returns true if the AI responds with RELEVANT, false if IRRELEVANT or on error.
func (s *AIService) CheckContext(userMessage string) bool {
	filterPrompt := s.promptService.BuildContextFilterPrompt(userMessage)

	model := s.cfg.Models.Free // lightweight model for filtering
	if model == "" {
		model = "gpt-4o-mini"
	}

	resp, err := s.callChadGPT(filterPrompt, model)
	if err != nil {
		log.Warn().Err(err).Msg("Context filter call failed, allowing message by default")
		return true // allow on error to avoid blocking users
	}

	isRelevant := IsRelevantResponse(resp)

	log.Debug().
		Str("filter_response", resp).
		Bool("is_relevant", isRelevant).
		Int("message_len", len(userMessage)).
		Msg("Context filter check completed")

	return isRelevant
}

// GetAIResponseWithContext sends the specialist prompt (from specialist.md) with
// pre-built conversation context to the AI model.
func (s *AIService) GetAIResponseWithContext(prompt, chatType, tier, contextHistory string) (string, error) {
	// Build the specialist prompt using the template
	contextSummary := s.promptService.BuildChatContextSummary(contextHistory)
	fullPrompt := s.promptService.BuildSpecialistPrompt(chatType, contextSummary, prompt)

	model := s.selectModel(tier)
	return s.callChadGPT(fullPrompt, model)
}

// GetAIResponse uses the specialist prompt without conversation context.
func (s *AIService) GetAIResponse(prompt, chatType, tier string) (string, error) {
	contextSummary := s.promptService.BuildChatContextSummary("")
	fullPrompt := s.promptService.BuildSpecialistPrompt(chatType, contextSummary, prompt)

	model := s.selectModel(tier)
	return s.callChadGPT(fullPrompt, model)
}

// selectModel returns the model name based on subscription tier.
func (s *AIService) selectModel(tier string) string {
	switch tier {
	case "pro":
		return s.cfg.Models.Pro
	case "ultra":
		return s.cfg.Models.Ultra
	default:
		return s.cfg.Models.Free
	}
}

// callChadGPT performs a real HTTP request to ChadGPT API.
// URL format: {chad_api_url}/{model}  (e.g. https://ask.chadgpt.ru/api/public/gpt-4o-mini)
// Body:   {"message": "...", "api_key": "..."}
// API key is sent in the request body (not in headers).
func (s *AIService) callChadGPT(prompt, model string) (string, error) {
	if s.cfg.ChadAPIKey == "" {
		log.Warn().Msg("No ChadGPT API key configured, using mock response")
		return s.getMockResponse(prompt), nil
	}

	// Build URL: base + / + model
	url := fmt.Sprintf("%s/%s", strings.TrimRight(s.cfg.ChadAPIURL, "/"), model)

	reqBody := chadRequest{
		Message: prompt,
		APIKey:  s.cfg.ChadAPIKey,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Err(err).Msg("Failed to marshal ChadGPT request")
		return s.getMockResponse(prompt), nil
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Err(err).Msg("Failed to create ChadGPT request")
		return s.getMockResponse(prompt), nil
	}

	req.Header.Set("Content-Type", "application/json")

	log.Debug().
		Str("url", s.cfg.ChadAPIURL).
		Int("prompt_len", len(prompt)).
		Msg("Calling ChadGPT API")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Err(err).Str("url", s.cfg.ChadAPIURL).Msg("ChadGPT API request failed")
		return s.getMockResponse(prompt), nil
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Err(err).Msg("Failed to read ChadGPT response body")
		return s.getMockResponse(prompt), nil
	}

	log.Debug().
		Int("status", resp.StatusCode).
		Int("body_len", len(respBytes)).
		Msg("ChadGPT API response received")

	if resp.StatusCode != http.StatusOK {
		log.Warn().
			Int("status", resp.StatusCode).
			Str("body", string(respBytes)).
			Msg("ChadGPT API returned non-200 status, using mock response")
		return s.getMockResponse(prompt), nil
	}

	var chadResp chadResponse
	if err := json.Unmarshal(respBytes, &chadResp); err != nil {
		log.Err(err).Str("raw", string(respBytes)).Msg("Failed to unmarshal ChadGPT response")
		return s.getMockResponse(prompt), nil
	}

	if !chadResp.IsSuccess {
		log.Warn().
			Str("error_code", chadResp.ErrorCode).
			Str("error_message", chadResp.ErrorMessage).
			Msg("ChadGPT API returned error, using mock response")
		return s.getMockResponse(prompt), nil
	}

	// Extract response text (try multiple fields in priority order)
	responseText := chadResp.Response
	if responseText == "" {
		responseText = chadResp.Content
	}
	if responseText == "" {
		responseText = chadResp.Text
	}
	if responseText == "" {
		responseText = chadResp.Message
	}

	// Clean the response: strip echoed prompt if present
	cleanedResponse := s.cleanAIResponse(responseText, prompt)

	if strings.TrimSpace(cleanedResponse) == "" || len(cleanedResponse) < 10 {
		log.Warn().
			Int("response_len", len(cleanedResponse)).
			Msg("ChadGPT response too short or empty, using mock response")
		return s.getMockResponse(prompt), nil
	}

	// Remove invalid UTF-8 sequences
	cleanedResponse = strings.ToValidUTF8(cleanedResponse, "")

	log.Info().
		Int("response_len", len(cleanedResponse)).
		Int("sparks_used", chadResp.UsedSparksCount).
		Int("tokens_used", chadResp.UsedTokensCount).
		Msg("ChadGPT API success")

	return cleanedResponse, nil
}

// cleanAIResponse removes the echoed prompt from the beginning of the API response.
func (s *AIService) cleanAIResponse(response, prompt string) string {
	response = strings.TrimSpace(response)
	prompt = strings.TrimSpace(prompt)

	// If response starts with the prompt, remove that prefix
	if strings.HasPrefix(response, prompt) {
		remaining := strings.TrimPrefix(response, prompt)
		remaining = strings.TrimSpace(remaining)
		if remaining != "" {
			return remaining
		}
	}

	// Also check for partial prompt echo (first 100 chars)
	if len(prompt) > 100 {
		promptPrefix := prompt[:100]
		if strings.HasPrefix(response, promptPrefix) {
			remaining := strings.TrimPrefix(response, promptPrefix)
			remaining = strings.TrimSpace(remaining)
			if remaining != "" {
				return remaining
			}
		}
	}

	return response
}

// getMockResponse returns a fallback message when the AI API is unavailable.
// This prevents the user from seeing an error.
func (s *AIService) getMockResponse(prompt string) string {
	now := time.Now().Format("15:04")

	// Provide a helpful response based on the prompt content
	return fmt.Sprintf(
		"Спасибо за ваше сообщение! 😊\n\n"+
			"Я — виртуальный ассистент MentalChat, и я здесь, чтобы поддержать вас.\n\n"+
			"Сейчас AI-сервис временно недоступен, но я всё равно слышу вас. "+
			"Ваше сообщение сохранено, и как только соединение восстановится, я смогу дать более развёрнутый ответ.\n\n"+
			"Пока вы ждёте, можете попробовать:\n"+
			"- Задать другой вопрос\n"+
			"- Переключить тип чата (психолог, коуч, друг)\n"+
			"- Воспользоваться голосовым вводом\n\n"+
			"_Отправлено в %s_", now)
}

// TranscribeVoice delegates to Yandex SpeechKit service (if enabled). Defaults to Russian.
func (s *AIService) TranscribeVoice(audioData []byte) (string, error) {
	if s.yandexSpeechKit == nil {
		return "", fmt.Errorf("voice transcription is not available: Yandex SpeechKit not configured")
	}
	return s.yandexSpeechKit.Transcribe(audioData, "ru-RU")
}
