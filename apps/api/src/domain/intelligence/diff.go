// Package intelligence defines domain types for prompt analysis features
// including semantic diffs and lint results.
package intelligence

// SemanticDiff represents the structured comparison between two prompt versions.
// It captures detected changes, tone shifts, and specificity differences.
type SemanticDiff struct {
	Summary     string       `json:"summary"`
	Changes     []DiffChange `json:"changes"`
	ToneShift   string       `json:"tone_shift,omitempty"` // e.g. "casual → formal"
	Specificity float64      `json:"specificity_change"`   // ratio of content-length change
}

// DiffChange represents a single change detected between prompt versions.
type DiffChange struct {
	Category    string `json:"category"`    // "added", "removed", "modified"
	Description string `json:"description"` // human-readable change summary
	Impact      string `json:"impact"`      // "high", "medium", "low"
}
