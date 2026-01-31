package utils

import (
	"testing"
)

func TestCheckConnectivity(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"Valid URL", "https://www.google.com", true},
		{"Invalid URL", "https://thisisnotavalidurlatall12345.com", false},
		{"Bad protocol", "htp://invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckConnectivity(tt.url)
			if result != tt.expected {
				t.Errorf("CheckConnectivity(%s) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestFormatYearRange(t *testing.T) {
	tests := []struct {
		start    int
		end      int
		expected string
	}{
		{2020, 2022, "2020-2022"},
		{1990, 1990, "1990"},
		{1970, 2025, "1970-2025"},
	}

	for _, tt := range tests {
		result := FormatYearRange(tt.start, tt.end)
		if result != tt.expected {
			t.Errorf("FormatYearRange(%d, %d) = %s, want %s", tt.start, tt.end, result, tt.expected)
		}
	}
}
