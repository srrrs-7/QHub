// Package diffservice generates semantic and text diffs between prompt versions.
//
// The semantic diff analyses content-length changes, variable additions/removals,
// and tone shifts. The text diff produces a line-by-line comparison using the
// longest common subsequence (LCS) algorithm.
package diffservice

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"api/src/domain/apperror"
	"api/src/domain/intelligence"
	"api/src/domain/prompt"
	"api/src/services/contentutil"

	"github.com/google/uuid"

	"utils/cache"
)

// diffCacheTTL is the duration that computed diffs are cached.
const diffCacheTTL = 24 * time.Hour

// DiffService generates semantic diffs between prompt versions.
type DiffService struct {
	versionRepo prompt.VersionRepository
	cache       *cache.Client
}

// NewDiffService creates a new DiffService with the given version repository
// and an optional cache client. If cache is nil, diffs are computed every time.
func NewDiffService(versionRepo prompt.VersionRepository, cache *cache.Client) *DiffService {
	return &DiffService{versionRepo: versionRepo, cache: cache}
}

// toneKeywords maps tone categories to their indicator words.
var toneKeywords = map[string][]string{
	"formal":   {"please", "kindly", "shall", "must", "ensure", "required"},
	"casual":   {"just", "simply", "like", "okay", "hey", "cool"},
	"strict":   {"must", "always", "never", "required", "mandatory", "shall not"},
	"friendly": {"feel free", "no worries", "happy to", "glad", "welcome"},
}

// diffCacheKey returns the cache key for a semantic diff between two versions.
func diffCacheKey(promptID uuid.UUID, fromVersion, toVersion int) string {
	return fmt.Sprintf("diff:%s:v%d:v%d", promptID.String(), fromVersion, toVersion)
}

// GenerateDiff computes a rule-based semantic diff between two prompt versions
// and stores the result in the database. Results are cached when a cache client
// is configured.
func (s *DiffService) GenerateDiff(ctx context.Context, promptID uuid.UUID, fromVersion, toVersion int) (*intelligence.SemanticDiff, error) {
	// Try cache first
	cacheKey := diffCacheKey(promptID, fromVersion, toVersion)
	if s.cache.Available() {
		var cached intelligence.SemanticDiff
		found, err := s.cache.Get(ctx, cacheKey, &cached)
		if err == nil && found {
			return &cached, nil
		}
	}

	pid := prompt.PromptIDFromUUID(promptID)

	fromV, err := s.versionRepo.FindByPromptAndNumber(ctx, pid, fromVersion)
	if err != nil {
		return nil, apperror.NewNotFoundError(
			fmt.Errorf("version %d not found for prompt %s", fromVersion, promptID),
			"PromptVersion",
		)
	}

	toV, err := s.versionRepo.FindByPromptAndNumber(ctx, pid, toVersion)
	if err != nil {
		return nil, apperror.NewNotFoundError(
			fmt.Errorf("version %d not found for prompt %s", toVersion, promptID),
			"PromptVersion",
		)
	}

	fromContent := contentutil.ExtractText(fromV.Content)
	toContent := contentutil.ExtractText(toV.Content)

	diff := buildDiff(fromContent, toContent)

	diffJSON, err := json.Marshal(diff)
	if err != nil {
		return nil, apperror.NewInternalServerError(fmt.Errorf("failed to marshal semantic diff: %w", err), "SemanticDiff")
	}

	if err := s.versionRepo.UpdateSemanticDiff(ctx, toV.ID, diffJSON); err != nil {
		return nil, apperror.NewDatabaseError(fmt.Errorf("failed to store semantic diff: %w", err), "SemanticDiff")
	}

	// Cache the result (best-effort; errors are ignored)
	if s.cache.Available() {
		_ = s.cache.Set(ctx, cacheKey, diff, diffCacheTTL)
	}

	return diff, nil
}

// TextDiffResult represents a line-by-line text diff result.
type TextDiffResult struct {
	FromVersion int            `json:"from_version"`
	ToVersion   int            `json:"to_version"`
	Hunks       []TextDiffHunk `json:"hunks"`
	Stats       TextDiffStats  `json:"stats"`
}

// TextDiffHunk groups contiguous diff lines.
type TextDiffHunk struct {
	Lines []TextDiffLine `json:"lines"`
}

// TextDiffLine represents a single line in the diff output.
type TextDiffLine struct {
	Type    string `json:"type"` // "equal", "added", "removed"
	Content string `json:"content"`
	OldLine int    `json:"old_line,omitempty"`
	NewLine int    `json:"new_line,omitempty"`
}

// TextDiffStats summarises the number of added, removed, and equal lines.
type TextDiffStats struct {
	Added   int `json:"added"`
	Removed int `json:"removed"`
	Equal   int `json:"equal"`
}

// GenerateTextDiff produces a line-by-line text diff between two prompt versions.
func (s *DiffService) GenerateTextDiff(ctx context.Context, promptID uuid.UUID, fromVersion, toVersion int) (*TextDiffResult, error) {
	pid := prompt.PromptIDFromUUID(promptID)

	fromV, err := s.versionRepo.FindByPromptAndNumber(ctx, pid, fromVersion)
	if err != nil {
		return nil, apperror.NewNotFoundError(
			fmt.Errorf("version %d not found for prompt %s", fromVersion, promptID),
			"PromptVersion",
		)
	}

	toV, err := s.versionRepo.FindByPromptAndNumber(ctx, pid, toVersion)
	if err != nil {
		return nil, apperror.NewNotFoundError(
			fmt.Errorf("version %d not found for prompt %s", toVersion, promptID),
			"PromptVersion",
		)
	}

	fromContent := contentutil.ExtractText(fromV.Content)
	toContent := contentutil.ExtractText(toV.Content)

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
	fromVars := contentutil.FindVariables(fromContent)
	toVars := contentutil.FindVariables(toContent)

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

// abs returns the absolute value of n.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
