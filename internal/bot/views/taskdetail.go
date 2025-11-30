package views

import (
	"fmt"
	"strings"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

// TaskDetailData holds data for rendering task detail view
type TaskDetailData struct {
	Task        *domain.Task
	IsCompleted bool
	CompletedBy []*ParticipantStatus
	NotYet      []*ParticipantStatus
}

// ParticipantStatus holds participant display info
type ParticipantStatus struct {
	Emoji string
	Name  string
}

// RenderTaskDetail renders the task detail view
func RenderTaskDetail(data TaskDetailData) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Task #%d: %s\n", data.Task.OrderNum, data.Task.Title))
	sb.WriteString("─────────────────────────\n\n")

	// Status
	if data.IsCompleted {
		sb.WriteString("Your status: ✅ Completed\n")
	} else {
		sb.WriteString("Your status: ⬜ Not completed\n")
	}

	// Description
	if data.Task.Description != "" {
		sb.WriteString("\nDescription:\n")
		sb.WriteString(data.Task.Description + "\n")
	}

	// Completed by
	if len(data.CompletedBy) > 0 {
		sb.WriteString("\nCompleted by:\n")
		names := make([]string, len(data.CompletedBy))
		for i, p := range data.CompletedBy {
			names[i] = fmt.Sprintf("%s %s", p.Emoji, p.Name)
		}
		sb.WriteString(strings.Join(names, " • ") + "\n")
	}

	// Not yet
	if len(data.NotYet) > 0 {
		sb.WriteString("\nNot yet:\n")
		names := make([]string, len(data.NotYet))
		for i, p := range data.NotYet {
			names[i] = fmt.Sprintf("%s %s", p.Emoji, p.Name)
		}
		sb.WriteString(strings.Join(names, " • ") + "\n")
	}

	return sb.String()
}

// HiddenTaskDetailData holds data for rendering hidden task detail view
type HiddenTaskDetailData struct {
	TaskOrderNum   int
	CurrentTaskNum int
}

// RenderHiddenTaskDetail renders the hidden task detail view
func RenderHiddenTaskDetail(data HiddenTaskDetailData) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("🔒 Task #%d: Hidden\n", data.TaskOrderNum))
	sb.WriteString("─────────────────────────\n\n")
	sb.WriteString("This task is not yet unlocked.\n\n")
	sb.WriteString("Complete your previous tasks first to reveal this task's details.\n\n")
	sb.WriteString(fmt.Sprintf("Your current task: Task #%d\n", data.CurrentTaskNum))

	return sb.String()
}
