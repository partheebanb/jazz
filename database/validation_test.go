package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLimit(t *testing.T) {
	tests := []struct {
		name         string
		limit        int
		defaultLimit int
		maxLimit     int
		expected     int
	}{
		{
			name:         "use provided limit",
			limit:        10,
			defaultLimit: 50,
			maxLimit:     1000,
			expected:     10,
		},
		{
			name:         "use default when zero",
			limit:        0,
			defaultLimit: 50,
			maxLimit:     1000,
			expected:     50,
		},
		{
			name:         "use default when negative",
			limit:        -10,
			defaultLimit: 50,
			maxLimit:     1000,
			expected:     50,
		},
		{
			name:         "cap at max",
			limit:        5000,
			defaultLimit: 50,
			maxLimit:     1000,
			expected:     1000,
		},
		{
			name:         "exactly at max",
			limit:        1000,
			defaultLimit: 50,
			maxLimit:     1000,
			expected:     1000,
		},
		{
			name:         "one below max",
			limit:        999,
			defaultLimit: 50,
			maxLimit:     1000,
			expected:     999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateLimit(tt.limit, tt.defaultLimit, tt.maxLimit)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateOffset(t *testing.T) {
	tests := []struct {
		name     string
		offset   int
		expected int
	}{
		{
			name:     "positive offset",
			offset:   10,
			expected: 10,
		},
		{
			name:     "zero offset",
			offset:   0,
			expected: 0,
		},
		{
			name:     "negative offset becomes zero",
			offset:   -10,
			expected: 0,
		},
		{
			name:     "large offset",
			offset:   1000000,
			expected: 1000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateOffset(tt.offset)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseRFC3339(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid RFC3339",
			input:   "2024-11-22T10:30:00Z",
			wantErr: false,
		},
		{
			name:    "valid with timezone offset",
			input:   "2024-11-22T10:30:00-08:00",
			wantErr: false,
		},
		{
			name:    "valid with milliseconds",
			input:   "2024-11-22T10:30:00.123Z",
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "2024-11-22",
			wantErr: true,
		},
		{
			name:    "invalid date",
			input:   "not-a-date",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseRFC3339(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, result.IsZero())
			} else {
				assert.NoError(t, err)
				assert.False(t, result.IsZero())
			}
		})
	}
}
