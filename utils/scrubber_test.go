package utils

import (
	"strings"
	"testing"
)

func TestQueryParamScrubber_Is(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid", "key1=value1&key2=value2", true},
		{"Invalid", "key1", false},
		{"Invalid_NoValue", "key1&key2", false},
	}

	scrubber := &QueryParamScrubber{keysToScrub: []string{"refresh_token"}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := scrubber.Is(test.input)
			if result != test.expected {
				t.Errorf("Expected Is('%s') to be %t, got %t", test.input, test.expected, result)
			}
		})
	}
}

func TestQueryParamScrubber_Scrub(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		containsKey bool
	}{
		{"WithRefreshToken", "key1=value1&refresh_token=abc123", "key1=value1&refresh_token=******", true},
		{"WithoutRefreshToken", "key1=value1&key2=value2", "key1=value1&key2=value2", false},
	}

	scrubber := &QueryParamScrubber{keysToScrub: []string{"refresh_token"}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := scrubber.Scrub(test.input)
			if test.containsKey && !strings.Contains(result, "******") {
				t.Errorf("Expected Scrub('%s') to contain '******', got '%s'", test.input, result)
			}
			if result != test.expected {
				t.Errorf("Expected Scrub('%s') to be '%s', got '%s'", test.input, test.expected, result)
			}
		})
	}
}

func TestScrubSensitiveData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"WithSensitiveData", "key1=value1&refresh_token=abc123&key2=value2", "key1=value1&refresh_token=******&key2=value2"},
		{"WithoutSensitiveData", "key1=value1&key2=value2", "key1=value1&key2=value2"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ScrubSensitiveData(test.input, []string{"refresh_token"})
			if result != test.expected {
				t.Errorf("Expected ScrubSensitiveData('%s') to be '%s', got '%s'", test.input, test.expected, result)
			}
		})
	}
}
