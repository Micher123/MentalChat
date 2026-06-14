package service

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed specialist.md
var specialistTemplate string

//go:embed context_filter.md
var contextFilterTemplate string

// PromptService fills the specialist.md template with dynamic values.
type PromptService struct{}

func NewPromptService() *PromptService {
	return &PromptService{}
}

// roleDisplayNames maps internal chat_type to Russian role names for the prompt.
var roleDisplayNames = map[string]string{
	"psychologist":   "психолог",
	"tarot":          "таролог",
	"sexologist":     "сексолог",
	"fortune_teller": "гадалка",
}

// BuildContextFilterPrompt fills {user_message} in the context_filter.md template.
// The template asks AI to respond with exactly one word: RELEVANT or IRRELEVANT.
func (ps *PromptService) BuildContextFilterPrompt(userMessage string) string {
	return strings.Replace(contextFilterTemplate, "{user_message}", userMessage, 1)
}

// IsRelevantResponse parses the AI response from the context filter prompt
// and returns true if the user message is considered relevant to the service.
// Handles common AI output variations (mixed case, extra whitespace, punctuation).
func IsRelevantResponse(filterResponse string) bool {
	resp := strings.TrimSpace(strings.ToUpper(filterResponse))
	// Remove trailing punctuation like dots, commas, etc.
	resp = strings.TrimRight(resp, ".,!;:。，！；：")

	// Check IRRELEVANT first because it CONTAINS the substring RELEVANT.
	if strings.Contains(resp, "IRRELEVANT") {
		return false
	}
	if strings.Contains(resp, "RELEVANT") {
		return true
	}

	// If AI didn't follow the format, default to true (allow) to avoid blocking users.
	// But log this so we can tune the prompt.
	return true
}

// BuildSpecialistPrompt fills {роль специалиста}, {краткий контекст чата},
// {текущее сообщение пользователя} in the specialist.md template.
func (ps *PromptService) BuildSpecialistPrompt(chatType, contextSummary, userMessage string) string {
	role := roleDisplayNames[chatType]
	if role == "" {
		role = "психолог"
	}

	prompt := strings.Replace(specialistTemplate, "{роль специалиста}", role, 1)
	prompt = strings.Replace(prompt, "{краткий контекст чата}", contextSummary, 1)
	prompt = strings.Replace(prompt, "{текущее сообщение пользователя}", userMessage, 1)

	return prompt
}

// BuildChatContextSummary creates a compact human-readable summary for {краткий контекст чата}.
// If the AI context from ContextBuilder is empty, returns a default message.
func (ps *PromptService) BuildChatContextSummary(aiContext string) string {
	aiContext = strings.TrimSpace(aiContext)
	if aiContext == "" {
		return "Мы только начинаем общение. Предыдущей переписки нет."
	}
	// Limit to ~600 chars to keep prompt size reasonable
	if len(aiContext) > 600 {
		aiContext = aiContext[:600]
		if lastNewline := strings.LastIndex(aiContext, "\n"); lastNewline > 400 {
			aiContext = aiContext[:lastNewline]
		}
		aiContext += "\n[дальше опущено для краткости]"
	}
	return fmt.Sprintf("Вот краткое содержание нашего диалога:\n%s", aiContext)
}
