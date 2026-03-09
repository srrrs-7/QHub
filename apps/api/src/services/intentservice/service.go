// Package intentservice provides a rule-based intent classifier for consulting chat messages.
//
// It categorizes user messages into predefined intent types (improve, compare,
// create, compliance, best_practice, explain, general) using keyword matching.
// Both English and Japanese patterns are supported.
package intentservice

import (
	"strings"
)

// Intent represents a classified user intent.
type Intent struct {
	Type       string            `json:"type"`
	Confidence float64           `json:"confidence"`
	Entities   map[string]string `json:"entities,omitempty"`
}

// Intent types
const (
	IntentImprove      = "improve"       // User wants to improve a prompt
	IntentCompare      = "compare"       // User wants to compare versions
	IntentExplain      = "explain"       // User wants explanation of a prompt
	IntentCreate       = "create"        // User wants to create a new prompt
	IntentCompliance   = "compliance"    // User asks about compliance
	IntentBestPractice = "best_practice" // User asks for best practices
	IntentGeneral      = "general"       // General conversation
)

// Classify analyzes a user message and returns the most likely intent.
func Classify(message string) Intent {
	lower := strings.ToLower(message)

	// Check patterns in priority order
	if matchesAny(lower, improvePatterns) {
		return Intent{Type: IntentImprove, Confidence: 0.8}
	}
	if matchesAny(lower, comparePatterns) {
		return Intent{Type: IntentCompare, Confidence: 0.8}
	}
	if matchesAny(lower, createPatterns) {
		return Intent{Type: IntentCreate, Confidence: 0.8}
	}
	if matchesAny(lower, compliancePatterns) {
		return Intent{Type: IntentCompliance, Confidence: 0.8}
	}
	if matchesAny(lower, bestPracticePatterns) {
		return Intent{Type: IntentBestPractice, Confidence: 0.7}
	}
	if matchesAny(lower, explainPatterns) {
		return Intent{Type: IntentExplain, Confidence: 0.7}
	}

	return Intent{Type: IntentGeneral, Confidence: 0.5}
}

var (
	improvePatterns = []string{
		"improve", "better", "optimize", "enhance", "refine",
		"改善", "良く", "最適化",
	}
	comparePatterns = []string{
		"compare", "difference", "diff", "versus", "vs",
		"比較", "違い", "差分",
	}
	createPatterns = []string{
		"create", "write", "generate", "draft", "new prompt",
		"作成", "書いて", "生成",
	}
	compliancePatterns = []string{
		"compliance", "compliant", "regulation", "hipaa", "gdpr", "legal",
		"コンプライアンス", "準拠", "規制",
	}
	bestPracticePatterns = []string{
		"best practice", "recommendation", "guideline", "tip", "advice",
		"ベストプラクティス", "推奨", "ガイドライン",
	}
	explainPatterns = []string{
		"explain", "what does", "how does", "why", "understand",
		"説明", "なぜ", "どういう",
	}
)

func matchesAny(text string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}
