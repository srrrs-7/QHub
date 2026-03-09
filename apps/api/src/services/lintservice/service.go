// Package lintservice performs rule-based quality analysis on prompt versions.
//
// Rules include: excessive content length, missing output format specification,
// undeclared template variables, vague instructions, missing constraints, and
// prompt injection risk detection. Each rule contributes to a 0–100 quality score.
package lintservice

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"api/src/domain/apperror"
	"api/src/domain/intelligence"
	"api/src/domain/prompt"
	"api/src/services/contentutil"

	"github.com/google/uuid"
)

// LintService performs rule-based linting on prompt versions.
type LintService struct {
	versionRepo prompt.VersionRepository
}

// NewLintService creates a new LintService with the given version repository.
func NewLintService(versionRepo prompt.VersionRepository) *LintService {
	return &LintService{versionRepo: versionRepo}
}

// vagueWords are words that indicate imprecise instructions.
var vagueWords = []string{"good", "nice", "appropriate", "proper", "suitable", "reasonable"}

// outputFormatKeywords indicate that the prompt specifies an output format.
var outputFormatKeywords = []string{"format", "json", "xml", "csv", "markdown", "yaml", "html"}

// maxContentLength is the threshold for the excessive-length rule.
const maxContentLength = 4000

// minConstraintLength is the minimum content length for the missing-constraints rule.
// Short prompts (<=200 chars) may not need explicit constraints.
const minConstraintLength = 200

// negativeConstraintKeywords indicate the prompt has negative constraints/boundaries.
var negativeConstraintKeywords = []string{
	"do not", "never", "avoid", "don't", "must not",
	// Japanese
	"してはいけない", "禁止",
}

// limitConstraintKeywords indicate the prompt has length/format constraints.
var limitConstraintKeywords = []string{
	"max", "limit", "within", "at most", "no more than",
	// Japanese
	"制限", "以内", "まで",
}

// injectionPatterns are regex patterns that indicate prompt injection risk.
var injectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)ignore\s+(previous|above)\s+(instructions?|prompts?|context)`),
	regexp.MustCompile(`(?i)ignore\s+above`),
	regexp.MustCompile(`(?i)(you\s+are\s+now|act\s+as)\b`),
	regexp.MustCompile(`(?i)do\s+anything`),
	regexp.MustCompile(`(?i)no\s+restrictions`),
	// Japanese
	regexp.MustCompile(`前の指示を無視`),
	regexp.MustCompile(`制限なし`),
}

// unguardedInputPattern matches user input placeholders like {{user_input}} or {{query}}
// that are not surrounded by guardrail instructions.
var userInputPlaceholderNames = []string{"user_input", "query", "user_query", "input"}

// LintVersion lints a specific prompt version and stores the result.
func (s *LintService) LintVersion(ctx context.Context, promptID uuid.UUID, versionNumber int) (*intelligence.LintResult, error) {
	v, err := s.versionRepo.FindByPromptAndNumber(ctx, prompt.PromptIDFromUUID(promptID), versionNumber)
	if err != nil {
		return nil, apperror.NewNotFoundError(
			fmt.Errorf("version %d not found for prompt %s", versionNumber, promptID),
			"PromptVersion",
		)
	}

	content := contentutil.ExtractText(v.Content)
	variables := extractVariablesFromJSON(v.Variables)

	result := runLintRules(content, variables)

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, apperror.NewInternalServerError(fmt.Errorf("failed to marshal lint result: %w", err), "LintResult")
	}

	if err := s.versionRepo.UpdateLintResult(ctx, v.ID, resultJSON); err != nil {
		return nil, apperror.NewDatabaseError(fmt.Errorf("failed to store lint result: %w", err), "LintResult")
	}

	return result, nil
}

// extractVariablesFromJSON parses the variables JSON array into a set of variable names.
func extractVariablesFromJSON(raw json.RawMessage) map[string]bool {
	vars := make(map[string]bool)
	if raw == nil {
		return vars
	}

	// Try array of objects with "name" field
	var arr []map[string]any
	if err := json.Unmarshal(raw, &arr); err == nil {
		for _, item := range arr {
			if name, ok := item["name"].(string); ok {
				vars[name] = true
			}
		}
		return vars
	}

	// Try array of strings
	var strArr []string
	if err := json.Unmarshal(raw, &strArr); err == nil {
		for _, name := range strArr {
			vars[name] = true
		}
	}

	return vars
}

// runLintRules applies all lint rules and returns the aggregated result.
func runLintRules(content string, variables map[string]bool) *intelligence.LintResult {
	var issues []intelligence.LintIssue
	var passed []string

	// Rule: excessive-length
	if len(content) > maxContentLength {
		issues = append(issues, intelligence.LintIssue{
			Rule:       "excessive-length",
			Severity:   "warning",
			Message:    fmt.Sprintf("Content is %d characters, exceeding the recommended limit of %d", len(content), maxContentLength),
			Suggestion: "Consider breaking the prompt into smaller, focused sections",
		})
	} else {
		passed = append(passed, "excessive-length")
	}

	// Rule: missing-output-format
	lower := strings.ToLower(content)
	hasOutputFormat := false
	for _, kw := range outputFormatKeywords {
		if strings.Contains(lower, kw) {
			hasOutputFormat = true
			break
		}
	}
	if !hasOutputFormat {
		issues = append(issues, intelligence.LintIssue{
			Rule:       "missing-output-format",
			Severity:   "warning",
			Message:    "No output format specification detected in the prompt",
			Suggestion: "Consider specifying the desired output format (e.g., JSON, markdown, plain text)",
		})
	} else {
		passed = append(passed, "missing-output-format")
	}

	// Rule: variable-check
	contentVars := contentutil.FindVariables(content)
	undeclared := findUndeclaredVariables(contentVars, variables)
	if len(undeclared) > 0 {
		issues = append(issues, intelligence.LintIssue{
			Rule:       "variable-check",
			Severity:   "error",
			Message:    fmt.Sprintf("Undeclared variables found in content: %s", strings.Join(undeclared, ", ")),
			Suggestion: "Add the missing variables to the variables array or remove them from the content",
		})
	} else {
		passed = append(passed, "variable-check")
	}

	// Rule: no-vague-instruction
	vagueFound := findVagueWords(content)
	if len(vagueFound) > 0 {
		issues = append(issues, intelligence.LintIssue{
			Rule:       "no-vague-instruction",
			Severity:   "info",
			Message:    fmt.Sprintf("Vague instructions detected: %s", strings.Join(vagueFound, ", ")),
			Suggestion: "Replace vague terms with specific, measurable criteria",
		})
	} else {
		passed = append(passed, "no-vague-instruction")
	}

	// Rule: missing-constraints
	if checkMissingConstraints(content) {
		issues = append(issues, intelligence.LintIssue{
			Rule:       "missing-constraints",
			Severity:   "warning",
			Message:    "Prompt lacks explicit constraints or boundaries",
			Suggestion: "Add negative constraints (e.g., \"do not\", \"never\") or limit constraints (e.g., \"at most\", \"within\") to guide model behavior",
		})
	} else {
		passed = append(passed, "missing-constraints")
	}

	// Rule: prompt-injection-risk
	if injectionIssues := checkPromptInjectionRisk(content); len(injectionIssues) > 0 {
		issues = append(issues, intelligence.LintIssue{
			Rule:       "prompt-injection-risk",
			Severity:   "error",
			Message:    fmt.Sprintf("Potential prompt injection risk detected: %s", strings.Join(injectionIssues, "; ")),
			Suggestion: "Remove or rewrite patterns that could be exploited for prompt injection, and add guardrails around user input placeholders",
		})
	} else {
		passed = append(passed, "prompt-injection-risk")
	}

	score := calculateScore(issues)

	return &intelligence.LintResult{
		Score:  score,
		Issues: issues,
		Passed: passed,
	}
}

// findUndeclaredVariables returns variable names used in content but not declared.
func findUndeclaredVariables(contentVars, declaredVars map[string]bool) []string {
	var undeclared []string
	for v := range contentVars {
		if !declaredVars[v] {
			undeclared = append(undeclared, "{{"+v+"}}")
		}
	}
	return undeclared
}

// findVagueWords returns vague words found in content.
func findVagueWords(content string) []string {
	lower := strings.ToLower(content)
	var found []string
	seen := make(map[string]bool)
	for _, word := range vagueWords {
		if strings.Contains(lower, word) && !seen[word] {
			found = append(found, word)
			seen[word] = true
		}
	}
	return found
}

// checkMissingConstraints returns true if the prompt is long enough to need
// constraints but lacks both negative constraints and limit constraints.
func checkMissingConstraints(content string) bool {
	if len(content) <= minConstraintLength {
		return false
	}
	lower := strings.ToLower(content)

	hasNegative := false
	for _, kw := range negativeConstraintKeywords {
		if strings.Contains(lower, kw) {
			hasNegative = true
			break
		}
	}

	hasLimit := false
	for _, kw := range limitConstraintKeywords {
		if strings.Contains(lower, kw) {
			hasLimit = true
			break
		}
	}

	// Triggers only when BOTH categories are missing
	return !hasNegative && !hasLimit
}

// checkPromptInjectionRisk scans for prompt injection patterns and returns
// a list of human-readable descriptions of detected risks.
func checkPromptInjectionRisk(content string) []string {
	var found []string
	seen := make(map[string]bool)

	// Check regex patterns
	for _, pat := range injectionPatterns {
		if match := pat.FindString(content); match != "" {
			desc := fmt.Sprintf("found pattern \"%s\"", match)
			if !seen[desc] {
				found = append(found, desc)
				seen[desc] = true
			}
		}
	}

	// Check for unguarded user input placeholders
	contentVars := contentutil.FindVariables(content)
	for _, name := range userInputPlaceholderNames {
		if contentVars[name] {
			if !hasGuardrailAroundPlaceholder(content, name) {
				desc := fmt.Sprintf("unguarded user input placeholder {{%s}}", name)
				if !seen[desc] {
					found = append(found, desc)
					seen[desc] = true
				}
			}
		}
	}

	return found
}

// hasGuardrailAroundPlaceholder checks if the content has guardrail
// instructions near a user input placeholder. Guardrails include words
// like "sanitize", "validate", "escape", "filter", "safe", "trusted".
func hasGuardrailAroundPlaceholder(content, varName string) bool {
	lower := strings.ToLower(content)
	guardrailWords := []string{"sanitize", "validate", "escape", "filter", "safe", "trusted", "verify", "check"}
	for _, gw := range guardrailWords {
		if strings.Contains(lower, gw) {
			return true
		}
	}
	return false
}

// calculateScore computes a lint score from 0-100 based on issues found.
// Each error deducts 25 points, each warning deducts 10 points, each info deducts 5 points.
func calculateScore(issues []intelligence.LintIssue) int {
	score := 100
	for _, issue := range issues {
		switch issue.Severity {
		case "error":
			score -= 25
		case "warning":
			score -= 10
		case "info":
			score -= 5
		}
	}
	if score < 0 {
		score = 0
	}
	return score
}
