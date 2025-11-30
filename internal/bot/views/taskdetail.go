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

	sb.WriteString(fmt.Sprintf("📌 Task #%d: <b>%s</b>\n\n", data.Task.OrderNum, data.Task.Title))

	// Description
	if data.Task.Description != "" {
		sb.WriteString("<i>" + data.Task.Description + "</i>\n\n")
	}

	// Status
	if data.IsCompleted {
		sb.WriteString("🎉 You did it!\n")
	} else {
		sb.WriteString("⬜ Not done yet\n")
	}

	// Completed by
	if len(data.CompletedBy) > 0 {
		sb.WriteString("\n<b>Done:</b>\n")
		names := make([]string, len(data.CompletedBy))
		for i, p := range data.CompletedBy {
			names[i] = fmt.Sprintf("%s %s", p.Emoji, p.Name)
		}
		sb.WriteString(strings.Join(names, " • ") + "\n")
	}

	// Not yet
	if len(data.NotYet) > 0 {
		sb.WriteString("\n<b>Still working on it:</b>\n")
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

	sb.WriteString(fmt.Sprintf("🔒 Task #%d\n\n", data.TaskOrderNum))
	sb.WriteString("Whoa there! This one's still locked!\n\n")
	sb.WriteString("<i>Finish your current tasks first to unlock it.</i>\n\n")
	sb.WriteString(fmt.Sprintf("👉 You're on <b>Task #%d</b> right now\n", data.CurrentTaskNum))

	return sb.String()
}
