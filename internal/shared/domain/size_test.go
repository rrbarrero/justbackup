package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"", 0, false},
		{"1024", 1024, false},
		{"1KB", 1024, false},
		{"1.5KB", 1536, false},
		{"1MB", 1024 * 1024, false},
		{"1.2MB", 1258291, false}, // 1.2 * 1024 * 1024 = 1258291.2 -> 1258291
		{"1GB", 1024 * 1024 * 1024, false},
		{"1.5GB", 1610612736, false},
		{"1TB", 1024 * 1024 * 1024 * 1024, false},
		{"  500 MB  ", 500 * 1024 * 1024, false},
		{"1.2.3MB", 0, true},
		{"abc", 0, true},
		{"1ZB", 0, true}, // Unknown unit
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseSize(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1024 * 1024, "1.00 MB"},
		{1288490188, "1.20 GB"},
		{1024 * 1024 * 1024 * 1024, "1.00 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := FormatSize(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
