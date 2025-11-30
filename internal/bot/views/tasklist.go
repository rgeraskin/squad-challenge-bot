package views

import (
	"fmt"
	"strings"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

// TaskListData holds data for rendering task list
type TaskListData struct {
	ChallengeName        string
	ChallengeDescription string
	TotalTasks           int
	CompletedTasks       int
	ParticipantCount     int
	Tasks                []*domain.Task
	CompletedTaskIDs     map[int64]bool
	ParticipantEmojis    map[int64][]string // task ID -> list of emojis of participants on that task
	CurrentUserEmoji     string
	CurrentTaskNum       int
	HideFutureTasks      bool // hide task names after current task
}

// RenderTaskList renders the main challenge view with task list
func RenderTaskList(data TaskListData) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("🏆 %s\n", data.ChallengeName))
	if data.ChallengeDescription != "" {
		sb.WriteString(fmt.Sprintf("%s\n", data.ChallengeDescription))
	}
	sb.WriteString(fmt.Sprintf("Progress: %d/%d tasks • %d members\n", data.CompletedTasks, data.TotalTasks, data.ParticipantCount))
	sb.WriteString("─────────────────────────\n\n")

	if len(data.Tasks) == 0 {
		sb.WriteString("📭 No tasks yet\n")
		sb.WriteString("Waiting for admin to add tasks...\n")
	} else {
		// Calculate visible range: 2 prev + current + 2 next (max 5)
		startIdx, endIdx := CalculateVisibleRange(data.CurrentTaskNum, len(data.Tasks))

		// Show count of tasks before the visible range
		if startIdx > 0 {
			sb.WriteString(fmt.Sprintf("↑ %d more task(s)\n", startIdx))
		}

		for i := startIdx; i <= endIdx && i < len(data.Tasks); i++ {
			task := data.Tasks[i]
			isCompleted := data.CompletedTaskIDs[task.ID]

			// Status emoji
			var status string
			if isCompleted {
				status = "✅"
			} else {
				status = "⬜"
			}

			// Check if task should be hidden
			isHidden := data.HideFutureTasks && task.OrderNum > data.CurrentTaskNum

			// Task line
			var line string
			if isHidden {
				line = fmt.Sprintf("%s %d. <tg-spoiler>🔒 Complete previous tasks to unlock</tg-spoiler>", status, task.OrderNum)
			} else {
				line = fmt.Sprintf("%s %d. %s", status, task.OrderNum, task.Title)

				// Add participant emojis (only for visible tasks)
				if emojis, ok := data.ParticipantEmojis[task.ID]; ok && len(emojis) > 0 {
					if len(emojis) <= 4 {
						line += "    " + strings.Join(emojis, "")
					} else {
						line += "    " + strings.Join(emojis[:4], "") + fmt.Sprintf(" +%d", len(emojis)-4)
					}
				}

				// Mark current user's position
				if task.OrderNum == data.CurrentTaskNum && data.CurrentUserEmoji != "" {
					line += "    ← YOU"
				}
			}

			sb.WriteString(line + "\n")
		}

		// Show count of tasks after the visible range
		if endIdx < len(data.Tasks)-1 {
			remaining := len(data.Tasks) - 1 - endIdx
			sb.WriteString(fmt.Sprintf("↓ %d more task(s)\n", remaining))
		}
	}

	return sb.String()
}

// CalculateVisibleRange returns start and end indices for visible tasks
// Shows 2 previous + current + 2 next (max 5 visible)
func CalculateVisibleRange(currentTaskNum, totalTasks int) (int, int) {
	if totalTasks == 0 {
		return 0, 0
	}

	// currentTaskNum is 1-based, convert to 0-based index
	currentIdx := currentTaskNum - 1
	if currentIdx < 0 {
		currentIdx = 0
	}
	if currentIdx >= totalTasks {
		currentIdx = totalTasks - 1
	}

	// Calculate range: 2 before + current + 2 after
	startIdx := currentIdx - 2
	if startIdx < 0 {
		startIdx = 0
	}

	endIdx := currentIdx + 2
	if endIdx >= totalTasks {
		endIdx = totalTasks - 1
	}

	return startIdx, endIdx
}

// AllTasksData holds data for rendering the full task list
type AllTasksData struct {
	ChallengeName    string
	Tasks            []*domain.Task
	CompletedTaskIDs map[int64]bool
	HideFutureTasks  bool
	CurrentTaskNum   int
}

// RenderAllTasks renders the full list of all tasks
func RenderAllTasks(data AllTasksData) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("📋 %s - All Tasks\n", data.ChallengeName))
	sb.WriteString("─────────────────────────\n\n")

	if len(data.Tasks) == 0 {
		sb.WriteString("📭 No tasks yet\n")
	} else {
		for _, task := range data.Tasks {
			isCompleted := data.CompletedTaskIDs[task.ID]

			var status string
			if isCompleted {
				status = "✅"
			} else {
				status = "⬜"
			}

			// Check if task should be hidden
			isHidden := data.HideFutureTasks && task.OrderNum > data.CurrentTaskNum

			var line string
			if isHidden {
				line = fmt.Sprintf("%s %d. <tg-spoiler>🔒 Complete previous tasks to unlock</tg-spoiler>", status, task.OrderNum)
			} else {
				line = fmt.Sprintf("%s %d. %s", status, task.OrderNum, task.Title)
			}
			sb.WriteString(line + "\n")
		}
	}

	return sb.String()
}
