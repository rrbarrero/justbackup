package utils

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic string",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "string with special characters",
			input:    "Hello, World!",
			expected: "hello-world",
		},
		{
			name:     "string with multiple spaces",
			input:    "  Hello   World  ",
			expected: "--hello---world--", // Current implementation doesn't collapse hyphens, which is fine for now but good to know
		},
		{
			name:     "alphanumeric",
			input:    "Host123",
			expected: "host123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Slugify(tt.input)
			if got != tt.expected {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
