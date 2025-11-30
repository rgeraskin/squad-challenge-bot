package util

import (
	"regexp"
	"unicode"
)

// SuggestedEmojis is a list of suggested emojis for participants
var SuggestedEmojis = []string{
	"💪", "🔥", "⭐", "🎯", "🚀",
	"💎", "🌟", "⚡", "🏆", "🎮",
	"🦁", "🐯", "🦊", "🐺", "🦅",
	"🌈", "☀️", "🌙", "❤️", "💜",
}

// emojiRegex matches most common emojis
var emojiRegex = regexp.MustCompile(`^[\p{So}\p{Sk}]+$`)

// IsValidEmoji checks if the string is a single emoji
func IsValidEmoji(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Count grapheme clusters (visible characters)
	count := 0
	for _, r := range s {
		if unicode.Is(unicode.So, r) || unicode.Is(unicode.Sk, r) || unicode.Is(unicode.Sm, r) {
			count++
		} else if !unicode.Is(unicode.Mn, r) && !unicode.Is(unicode.Me, r) && r != 0xFE0F && r != 0x200D {
			// Allow modifiers, combining marks, variation selectors, and zero-width joiners
			return false
		}
	}

	// Should have at least one emoji symbol
	return count >= 1 && count <= 3 // Allow compound emojis like flags
}

// FilterAvailableEmojis returns emojis not already taken by participants
func FilterAvailableEmojis(taken []string) []string {
	takenMap := make(map[string]bool)
	for _, e := range taken {
		takenMap[e] = true
	}

	available := make([]string, 0)
	for _, e := range SuggestedEmojis {
		if !takenMap[e] {
			available = append(available, e)
		}
	}
	return available
}
