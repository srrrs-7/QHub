// Package contentutil provides shared utilities for extracting and
// analysing prompt content stored as JSON.
//
// Multiple service packages (diffservice, lintservice, embeddingservice)
// need to pull plain-text from the JSONB "content" column and detect
// {{variable}} placeholders. This package centralises that logic so
// it is defined once and tested once.
package contentutil

import (
	"encoding/json"
	"regexp"
)

// VariablePattern matches {{variable}} placeholders in prompt content.
// The first capture group holds the variable name (word characters only).
var VariablePattern = regexp.MustCompile(`\{\{(\w+)\}\}`)

// ExtractText extracts the plain-text portion from a JSON content field.
//
// It attempts to unmarshal the raw JSON as an object and returns the first
// string value found for the keys "content", "text", "body", "system",
// or "user" (checked in that order). If none match—or if the data is not
// valid JSON—the raw bytes are returned as-is.
func ExtractText(raw json.RawMessage) string {
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return string(raw)
	}
	for _, key := range []string{"content", "text", "body", "system", "user"} {
		if v, ok := obj[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return string(raw)
}

// FindVariables scans content for {{variable}} placeholders and returns
// a set of the variable names found. Duplicate names are de-duplicated.
func FindVariables(content string) map[string]bool {
	vars := make(map[string]bool)
	matches := VariablePattern.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		vars[m[1]] = true
	}
	return vars
}
