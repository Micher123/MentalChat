package service

import (
	"context"
	"mentalchat/internal/config"
	"mentalchat/internal/model"
	"time"

	"github.com/rs/zerolog/log"
)

type AIService struct {
	cfg             *config.AIConfig
	yandexSpeechKit *YandexSpeechKitService
}

func NewAIService(cfg *config.AIConfig) *AIService {
	aiService := &AIService{cfg: cfg}

	// Инициализируем Yandex SpeechKit если включен
	if cfg.YandexSpeechKit.Enabled {
		aiService.yandexSpeechKit = NewYandexSpeechKitService(&cfg.YandexSpeechKit)

		// Запускаем авто-обновление токена
		go aiService.yandexSpeechKit.StartAutoTokenRefresh(context.Background())

		log.Info().Msg("Yandex SpeechKit initialized successfully")
	}

	return aiService
}

// GetAIResponseWithContext sends prompt with pre-built conversation context to AI.
func (s *AIService) GetAIResponseWithContext(prompt, chatType, tier, contextHistory string) (string, error) {
	model := s.cfg.Models.Free
	switch tier {
	case "pro":
		model = s.cfg.Models.Pro
	case "ultra":
		model = s.cfg.Models.Ultra
	}

	systemPrompt := s.getSystemPrompt(chatType)
	userContext := s.getUserContext(tier)

	fullPrompt := systemPrompt + "\n\n" + userContext + "\n\n" +
		"История диалога:\n" + contextHistory + "\n\n" +
		"User: " + prompt + "\nAI:"

	response, err := s.callAI_API(fullPrompt, model)
	if err != nil {
		return "", err
	}

	refinedResponse, err := s.refineResponse(response, chatType)
	if err != nil {
		return "", err
	}

	return refinedResponse, nil
}

func (s *AIService) GetAIResponse(prompt, chatType, tier string) (string, error) {
	// Determine the model based on tier
	model := s.cfg.Models.Free
	switch tier {
	case "pro":
		model = s.cfg.Models.Pro
	case "ultra":
		model = s.cfg.Models.Ultra
	}

	// Construct the full prompt with context
	fullPrompt := s.buildPrompt(prompt, chatType, tier)

	// Call AI API (placeholder implementation)
	response, err := s.callAI_API(fullPrompt, model)
	if err != nil {
		return "", err
	}

	// Refine the response for better quality
	refinedResponse, err := s.refineResponse(response, chatType)
	if err != nil {
		return "", err
	}

	return refinedResponse, nil
}

func (s *AIService) buildPrompt(prompt, chatType, tier string) string {
	// Base system prompt based on chat type
	systemPrompt := s.getSystemPrompt(chatType)

	// Add user context based on tier
	userContext := s.getUserContext(tier)

	return systemPrompt + "\n\n" + userContext + "\n\nUser: " + prompt + "\nAI:"
}

func (s *AIService) getSystemPrompt(chatType string) string {
	prompts := map[string]string{
		"psychologist":   `Ты профессиональный психолог, который помогает женщинам разбираться в их эмоциях, находить баланс и улучшать качество жизни. Твой стиль общения теплый, поддерживающий и эмпатичный.`,
		"tarot":          `Ты опытный таролог, который помогает женщинам найти ответы на важные вопросы через карты Таро. Твой стиль - мудрый, вдохновляющий и ориентированный на духовный рост.`,
		"sexologist":     `Ты квалифицированный сексолог, который помогает женщинам понять себя, свои желания и улучшить свою сексуальную жизнь. Твой стиль - открытый, непредвзятый и поддерживающий.`,
		"fortune_teller": `Ты чувственная гадалка, которая помогает женщинам увидеть путь вперед через интуицию и метафизические практики. Твой стиль - вдохновляющий и ориентированный на позитивные изменения.`,
	}

	if prompt, ok := prompts[chatType]; ok {
		return prompt
	}
	return prompts["psychologist"]
}

func (s *AIService) getUserContext(tier string) string {
	switch tier {
	case "pro":
		return "Пользователь использует PRO-тариф с улучшенной моделью ИИ. Ему доступны более глубокие аналитические возможности и персонализированные советы."
	case "ultra":
		return "Пользователь использует ULTRA-тариф с максимальной моделью ИИ. Ему доступны самые продвинутые возможности, включая анализ сложных ситуаций и долгосрочное планирование."
	default:
		return "Пользователь использует бесплатный тариф. Ему доступны основные возможности ИИ для поддержки в повседневных вопросах."
	}
}

func (s *AIService) callAI_API(prompt, model string) (string, error) {
	// Placeholder for actual AI API call
	// In production, this would call Yandex Chad API or another AI service

	// Simulate AI response
	response := "Я понимаю, что вы хотите обсудить. Давайте разберемся вместе. " +
		"Важно помнить, что ваши чувства и переживания важны и заслуживают внимания. " +
		"Попробуйте сосредоточиться на своем дыхании и почувствовать себя в безопасности."

	return response, nil
}

func (s *AIService) refineResponse(response, chatType string) (string, error) {
	// Refine response to make it more natural and emotionally rich
	// This could involve calling the AI API again with specific instructions

	// Add emotional coloring and warmth
	warmthPhrases := map[string][]string{
		"psychologist": []string{
			"Милая,",
			"Дорогая,",
			"Золотце,",
			"Milочка,",
		},
		"tarot": []string{
			"Духовная душа,",
			"Светлая душа,",
			"Ты светлая натура,",
			"Душа моя,",
		},
		"sexologist": []string{
			"Моя дорогая,",
			"Дорогая моя,",
			"Ты прекрасна,",
			"Душа моя,",
		},
		"fortune_teller": []string{
			"Моя дорогая,",
			"Ты светлая натура,",
			"Душа моя,",
			"Милая моя,",
		},
	}

	phrases := warmthPhrases[chatType]
	if len(phrases) > 0 {
		// Select random phrase (simplified - in production use proper random)
		response = phrases[0] + " " + response
	}

	// Add closing reassurance
	reassurance := "\n\nПомни, ты прекрасна такой, какая есть. Всё будет хорошо! 🌸"
	response += reassurance

	return response, nil
}

func (s *AIService) TranscribeVoice(voiceData []byte) (string, error) {
	// Если Yandex SpeechKit включен, используем его
	if s.yandexSpeechKit != nil {
		transcript, err := s.yandexSpeechKit.Transcribe(voiceData, "ru-RU")
		if err != nil {
			log.Err(err).Msg("Yandex SpeechKit transcription failed, falling back to default")
			// Fallback на заглушку
			return s.fallbackTranscribe(voiceData)
		}
		return transcript, nil
	}

	// Fallback на заглушку если Yandex SpeechKit не включен
	return s.fallbackTranscribe(voiceData)
}

func (s *AIService) fallbackTranscribe(voiceData []byte) (string, error) {
	// Заглушка для тестирования
	// В продакшене здесь должна быть интеграция с другим STT сервисом
	log.Warn().Msg("Using fallback transcription (placeholder)")
	return "Это тестовая транскрипция. Настройте Yandex SpeechKit для реальной работы.", nil
}

func (s *AIService) ProcessVoiceMessage(userID uint, voiceFilePath string) (*model.VoiceMessage, error) {
	// Read voice file
	// Transcribe voice to text
	// Get AI response
	// Store voice message and transcript

	vm := &model.VoiceMessage{
		UserID:     userID,
		FilePath:   voiceFilePath,
		FileName:   "voice_" + time.Now().Format("20060102_150405") + ".ogg",
		FileSize:   1024,
		Duration:   15.5,
		Transcript: "Транскрипция голосового сообщения",
	}

	return vm, nil
}
