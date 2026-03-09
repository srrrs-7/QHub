package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// resourceTablePrinters maps resource-aware print functions for specific data shapes.
// These provide curated column selections for known resource types.

// printOrgTable prints organization data with curated columns.
// Columns: SLUG | NAME | PLAN | CREATED
func printOrgTable(data any) {
	printResourceTable(data, []string{"slug", "name", "plan", "created_at"})
}

// printProjectTable prints project data with curated columns.
// Columns: SLUG | NAME | DESCRIPTION | CREATED
func printProjectTable(data any) {
	printResourceTable(data, []string{"slug", "name", "description", "created_at"})
}

// printPromptTable prints prompt data with curated columns.
// Columns: SLUG | NAME | TYPE | LATEST VERSION | CREATED
func printPromptTable(data any) {
	printResourceTable(data, []string{"slug", "name", "prompt_type", "latest_version", "created_at"})
}

// printVersionTable prints version data with curated columns.
// Columns: VERSION | STATUS | CHANGE DESCRIPTION | CREATED
func printVersionTable(data any) {
	printResourceTable(data, []string{"version_number", "status", "change_description", "created_at"})
}

// printLogTable prints execution log data with curated columns.
// Columns: ID | MODEL | LATENCY | TOKENS | STATUS | EXECUTED AT
func printLogTable(data any) {
	printResourceTable(data, []string{"id", "model", "latency_ms", "total_tokens", "status", "executed_at"})
}

// printEvalTable prints evaluation data with curated columns.
// Columns: ID | OVERALL | ACCURACY | RELEVANCE | EVALUATOR | CREATED
func printEvalTable(data any) {
	printResourceTable(data, []string{"id", "overall_score", "accuracy_score", "relevance_score", "evaluator_type", "created_at"})
}

// printLintTable prints lint results with score and issues.
func printLintTable(data any) {
	m, ok := data.(map[string]any)
	if !ok {
		printTable(data)
		return
	}

	// Print overall score.
	if score, ok := m["score"]; ok {
		fmt.Fprintf(os.Stderr, "Lint score: %s/100\n\n", formatValue(score))
	}

	// Print issues as a table.
	if issues, ok := m["issues"]; ok {
		if arr, ok := issues.([]any); ok && len(arr) > 0 {
			printCuratedArrayTable(arr, []string{"rule", "severity", "message"})
			return
		}
	}

	fmt.Println("No lint issues found.")
}

// printDiffTable prints semantic diff results.
func printDiffTable(data any) {
	m, ok := data.(map[string]any)
	if !ok {
		printTable(data)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ASPECT\tCHANGE")

	for _, key := range []string{"length_change", "variables_change", "tone_change", "specificity_change"} {
		if v, ok := m[key]; ok {
			label := strings.ReplaceAll(strings.TrimSuffix(key, "_change"), "_", " ")
			label = strings.ToUpper(label[:1]) + label[1:]
			fmt.Fprintf(w, "%s\t%s\n", label, formatValue(v))
		}
	}
	w.Flush()

	// Print text diffs if present.
	if diffs, ok := m["diffs"]; ok {
		if arr, ok := diffs.([]any); ok && len(arr) > 0 {
			fmt.Println()
			printCuratedArrayTable(arr, []string{"type", "content"})
		}
	}
}

// printAnalyticsTable prints analytics data as key-value pairs or nested tables.
func printAnalyticsTable(data any) {
	m, ok := data.(map[string]any)
	if !ok {
		printTable(data)
		return
	}

	// Print top-level stats as key-value.
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, key := range []string{"total_executions", "avg_latency_ms", "avg_tokens", "success_rate", "total_cost"} {
		if v, ok := m[key]; ok {
			label := strings.ReplaceAll(key, "_", " ")
			label = strings.ToUpper(label[:1]) + label[1:]
			fmt.Fprintf(w, "%s\t%s\n", label, formatValue(v))
		}
	}
	w.Flush()

	// Print trend data if present.
	if trend, ok := m["trend"]; ok {
		if arr, ok := trend.([]any); ok && len(arr) > 0 {
			fmt.Println()
			printCuratedArrayTable(arr, []string{"date", "executions", "avg_latency_ms", "success_rate"})
		}
	}
}

// printSearchTable prints search results with curated columns.
// Columns: PROMPT | VERSION | SIMILARITY | CHANGE DESC
func printSearchTable(data any) {
	items := extractArray(data, "results")
	if items == nil {
		printTable(data)
		return
	}
	printCuratedArrayTable(items, []string{"prompt_slug", "version_number", "similarity_score", "change_description"})
}

// printTagTable prints tag data with curated columns.
func printTagTable(data any) {
	printResourceTable(data, []string{"id", "name", "color", "created_at"})
}

// printSessionTable prints consulting session data.
func printSessionTable(data any) {
	printResourceTable(data, []string{"id", "title", "status", "created_at"})
}

// printMessageTable prints consulting message data.
func printMessageTable(data any) {
	printResourceTable(data, []string{"role", "content", "created_at"})
}

// printMemberTable prints organization member data.
func printMemberTable(data any) {
	printResourceTable(data, []string{"user_id", "role", "created_at"})
}

// printAPIKeyTable prints API key data.
func printAPIKeyTable(data any) {
	printResourceTable(data, []string{"id", "name", "created_at"})
}

// printIndustryTable prints industry configuration data.
func printIndustryTable(data any) {
	printResourceTable(data, []string{"slug", "name", "description", "created_at"})
}

// --- Helpers ---

// printResourceTable prints data using a curated set of columns.
// It unwraps nested array keys if the data is a wrapper object.
func printResourceTable(data any, columns []string) {
	switch d := data.(type) {
	case []any:
		printCuratedArrayTable(d, columns)
	case map[string]any:
		// Try to unwrap nested arrays.
		items := extractArray(data, "")
		if items != nil {
			printCuratedArrayTable(items, columns)
			return
		}
		// Single object: print key-value with curated keys.
		printCuratedObjectTable(d, columns)
	default:
		printTable(data)
	}
}

// printCuratedArrayTable renders only the specified columns from an array of maps.
func printCuratedArrayTable(items []any, columns []string) {
	if len(items) == 0 {
		fmt.Println("No results.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Header.
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = strings.ToUpper(strings.ReplaceAll(col, "_", " "))
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Rows.
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		vals := make([]string, len(columns))
		for i, col := range columns {
			vals[i] = formatValue(m[col])
		}
		fmt.Fprintln(w, strings.Join(vals, "\t"))
	}

	w.Flush()
}

// printCuratedObjectTable renders a single object with only the specified keys.
func printCuratedObjectTable(m map[string]any, columns []string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, col := range columns {
		if v, ok := m[col]; ok {
			label := strings.ToUpper(strings.ReplaceAll(col, "_", " "))
			fmt.Fprintf(w, "%s\t%s\n", label, formatValue(v))
		}
	}
	// Also print any remaining keys not in the curated list.
	for _, key := range orderedColumns(m) {
		found := false
		for _, col := range columns {
			if col == key {
				found = true
				break
			}
		}
		if !found {
			label := strings.ToUpper(strings.ReplaceAll(key, "_", " "))
			fmt.Fprintf(w, "%s\t%s\n", label, formatValue(m[key]))
		}
	}
	w.Flush()
}

// extractArray looks for a nested array within a map. If key is empty, it tries common keys.
func extractArray(data any, key string) []any {
	m, ok := data.(map[string]any)
	if !ok {
		return nil
	}
	if key != "" {
		if arr, ok := m[key].([]any); ok {
			return arr
		}
		return nil
	}
	for _, k := range []string{"data", "results", "items", "versions", "sessions", "messages", "members", "tags", "logs", "evaluations", "benchmarks", "industries"} {
		if arr, ok := m[k].([]any); ok {
			return arr
		}
	}
	return nil
}
