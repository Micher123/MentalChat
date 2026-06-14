package service

import (
	"math"
	"mentalchat/internal/model"
	"mentalchat/internal/storage"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// ContextConfig controls how AI context is built
type ContextConfig struct {
	MaxTokens       int     `json:"max_tokens"`        // Hard limit for context sent to AI
	CharsPerToken   float64 `json:"chars_per_token"`   // Approximate: ~4 for English, ~2 for Cyrillic
	RecentCount     int     `json:"recent_count"`      // Always include last N messages
	ImportantCount  int     `json:"important_count"`   // Max "important" messages to pull from history
	SummaryMaxChars int     `json:"summary_max_chars"` // Max chars for auto-summary of older messages
}

func DefaultContextConfig() ContextConfig {
	return ContextConfig{
		MaxTokens:       4096,
		CharsPerToken:   2.0, // Cyrillic is ~2 chars/token in YandexGPT
		RecentCount:     10,
		ImportantCount:  5,
		SummaryMaxChars: 800,
	}
}

// ContextBuilder builds ultra-compact context for AI models from chat history.
// Strategy:
//  1. Always include the last N messages (recency bias)
//  2. Select K most "important" older messages (keyword/key-phrase relevance)
//  3. Auto-summarize everything older that didn't make the cut
//  4. Truncate to fit within token budget
type ContextBuilder struct {
	store  *storage.Storage
	config ContextConfig
}

func NewContextBuilder(store *storage.Storage, cfg ContextConfig) *ContextBuilder {
	if cfg.MaxTokens == 0 {
		cfg = DefaultContextConfig()
	}
	return &ContextBuilder{store: store, config: cfg}
}

// BuildContext returns a compact string ready to pass to AI as conversation history.
// The returned string includes only the most relevant messages, truncated/summarized.
func (cb *ContextBuilder) BuildContext(userID uint, chatType string, currentPrompt string) (string, []model.ContextEntry, error) {
	// 1. Fetch all messages (we'll filter in-memory for flexibility)
	recent, err := cb.store.GetRecentMessagesForContext(userID, chatType, cb.config.RecentCount+50)
	if err != nil {
		log.Err(err).Msg("Failed to fetch recent messages for context")
		return "", nil, err
	}

	if len(recent) == 0 {
		return "", nil, nil
	}

	// 2. Separate "always include" recent from rest
	recentCount := cb.config.RecentCount
	if recentCount > len(recent) {
		recentCount = len(recent)
	}

	alwaysInclude := recent[len(recent)-recentCount:] // Last N
	older := recent[:len(recent)-recentCount]

	// 3. Score and pick important older messages
	important := cb.pickImportant(older, currentPrompt, cb.config.ImportantCount)

	// 4. Summarize the rest
	summary := cb.summarizeOlder(older, important)

	// 5. Build entries list
	entries := make([]model.ContextEntry, 0, len(alwaysInclude)+len(important))

	// Add summary first if exists
	if summary != "" {
		entries = append(entries, model.ContextEntry{
			Role:       "system",
			Content:    summary,
			Importance: 0.5,
		})
	}

	// Add important older messages
	for _, m := range important {
		entries = append(entries, model.ContextEntry{
			Role:       m.Role,
			Content:    cb.truncateContent(m.Content, 300),
			Hash:       m.ContentHash,
			Importance: 0.7,
		})
	}

	// Add always-include recent messages
	for _, m := range alwaysInclude {
		entries = append(entries, model.ContextEntry{
			Role:       m.Role,
			Content:    cb.truncateContent(m.Content, 500),
			Hash:       m.ContentHash,
			Importance: 0.9,
		})
	}

	// 6. Fit to token budget
	contextStr := cb.fitToBudget(entries)

	log.Info().
		Uint("user_id", userID).
		Str("chat_type", chatType).
		Int("total_messages", len(recent)).
		Int("recent_included", len(alwaysInclude)).
		Int("important_included", len(important)).
		Int("context_chars", len(contextStr)).
		Msg("Context built")

	return contextStr, entries, nil
}

// pickImportant scores older messages by keyword overlap with current prompt and picks top-k.
func (cb *ContextBuilder) pickImportant(messages []*model.Message, currentPrompt string, k int) []*model.Message {
	if len(messages) == 0 || k <= 0 {
		return nil
	}

	// Extract keywords from current prompt
	promptKeywords := extractKeywords(currentPrompt)

	type scored struct {
		msg   *model.Message
		score float64
	}

	scoredList := make([]scored, 0, len(messages))

	for _, m := range messages {
		msgKeywords := extractKeywords(m.Content)
		score := keywordOverlap(promptKeywords, msgKeywords)
		// Bonus: AI responses that contain user-style keywords are more relevant
		if m.IsFromAI && score > 0 {
			score *= 1.2
		}
		// Bonus: longer messages may carry more context
		if len(m.Content) > 200 {
			score *= 1.1
		}
		// Penalty: too old
		ageHours := time.Since(m.Timestamp).Hours()
		if ageHours > 24 {
			score *= math.Max(0.3, 1.0-(ageHours/720.0)) // Decay over 30 days
		}

		scoredList = append(scoredList, scored{msg: m, score: score})
	}

	// Sort by score descending
	sort.Slice(scoredList, func(i, j int) bool {
		return scoredList[i].score > scoredList[j].score
	})

	// Pick top-k with non-zero scores
	result := make([]*model.Message, 0, k)
	for _, s := range scoredList {
		if s.score <= 0 {
			break
		}
		result = append(result, s.msg)
		if len(result) >= k {
			break
		}
	}

	// Restore chronological order
	sort.Slice(result, func(i, j int) bool {
		return result[i].SequenceNumber < result[j].SequenceNumber
	})

	return result
}

// summarizeOlder creates a brief summary of messages that won't be individually included.
func (cb *ContextBuilder) summarizeOlder(all []*model.Message, included []*model.Message) string {
	if len(all) == 0 {
		return ""
	}

	// Build set of included hashes for fast lookup
	includedSet := make(map[string]bool, len(included))
	for _, m := range included {
		includedSet[m.ContentHash] = true
	}

	// Find excluded messages
	excluded := make([]*model.Message, 0)
	for _, m := range all {
		if !includedSet[m.ContentHash] {
			excluded = append(excluded, m)
		}
	}

	if len(excluded) == 0 {
		return ""
	}

	// Simple extractive summary: pick first sentences of first/last messages + count
	var parts []string
	parts = append(parts, "[Предыдущий диалог:")

	// First user message topic
	for _, m := range excluded {
		if !m.IsFromAI && m.Content != "" {
			first := firstSentence(m.Content, 100)
			if first != "" {
				parts = append(parts, "начало: \""+first+"\"")
				break
			}
		}
	}

	// Last exchange topic
	lastAI := ""
	lastUser := ""
	for i := len(excluded) - 1; i >= 0; i-- {
		m := excluded[i]
		if m.IsFromAI && lastAI == "" {
			lastAI = firstSentence(m.Content, 100)
		}
		if !m.IsFromAI && lastUser == "" {
			lastUser = firstSentence(m.Content, 100)
		}
		if lastAI != "" && lastUser != "" {
			break
		}
	}
	if lastUser != "" || lastAI != "" {
		parts = append(parts, "последнее: "+lastUser+" → "+lastAI)
	}

	parts = append(parts, "всего "+itoa(len(excluded))+" сообщений]")

	summary := strings.Join(parts, "; ")
	if len(summary) > cb.config.SummaryMaxChars {
		summary = summary[:cb.config.SummaryMaxChars] + "..."
	}
	return summary
}

// fitToBudget concatenates entries into a single string, respecting token limit.
// Always keeps the most recent entries if truncation is needed.
func (cb *ContextBuilder) fitToBudget(entries []model.ContextEntry) string {
	maxChars := int(float64(cb.config.MaxTokens) * cb.config.CharsPerToken)

	if len(entries) == 0 {
		return ""
	}

	var result strings.Builder
	charsUsed := 0

	// Build from oldest to newest (entries already sorted)
	for i, e := range entries {
		line := formatEntry(e)
		lineChars := len(line)

		if charsUsed+lineChars > maxChars {
			// Can't fit this + remaining. Truncate: skip middle entries, keep last ones.
			// Calculate how many entries we can fit from the end
			remaining := entries[i:]
			// Always keep at least the last 2 entries
			keepFromEnd := len(remaining)
			if keepFromEnd > 2 {
				keepFromEnd = 2
			}
			if len(remaining)-keepFromEnd > 0 {
				result.WriteString("[...опущено " + itoa(len(remaining)-keepFromEnd) + " сообщений...]\n")
				charsUsed += 50 // rough estimate
			}
			// Add the last entries
			for _, re := range remaining[len(remaining)-keepFromEnd:] {
				l := formatEntry(re)
				if charsUsed+len(l) <= maxChars {
					result.WriteString(l)
					charsUsed += len(l)
				}
			}
			break
		}

		result.WriteString(line)
		charsUsed += lineChars
	}

	return result.String()
}

func formatEntry(e model.ContextEntry) string {
	switch e.Role {
	case "user":
		return "👤: " + e.Content + "\n"
	case "ai":
		return "🤖: " + e.Content + "\n"
	case "system":
		return "📋 " + e.Content + "\n"
	default:
		return e.Content + "\n"
	}
}

func (cb *ContextBuilder) truncateContent(content string, maxChars int) string {
	if len(content) <= maxChars {
		return content
	}
	// Try to break at sentence boundary
	cut := content[:maxChars]
	lastDot := strings.LastIndexAny(cut, ".!?")
	if lastDot > maxChars/2 {
		return cut[:lastDot+1]
	}
	return cut + "..."
}

// ---- Helper functions ----

// extractKeywords returns lowercase word tokens of 3+ chars, deduplicated
func extractKeywords(text string) map[string]bool {
	words := strings.Fields(strings.ToLower(text))
	kw := make(map[string]bool, len(words))
	for _, w := range words {
		// Strip punctuation
		w = strings.Trim(w, ".,!?;:()[]{}«»\"'")
		if len(w) >= 3 {
			kw[w] = true
		}
	}
	return kw
}

// keywordOverlap returns Jaccard-like overlap ratio between two keyword sets
func keywordOverlap(a, b map[string]bool) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	intersection := 0
	for k := range a {
		if b[k] {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// firstSentence returns the first sentence of text, up to maxChars
func firstSentence(text string, maxChars int) string {
	// Find first sentence-ending punctuation
	endings := []string{". ", "! ", "? ", ".\n", "!\n", "?\n"}
	best := len(text)
	for _, e := range endings {
		idx := strings.Index(text, e)
		if idx > 0 && idx < best {
			best = idx + 1
		}
	}
	if best < len(text) && best <= maxChars {
		return strings.TrimSpace(text[:best])
	}
	if len(text) <= maxChars {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(text[:maxChars]) + "..."
}

func itoa(n int) string {
	if n < 0 {
		return "0"
	}
	digits := "0123456789"
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = digits[n%10]
		n /= 10
	}
	return string(buf[i:])
}
