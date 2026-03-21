package jsonutil

import (
	"encoding/json"
	"strings"
)

// IsValidJSON checks if a string is valid JSON.
func IsValidJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

// AutoFixJSON attempts to fix common JSON errors (like single quotes)
// and returns the fixed string and whether it was modified.
func AutoFixJSON(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return s, false
	}

	if IsValidJSON(s) {
		return s, false
	}

	// Try replacing single quotes with double quotes
	// This is a naive approach but works for many simple cases
	// A more robust approach would involve a proper parser if this isn't enough
	fixed := strings.ReplaceAll(s, "'", "\"")
	if IsValidJSON(fixed) {
		return fixed, true
	}

	return s, false
}
