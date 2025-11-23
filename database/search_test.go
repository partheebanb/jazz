package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchQueryParser_Parse(t *testing.T) {
	parser := NewSearchQueryParser()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "basic two words",
			input:    "hello world",
			expected: "hello & world",
			wantErr:  false,
		},
		{
			name:     "single word",
			input:    "error",
			expected: "error",
			wantErr:  false,
		},
		{
			name:     "multiple words",
			input:    "database connection timeout",
			expected: "database & connection & timeout",
			wantErr:  false,
		},
		{
			name:     "mixed case",
			input:    "Database ERROR Timeout",
			expected: "database & error & timeout",
			wantErr:  false,
		},
		{
			name:     "extra whitespace",
			input:    "  hello   world  ",
			expected: "hello & world",
			wantErr:  false,
		},
		{
			name:     "with quotes removed",
			input:    `"connection failed"`,
			expected: "connection & failed",
			wantErr:  false,
		},
		{
			name:     "with special characters",
			input:    "error (timeout)",
			expected: "error & timeout",
			wantErr:  false,
		},
		{
			name:    "too short",
			input:   "ab",
			wantErr: true,
			errMsg:  "must be at least 3 characters",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errMsg:  "must be at least 3 characters",
		},
		{
			name:    "only whitespace",
			input:   "   ",
			wantErr: true,
			errMsg:  "must be at least 3 characters",
		},
		{
			name:    "only short words",
			input:   "a b c",
			wantErr: true,
			errMsg:  "no valid search terms",
		},
		{
			name:     "mixed short and long words",
			input:    "a database b error",
			expected: "database & error",
			wantErr:  false,
		},
		{
			name:    "too long",
			input:   string(make([]byte, 1001)),
			wantErr: true,
			errMsg:  "too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSearchQueryParser_Sanitize(t *testing.T) {
	parser := NewSearchQueryParser()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes quotes",
			input:    `"hello"`,
			expected: "hello",
		},
		{
			name:     "removes single quotes",
			input:    "'hello'",
			expected: "hello",
		},
		{
			name:     "removes parentheses",
			input:    "(hello)",
			expected: "hello",
		},
		{
			name:     "removes multiple special chars",
			input:    `"(hello)" '(world)'`,
			expected: "hello world",
		},
		{
			name:     "keeps normal text",
			input:    "hello world",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.sanitize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSearchQueryParser_FilterValidWords(t *testing.T) {
	parser := NewSearchQueryParser()

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "all valid words",
			input:    []string{"hello", "world"},
			expected: []string{"hello", "world"},
		},
		{
			name:     "filters single char",
			input:    []string{"a", "hello", "b"},
			expected: []string{"hello"},
		},
		{
			name:     "keeps two char words",
			input:    []string{"ab", "hello"},
			expected: []string{"ab", "hello"},
		},
		{
			name:     "converts to lowercase",
			input:    []string{"Hello", "WORLD"},
			expected: []string{"hello", "world"},
		},
		{
			name:     "all single chars",
			input:    []string{"a", "b", "c"},
			expected: []string{},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.filterValidWords(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
