package diffservice

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"

	"api/src/domain/apperror"
	"api/src/domain/intelligence"
	db "utils/db/db"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

// DiffService generates semantic diffs between prompt versions.
type DiffService struct {
	q db.Querier
}

// NewDiffService creates a new DiffService with the given querier.
func NewDiffService(q db.Querier) *DiffService {
	return &DiffService{q: q}
}

// toneKeywords maps tone categories to their indicator words.
var toneKeywords = map[string][]string{
	"formal":   {"please", "kindly", "shall", "must", "ensure", "required"},
	"casual":   {"just", "simply", "like", "okay", "hey", "cool"},
	"strict":   {"must", "always", "never", "required", "mandatory", "shall not"},
	"friendly": {"feel free", "no worries", "happy to", "glad", "welcome"},
}

// variablePattern matches {{variable}} patterns in prompt content.
var variablePattern = regexp.MustCompile(`\{\{(\w+)\}\}`)

// GenerateDiff computes a rule-based semantic diff between two prompt versions
// and stores the result in the database.
func (s *DiffService) GenerateDiff(ctx context.Context, promptID uuid.UUID, fromVersion, toVersion int) (*intelligence.SemanticDiff, error) {
	fromRow, err := s.q.GetPromptVersion(ctx, db.GetPromptVersionParams{
		PromptID:      promptID,
		VersionNumber: int32(fromVersion),
	})
	if err != nil {
		return nil, apperror.NewNotFoundError(
			fmt.Errorf("version %d not found for prompt %s", fromVersion, promptID),
			"PromptVersion",
		)
	}

	toRow, err := s.q.GetPromptVersion(ctx, db.GetPromptVersionParams{
		PromptID:      promptID,
		VersionNumber: int32(toVersion),
	})
	if err != nil {
		return nil, apperror.NewNotFoundError(
			fmt.Errorf("version %d not found for prompt %s", toVersion, promptID),
			"PromptVersion",
		)
	}

	fromContent := extractContent(fromRow.Content)
	toContent := extractContent(toRow.Content)

	diff := buildDiff(fromContent, toContent)

	diffJSON, err := json.Marshal(diff)
	if err != nil {
		return nil, apperror.NewInternalServerError(fmt.Errorf("failed to marshal semantic diff: %w", err), "SemanticDiff")
	}

	if err := s.q.UpdatePromptVersionSemanticDiff(ctx, db.UpdatePromptVersionSemanticDiffParams{
		ID: toRow.ID,
		SemanticDiff: pqtype.NullRawMessage{
			RawMessage: diffJSON,
			Valid:      true,
		},
	}); err != nil {
		return nil, apperror.NewDatabaseError(fmt.Errorf("failed to store semantic diff: %w", err), "SemanticDiff")
	}

	return diff, nil
}

// TextDiffResult represents a line-by-line text diff result.
type TextDiffResult struct {
	FromVersion int             `json:"from_version"`
	ToVersion   int             `json:"to_version"`
	Hunks       []TextDiffHunk  `json:"hunks"`
	Stats       TextDiffStats   `json:"stats"`
}

type TextDiffHunk struct {
	Lines []TextDiffLine `json:"lines"`
}

type TextDiffLine struct {
	Type    string `json:"type"` // "equal", "added", "removed"
	Content string `json:"content"`
	OldLine int    `json:"old_line,omitempty"`
	NewLine int    `json:"new_line,omitempty"`
}

type TextDiffStats struct {
	Added   int `json:"added"`
	Removed int `json:"removed"`
	Equal   int `json:"equal"`
}

// GenerateTextDiff produces a line-by-line text diff between two prompt versions.
func (s *DiffService) GenerateTextDiff(ctx context.Context, promptID uuid.UUID, fromVersion, toVersion int) (*TextDiffResult, error) {
	fromRow, err := s.q.GetPromptVersion(ctx, db.GetPromptVersionParams{
		PromptID:      promptID,
		VersionNumber: int32(fromVersion),
	})
	if err != nil {
		return nil, apperror.NewNotFoundError(
			fmt.Errorf("version %d not found for prompt %s", fromVersion, promptID),
			"PromptVersion",
		)
	}

	toRow, err := s.q.GetPromptVersion(ctx, db.GetPromptVersionParams{
		PromptID:      promptID,
		VersionNumber: int32(toVersion),
	})
	if err != nil {
		return nil, apperror.NewNotFoundError(
			fmt.Errorf("version %d not found for prompt %s", toVersion, promptID),
			"PromptVersion",
		)
	}

	fromContent := extractContent(fromRow.Content)
	toContent := extractContent(toRow.Content)

	fromLines := strings.Split(fromContent, "\n")
	toLines := strings.Split(toContent, "\n")

	lcs := lcsStrings(fromLines, toLines)

	var lines []TextDiffLine
	var stats TextDiffStats
	fi, ti, li := 0, 0, 0
	oldLine, newLine := 1, 1

	for li < len(lcs) {
		for fi < len(fromLines) && fromLines[fi] != lcs[li] {
			lines = append(lines, TextDiffLine{Type: "removed", Content: fromLines[fi], OldLine: oldLine})
			stats.Removed++
			fi++
			oldLine++
		}
		for ti < len(toLines) && toLines[ti] != lcs[li] {
			lines = append(lines, TextDiffLine{Type: "added", Content: toLines[ti], NewLine: newLine})
			stats.Added++
			ti++
			newLine++
		}
		lines = append(lines, TextDiffLine{Type: "equal", Content: lcs[li], OldLine: oldLine, NewLine: newLine})
		stats.Equal++
		fi++
		ti++
		li++
		oldLine++
		newLine++
	}
	for fi < len(fromLines) {
		lines = append(lines, TextDiffLine{Type: "removed", Content: fromLines[fi], OldLine: oldLine})
		stats.Removed++
		fi++
		oldLine++
	}
	for ti < len(toLines) {
		lines = append(lines, TextDiffLine{Type: "added", Content: toLines[ti], NewLine: newLine})
		stats.Added++
		ti++
		newLine++
	}

	return &TextDiffResult{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Hunks:       []TextDiffHunk{{Lines: lines}},
		Stats:       stats,
	}, nil
}

// lcsStrings computes the longest common subsequence of two string slices.
func lcsStrings(a, b []string) []string {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}
	result := make([]string, 0, dp[m][n])
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			result = append(result, a[i-1])
			i--
			j--
		} else if dp[i-1][j] >= dp[i][j-1] {
			i--
		} else {
			j--
		}
	}
	for l, r := 0, len(result)-1; l < r; l, r = l+1, r-1 {
		result[l], result[r] = result[r], result[l]
	}
	return result
}

// extractContent pulls the text content from the JSON content field.
// It tries to extract a "content" or "text" string field; otherwise returns the raw JSON as string.
func extractContent(raw json.RawMessage) string {
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return string(raw)
	}
	for _, key := range []string{"content", "text", "body"} {
		if v, ok := obj[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return string(raw)
}

// buildDiff performs a rule-based comparison of two content strings.
func buildDiff(fromContent, toContent string) *intelligence.SemanticDiff {
	var changes []intelligence.DiffChange

	// Length change analysis
	lenDiff := len(toContent) - len(fromContent)
	if lenDiff != 0 {
		impact := "low"
		if abs(lenDiff) > 500 {
			impact = "high"
		} else if abs(lenDiff) > 100 {
			impact = "medium"
		}

		category := "modified"
		desc := fmt.Sprintf("Content length changed from %d to %d characters (%+d)", len(fromContent), len(toContent), lenDiff)
		changes = append(changes, intelligence.DiffChange{
			Category:    category,
			Description: desc,
			Impact:      impact,
		})
	}

	// Variable changes
	fromVars := findVariables(fromContent)
	toVars := findVariables(toContent)

	for v := range toVars {
		if !fromVars[v] {
			changes = append(changes, intelligence.DiffChange{
				Category:    "added",
				Description: fmt.Sprintf("Variable {{%s}} added", v),
				Impact:      "high",
			})
		}
	}
	for v := range fromVars {
		if !toVars[v] {
			changes = append(changes, intelligence.DiffChange{
				Category:    "removed",
				Description: fmt.Sprintf("Variable {{%s}} removed", v),
				Impact:      "high",
			})
		}
	}

	// Tone shift detection
	fromTone := detectTone(fromContent)
	toTone := detectTone(toContent)
	toneShift := ""
	if fromTone != toTone {
		toneShift = fmt.Sprintf("%s → %s", fromTone, toTone)
		changes = append(changes, intelligence.DiffChange{
			Category:    "modified",
			Description: fmt.Sprintf("Tone shifted from %s to %s", fromTone, toTone),
			Impact:      "medium",
		})
	}

	// Specificity change: ratio of content length change as a simple proxy
	specificity := 0.0
	if len(fromContent) > 0 {
		specificity = math.Round((float64(len(toContent))/float64(len(fromContent))-1.0)*100) / 100
	}

	// Build summary
	summary := buildSummary(changes)

	return &intelligence.SemanticDiff{
		Summary:     summary,
		Changes:     changes,
		ToneShift:   toneShift,
		Specificity: specificity,
	}
}

// findVariables extracts all {{variable}} names from content.
func findVariables(content string) map[string]bool {
	vars := make(map[string]bool)
	matches := variablePattern.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		vars[m[1]] = true
	}
	return vars
}

// detectTone scores content against tone keyword categories and returns the dominant tone.
func detectTone(content string) string {
	lower := strings.ToLower(content)
	bestTone := "neutral"
	bestScore := 0

	for tone, keywords := range toneKeywords {
		score := 0
		for _, kw := range keywords {
			score += strings.Count(lower, kw)
		}
		if score > bestScore {
			bestScore = score
			bestTone = tone
		}
	}

	return bestTone
}

// buildSummary creates a human-readable summary from the list of changes.
func buildSummary(changes []intelligence.DiffChange) string {
	if len(changes) == 0 {
		return "No significant changes detected"
	}

	parts := make([]string, 0, len(changes))
	for _, c := range changes {
		parts = append(parts, c.Description)
	}

	return strings.Join(parts, "; ")
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
