package jsonutil

import (
	"testing"
)

func TestIsValidJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`{"key": "value"}`, true},
		{`[1, 2, 3]`, true},
		{`{"key": 'value'}`, false},
		{`invalid`, false},
		{``, false},
	}

	for _, tt := range tests {
		if got := IsValidJSON(tt.input); got != tt.expected {
			t.Errorf("IsValidJSON(%q) = %v; want %v", tt.input, got, tt.expected)
		}
	}
}

func TestAutoFixJSON(t *testing.T) {
	tests := []struct {
		input        string
		expected     string
		modified     bool
	}{
		{
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
			modified: false,
		},
		{
			input:    `{'key': 'value'}`,
			expected: `{"key": "value"}`,
			modified: true,
		},
		{
			input:    `{'society': 28, 'passcodeStatus': 'A'}`,
			expected: `{"society": 28, "passcodeStatus": "A"}`,
			modified: true,
		},
		{
			input:    `not json at all`,
			expected: `not json at all`,
			modified: false,
		},
	}

	for _, tt := range tests {
		got, modified := AutoFixJSON(tt.input)
		if got != tt.expected || modified != tt.modified {
			t.Errorf("AutoFixJSON(%q) = (%q, %v); want (%q, %v)", tt.input, got, modified, tt.expected, tt.modified)
		}
	}
}
