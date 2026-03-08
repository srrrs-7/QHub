package intelligence

// LintResult represents the outcome of linting a prompt version.
type LintResult struct {
	Score  int         `json:"score"` // 0-100
	Issues []LintIssue `json:"issues"`
	Passed []string    `json:"passed"`
}

// LintIssue represents a single lint violation found in a prompt version.
type LintIssue struct {
	Rule       string `json:"rule"`
	Severity   string `json:"severity"` // "error", "warning", "info"
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}
