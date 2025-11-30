package views

import (
	"strings"
	"testing"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

func TestRenderTaskList(t *testing.T) {
	tasks := []*domain.Task{
		{ID: 1, OrderNum: 1, Title: "Task 1"},
		{ID: 2, OrderNum: 2, Title: "Task 2"},
		{ID: 3, OrderNum: 3, Title: "Task 3"},
	}

	data := TaskListData{
		ChallengeName:     "Test Challenge",
		TotalTasks:        3,
		CompletedTasks:    1,
		ParticipantCount:  2,
		Tasks:             tasks,
		CompletedTaskIDs:  map[int64]bool{1: true},
		ParticipantEmojis: map[int64][]string{2: {"💪", "🔥"}},
		CurrentUserEmoji:  "💪",
		CurrentTaskNum:    2,
	}

	result := RenderTaskList(data)

	// Check header
	if !strings.Contains(result, "Test Challenge") {
		t.Error("Should contain challenge name")
	}
	if !strings.Contains(result, "1/3 tasks") {
		t.Error("Should contain progress")
	}
	if !strings.Contains(result, "2 members") {
		t.Error("Should contain member count")
	}

	// Check completed task has ✅
	if !strings.Contains(result, "✅ 1. Task 1") {
		t.Error("Completed task should have ✅")
	}

	// Check uncompleted task has ⬜
	if !strings.Contains(result, "⬜ 2. Task 2") {
		t.Error("Uncompleted task should have ⬜")
	}

	// Check participant emojis
	if !strings.Contains(result, "💪🔥") {
		t.Error("Should show participant emojis")
	}

	// Check YOU indicator
	if !strings.Contains(result, "← YOU") {
		t.Error("Should show YOU indicator on current task")
	}
}

func TestRenderTaskList_Empty(t *testing.T) {
	data := TaskListData{
		ChallengeName:    "Empty Challenge",
		TotalTasks:       0,
		CompletedTasks:   0,
		ParticipantCount: 1,
		Tasks:            []*domain.Task{},
	}

	result := RenderTaskList(data)

	if !strings.Contains(result, "No tasks yet") {
		t.Error("Empty challenge should show 'No tasks yet'")
	}
}

func TestCalculateVisibleRange(t *testing.T) {
	tests := []struct {
		name           string
		currentTaskNum int
		totalTasks     int
		wantStart      int
		wantEnd        int
	}{
		{"first task of 10", 1, 10, 0, 2},   // 2 prev + current + 2 next, clamped to start
		{"third task of 10", 3, 10, 0, 4},   // idx=2, start=0, end=4
		{"fifth task of 10", 5, 10, 2, 6},   // idx=4, start=2, end=6
		{"last task of 10", 10, 10, 7, 9},   // idx=9, start=7, end=9 (clamped)
		{"small challenge", 1, 3, 0, 2},
		{"empty", 0, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := CalculateVisibleRange(tt.currentTaskNum, tt.totalTasks)
			if start != tt.wantStart {
				t.Errorf("start = %d, want %d", start, tt.wantStart)
			}
			if end != tt.wantEnd {
				t.Errorf("end = %d, want %d", end, tt.wantEnd)
			}
		})
	}
}

func TestRenderTaskList_EmojiOverflow(t *testing.T) {
	tasks := []*domain.Task{
		{ID: 1, OrderNum: 1, Title: "Task 1"},
	}

	// More than 4 emojis
	data := TaskListData{
		ChallengeName:     "Test",
		TotalTasks:        1,
		CompletedTasks:    0,
		ParticipantCount:  6,
		Tasks:             tasks,
		CompletedTaskIDs:  map[int64]bool{},
		ParticipantEmojis: map[int64][]string{1: {"💪", "🔥", "⭐", "🎯", "🚀", "💎"}},
		CurrentTaskNum:    1,
	}

	result := RenderTaskList(data)

	// Should show first 4 emojis + overflow count
	if !strings.Contains(result, "+2") {
		t.Error("Should show +2 for emoji overflow")
	}
}
