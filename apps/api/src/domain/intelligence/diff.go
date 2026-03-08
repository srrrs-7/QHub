package intelligence

// SemanticDiff represents the structured comparison between two prompt versions.
type SemanticDiff struct {
	Summary     string       `json:"summary"`
	Changes     []DiffChange `json:"changes"`
	ToneShift   string       `json:"tone_shift,omitempty"`
	Specificity float64      `json:"specificity_change"`
}

// DiffChange represents a single change detected between prompt versions.
type DiffChange struct {
	Category    string `json:"category"`    // "added", "removed", "modified"
	Description string `json:"description"`
	Impact      string `json:"impact"` // "high", "medium", "low"
}
