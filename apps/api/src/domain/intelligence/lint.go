package intelligence

// LintResult represents the outcome of linting a prompt version.
// Score ranges from 0 (many issues) to 100 (no issues).
type LintResult struct {
	Score  int         `json:"score"`  // 0–100
	Issues []LintIssue `json:"issues"` // detected violations
	Passed []string    `json:"passed"` // rule names that passed
}

// LintIssue represents a single lint violation found in a prompt version.
type LintIssue struct {
	Rule       string `json:"rule"`                 // e.g. "excessive-length"
	Severity   string `json:"severity"`             // "error", "warning", "info"
	Message    string `json:"message"`              // human-readable explanation
	Suggestion string `json:"suggestion,omitempty"` // how to fix it
}
