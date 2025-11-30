package views

import (
	"strings"
	"testing"
)

func TestRenderTeamProgress(t *testing.T) {
	data := TeamProgressData{
		ChallengeName: "Test Challenge",
		Participants: []*ParticipantProgress{
			{Emoji: "💪", Name: "John", IsAdmin: true, CompletedTasks: 8, TotalTasks: 10},
			{Emoji: "🔥", Name: "Sarah", IsAdmin: false, CompletedTasks: 6, TotalTasks: 10},
			{Emoji: "⭐", Name: "Mike", IsAdmin: false, CompletedTasks: 4, TotalTasks: 10},
		},
	}

	result := RenderTeamProgress(data)

	// Check header
	if !strings.Contains(result, "Squad Progress") {
		t.Error("Should contain 'Squad Progress'")
	}

	// Check admin indicator
	if !strings.Contains(result, "(admin)") {
		t.Error("Should show admin indicator")
	}

	// Check progress bars exist
	if !strings.Contains(result, "█") {
		t.Error("Should contain progress bar filled portion")
	}
	if !strings.Contains(result, "░") {
		t.Error("Should contain progress bar empty portion")
	}

	// Check percentages
	if !strings.Contains(result, "80%") {
		t.Error("Should show 80% for John")
	}
	if !strings.Contains(result, "60%") {
		t.Error("Should show 60% for Sarah")
	}
}

func TestRenderProgressBar(t *testing.T) {
	tests := []struct {
		pct      int
		wantFill int
	}{
		{0, 0},
		{10, 1},
		{50, 5},
		{100, 10},
	}

	for _, tt := range tests {
		bar := renderProgressBar(tt.pct)
		filled := strings.Count(bar, "█")
		if filled != tt.wantFill {
			t.Errorf("renderProgressBar(%d) has %d filled, want %d", tt.pct, filled, tt.wantFill)
		}
		// Count runes (characters), not bytes
		runeCount := len([]rune(bar))
		if runeCount != 10 {
			t.Errorf("renderProgressBar(%d) rune length = %d, want 10", tt.pct, runeCount)
		}
	}
}

func TestTeamProgressSorting(t *testing.T) {
	data := TeamProgressData{
		ChallengeName: "Test",
		Participants: []*ParticipantProgress{
			{Emoji: "⭐", Name: "Low", CompletedTasks: 2, TotalTasks: 10},
			{Emoji: "💪", Name: "High", CompletedTasks: 8, TotalTasks: 10},
			{Emoji: "🔥", Name: "Mid", CompletedTasks: 5, TotalTasks: 10},
		},
	}

	result := RenderTeamProgress(data)

	// High should appear before Mid, Mid before Low
	highIdx := strings.Index(result, "High")
	midIdx := strings.Index(result, "Mid")
	lowIdx := strings.Index(result, "Low")

	if highIdx > midIdx || midIdx > lowIdx {
		t.Error("Participants should be sorted by completion percentage descending")
	}
}
