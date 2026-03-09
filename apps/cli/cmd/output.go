package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

// printResult dispatches to JSON or table based on --output flag.
func printResult(v any) {
	if outputFmt == "table" {
		printTable(v)
	} else {
		printJSON(v)
	}
}

// printTable renders data as a formatted table.
// Handles: single object (map), array of objects, or nested response with data array.
func printTable(v any) {
	switch data := v.(type) {
	case []any:
		printArrayTable(data)
	case map[string]any:
		// Check for nested array keys commonly returned by the API.
		for _, key := range []string{"data", "results", "items", "versions", "sessions", "messages", "members", "tags", "logs", "evaluations", "benchmarks", "industries"} {
			if arr, ok := data[key]; ok {
				if items, ok := arr.([]any); ok {
					printArrayTable(items)
					return
				}
			}
		}
		// Single object: render key-value pairs vertically.
		printObjectTable(data)
	default:
		// Fallback to JSON for types we cannot render as table.
		printJSON(v)
	}
}

// printArrayTable renders a slice of maps as a columnar table.
func printArrayTable(items []any) {
	if len(items) == 0 {
		fmt.Println("No results.")
		return
	}

	// Collect all keys from the first item.
	firstMap, ok := items[0].(map[string]any)
	if !ok {
		printJSON(items)
		return
	}

	columns := orderedColumns(firstMap)
	if len(columns) == 0 {
		printJSON(items)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Header row.
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = strings.ToUpper(col)
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Data rows.
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

// printObjectTable renders a single map as vertical key-value pairs.
func printObjectTable(m map[string]any) {
	if len(m) == 0 {
		fmt.Println("No data.")
		return
	}

	columns := orderedColumns(m)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, key := range columns {
		fmt.Fprintf(w, "%s\t%s\n", strings.ToUpper(key), formatValue(m[key]))
	}
	w.Flush()
}

// orderedColumns returns map keys in a sensible display order:
// id first, name/slug/title second, status third, dates last, rest alphabetically in between.
func orderedColumns(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	priority := func(k string) int {
		lower := strings.ToLower(k)
		switch {
		case lower == "id":
			return 0
		case lower == "slug":
			return 1
		case lower == "name" || lower == "title":
			return 2
		case lower == "status":
			return 3
		case lower == "version" || lower == "version_number":
			return 4
		case strings.Contains(lower, "score"):
			return 5
		case strings.HasSuffix(lower, "_at") || strings.HasSuffix(lower, "_date") || lower == "created" || lower == "updated":
			return 100
		default:
			return 50
		}
	}

	sort.SliceStable(keys, func(i, j int) bool {
		pi, pj := priority(keys[i]), priority(keys[j])
		if pi != pj {
			return pi < pj
		}
		return keys[i] < keys[j]
	})

	return keys
}

// formatValue converts a value to a display string for table output.
func formatValue(v any) string {
	if v == nil {
		return "-"
	}
	switch val := v.(type) {
	case bool:
		if val {
			return "yes"
		}
		return "no"
	case float64:
		// JSON numbers are float64. Display integers without decimal.
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.2f", val)
	case string:
		return truncate(val, 50)
	case []any:
		if len(val) == 0 {
			return "-"
		}
		parts := make([]string, 0, len(val))
		for _, item := range val {
			parts = append(parts, fmt.Sprintf("%v", item))
		}
		return truncate(strings.Join(parts, ", "), 50)
	case map[string]any:
		return "{...}"
	default:
		s := fmt.Sprintf("%v", val)
		return truncate(s, 50)
	}
}

// truncate shortens a string to maxLen characters, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// printSuccess prints a green success message to stderr.
func printSuccess(msg string) {
	fmt.Fprintf(os.Stderr, "\033[32m%s\033[0m %s\n", "ok:", msg)
}

// printInfo prints a cyan info message to stderr.
func printInfo(msg string) {
	fmt.Fprintf(os.Stderr, "\033[36minfo:\033[0m %s\n", msg)
}

// printWarning prints a yellow warning to stderr.
func printWarning(msg string) {
	fmt.Fprintf(os.Stderr, "\033[33mwarn:\033[0m %s\n", msg)
}
