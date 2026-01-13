package utils

import "testing"

func TestExtractAddressFromTopic(t *testing.T) {
	tests := []struct {
		name     string
		topic    interface{}
		expected string
	}{
		{
			"full topic",
			"0x00000000000000000000000076bf0f26a080ec0dbe68c842c99b7f6a7cb116db",
			"0x76bf0f26a080ec0dbe68c842c99b7f6a7cb116db",
		},
		{
			"short input",
			"0x1234",
			"0x1234",
		},
		{
			"non-string",
			123,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractAddressFromTopic(tt.topic)
			if got != tt.expected {
				t.Errorf("ExtractAddressFromTopic() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTruncateAddress(t *testing.T) {
	tests := []struct {
		addr     string
		expected string
	}{
		{"0x76bf0f26a080ec0dbe68c842c99b7f6a7cb116db", "0x76bf...16db"},
		{"short", "short"},
		{"0x1234567890", "0x1234...7890"},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			got := TruncateAddress(tt.addr)
			if got != tt.expected {
				t.Errorf("TruncateAddress() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsZeroAddress(t *testing.T) {
	tests := []struct {
		addr     string
		expected bool
	}{
		{ZeroAddress, true},
		{"0x0000000000000000000000000000000000000000", true},
		{"0x76bf0f26a080ec0dbe68c842c99b7f6a7cb116db", false},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			got := IsZeroAddress(tt.addr)
			if got != tt.expected {
				t.Errorf("IsZeroAddress() = %v, want %v", got, tt.expected)
			}
		})
	}
}
