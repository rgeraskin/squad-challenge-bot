package util

import (
	"testing"
)

func TestIsValidEmoji(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"simple emoji", "💪", true},
		{"fire emoji", "🔥", true},
		{"star emoji", "⭐", true},
		{"flag emoji", "🇺🇸", true},
		{"empty string", "", false},
		{"letter", "A", false},
		{"number", "1", false},
		{"word", "hello", false},
		{"mixed", "a💪", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidEmoji(tt.input)
			if got != tt.want {
				t.Errorf("IsValidEmoji(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFilterAvailableEmojis(t *testing.T) {
	taken := []string{"💪", "🔥"}
	available := FilterAvailableEmojis(taken)

	// Check taken emojis are not in available
	for _, emoji := range taken {
		for _, avail := range available {
			if emoji == avail {
				t.Errorf("FilterAvailableEmojis() should not include taken emoji %s", emoji)
			}
		}
	}

	// Check we still have some available
	if len(available) == 0 {
		t.Error("FilterAvailableEmojis() returned empty list")
	}

	// Check ⭐ is available (not taken)
	found := false
	for _, avail := range available {
		if avail == "⭐" {
			found = true
			break
		}
	}
	if !found {
		t.Error("FilterAvailableEmojis() should include ⭐")
	}
}
