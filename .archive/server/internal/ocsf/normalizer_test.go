package ocsf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractValueByPath(t *testing.T) {
	data := map[string]interface{}{
		"level": "INFO",
		"message": "Test message",
		"user": map[string]interface{}{
			"name": "john_doe",
			"id":   "12345",
		},
	}

	tests := []struct {
		name     string
		path     string
		expected interface{}
	}{
		{
			name:     "simple field",
			path:     "level",
			expected: "INFO",
		},
		{
			name:     "nested field",
			path:     "user.name",
			expected: "john_doe",
		},
		{
			name:     "non-existent field",
			path:     "user.email",
			expected: nil,
		},
		{
			name:     "non-existent parent",
			path:     "admin.name",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractValueByPath(data, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSeverityToID(t *testing.T) {
	tests := []struct {
		name     string
		severity string
		expected int
	}{
		{
			name:     "critical",
			severity: "critical",
			expected: 4,
		},
		{
			name:     "high",
			severity: "HIGH",
			expected: 4,
		},
		{
			name:     "medium",
			severity: "medium",
			expected: 3,
		},
		{
			name:     "low",
			severity: "LOW",
			expected: 2,
		},
		{
			name:     "info",
			severity: "info",
			expected: 1,
		},
		{
			name:     "unknown",
			severity: "unknown",
			expected: 2, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := severityToID(tt.severity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetStringValue(t *testing.T) {
	data := map[string]interface{}{
		"message": "test message",
		"level":   "INFO",
	}

	tests := []struct {
		name     string
		key      string
		def      string
		expected string
	}{
		{
			name:     "existing key",
			key:      "message",
			def:      "default",
			expected: "test message",
		},
		{
			name:     "non-existing key",
			key:      "missing",
			def:      "default",
			expected: "default",
		},
		{
			name:     "non-string value",
			key:      "level",
			def:      "default",
			expected: "INFO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringValue(data, tt.key, tt.def)
			assert.Equal(t, tt.expected, result)
		})
	}
}