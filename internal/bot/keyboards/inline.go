package keyboards

import (
	"fmt"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/util"
	tele "gopkg.in/telebot.v3"
)

// StartMenu creates the main menu keyboard
func StartMenu(challenges []*domain.Challenge, taskCounts, completedCounts map[string]int) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	rows := make([]tele.Row, 0)

	// Add challenge buttons
	for _, c := range challenges {
		total := taskCounts[c.ID]
		completed := completedCounts[c.ID]

		text := fmt.Sprintf("🏆 %s (%d/%d", c.Name, completed, total)
		if total > 0 && completed >= total {
			text += " ✅)"
		} else {
			text += " tasks)"
		}

		btn := menu.Data(text, "open_challenge", c.ID)
		rows = append(rows, menu.Row(btn))
	}

	// Add create/join buttons
	createBtn := menu.Data("🎯 Create Challenge", "create_challenge")
	joinBtn := menu.Data("🚀 Join Challenge", "join_challenge")
	rows = append(rows, menu.Row(createBtn, joinBtn))

	menu.Inline(rows...)
	return menu
}

// CancelOnly creates a keyboard with just a cancel button
func CancelOnly() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	cancelBtn := menu.Data("❌ Cancel", "cancel")
	menu.Inline(menu.Row(cancelBtn))
	return menu
}

// SkipCancel creates a keyboard with skip and cancel buttons
func SkipCancel() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	skipBtn := menu.Data("⏭ Skip", "skip")
	cancelBtn := menu.Data("❌ Cancel", "cancel")
	menu.Inline(menu.Row(skipBtn, cancelBtn))
	return menu
}

// EmojiSelector creates an emoji selection keyboard
func EmojiSelector(usedEmojis []string) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	available := util.FilterAvailableEmojis(usedEmojis)

	rows := make([]tele.Row, 0)
	row := make([]tele.Btn, 0)

	for i, emoji := range available {
		btn := menu.Data(emoji, "select_emoji", emoji)
		row = append(row, btn)

		if len(row) == 5 || i == len(available)-1 {
			rows = append(rows, menu.Row(row...))
			row = make([]tele.Btn, 0)
		}
	}

	cancelBtn := menu.Data("❌ Cancel", "cancel")
	rows = append(rows, menu.Row(cancelBtn))

	menu.Inline(rows...)
	return menu
}

// TaskButton represents a task for keyboard display
type TaskButton struct {
	ID          int64
	OrderNum    int
	Title       string
	IsCompleted bool
	IsCurrent   bool
}

// MainChallengeView creates the main challenge view keyboard with clickable tasks
func MainChallengeView(challengeID string, currentTaskNum int, hasAdmin bool, tasks []TaskButton) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	rows := make([]tele.Row, 0)

	// Task buttons grid (7 columns)
	if len(tasks) > 0 {
		row := make([]tele.Btn, 0, 7)
		for _, task := range tasks {
			var status string
			if task.IsCompleted {
				status = "✅"
			} else {
				status = "⬜"
			}

			text := fmt.Sprintf("%s %d", status, task.OrderNum)
			btn := menu.Data(text, "task_detail", fmt.Sprintf("%d", task.ID))
			row = append(row, btn)

			// Add row when we have 7 buttons
			if len(row) == 7 {
				rows = append(rows, menu.Row(row...))
				row = make([]tele.Btn, 0, 7)
			}
		}
		// Add remaining buttons, padding with empty buttons to fill the row
		if len(row) > 0 {
			for len(row) < 7 {
				row = append(row, menu.Data(" ", "noop"))
			}
			rows = append(rows, menu.Row(row...))
		}
	}

	// Row: Complete current task (if there is one)
	if currentTaskNum > 0 {
		completeBtn := menu.Data(fmt.Sprintf("✅ Complete #%d", currentTaskNum), "complete_current")
		rows = append(rows, menu.Row(completeBtn))
	}

	// Row: Squad Progress, List All
	teamBtn := menu.Data("👥 Squad stats", "team_progress")
	listAllBtn := menu.Data("📋 List all tasks", "list_all_tasks")
	rows = append(rows, menu.Row(teamBtn, listAllBtn))

	// Row: Admin (optional), Settings, Exit
	if hasAdmin {
		adminBtn := menu.Data("🔧 Admin", "admin_panel")
		settingsBtn := menu.Data("⚙️ Settings", "settings")
		exitBtn := menu.Data("🚪 Exit", "exit_challenge")
		rows = append(rows, menu.Row(adminBtn, settingsBtn, exitBtn))
	} else {
		settingsBtn := menu.Data("⚙️ Settings", "settings")
		exitBtn := menu.Data("🚪 Exit", "exit_challenge")
		rows = append(rows, menu.Row(settingsBtn, exitBtn))
	}

	menu.Inline(rows...)
	return menu
}

// TaskList creates inline task buttons
func TaskList(tasks []*domain.Task) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	rows := make([]tele.Row, 0)

	for _, task := range tasks {
		btn := menu.Data(fmt.Sprintf("%d. %s", task.OrderNum, task.Title), "task_detail", fmt.Sprintf("%d", task.ID))
		rows = append(rows, menu.Row(btn))
	}

	menu.Inline(rows...)
	return menu
}

// TaskDetail creates the task detail keyboard
func TaskDetail(taskID int64, isCompleted bool) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}

	var actionBtn tele.Btn
	if isCompleted {
		actionBtn = menu.Data("↩️ Uncomplete", "uncomplete_task", fmt.Sprintf("%d", taskID))
	} else {
		actionBtn = menu.Data("✅ Complete", "complete_task", fmt.Sprintf("%d", taskID))
	}
	backBtn := menu.Data("⬅️ Back", "back_to_main")

	menu.Inline(menu.Row(actionBtn, backBtn))
	return menu
}

// TeamProgress creates the team progress keyboard
func TeamProgress() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	backBtn := menu.Data("⬅️ Back", "back_to_main")
	menu.Inline(menu.Row(backBtn))
	return menu
}

// ShareID creates the share ID keyboard with copy-to-clipboard buttons
func ShareID(challengeID string, botUsername string) *CopyTextKeyboard {
	link := fmt.Sprintf("t.me/%s?start=%s", botUsername, challengeID)
	return NewCopyTextKeyboard(challengeID, link)
}

// CopyText contains the text to copy to clipboard (Bot API 7.1+)
type CopyText struct {
	Text string `json:"text"`
}

// CopyTextInlineButton extends InlineButton with copy_text support
type CopyTextInlineButton struct {
	Text     string    `json:"text"`
	CopyText *CopyText `json:"copy_text,omitempty"`
	Data     string    `json:"callback_data,omitempty"`
}

// CopyTextKeyboard is a custom keyboard that supports copy_text buttons
type CopyTextKeyboard struct {
	InlineKeyboard [][]CopyTextInlineButton `json:"inline_keyboard"`
}

// NewCopyTextKeyboard creates a keyboard with copy-to-clipboard buttons
func NewCopyTextKeyboard(challengeID, link string) *CopyTextKeyboard {
	return &CopyTextKeyboard{
		InlineKeyboard: [][]CopyTextInlineButton{
			{
				{Text: "📋 Copy ID", CopyText: &CopyText{Text: challengeID}},
				{Text: "🔗 Copy Link", CopyText: &CopyText{Text: link}},
			},
			{
				{Text: "⬅️ Back", Data: "\fback_to_main"},
			},
		},
	}
}

// AdminPanel creates the admin panel keyboard
func AdminPanel() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}

	addTaskBtn := menu.Data("➕ Add Task", "add_task")
	editTasksBtn := menu.Data("📋 Edit Tasks", "edit_tasks")
	editNameBtn := menu.Data("✏️ Name", "edit_challenge_name")
	editDescBtn := menu.Data("📝 Description", "edit_challenge_description")
	deleteBtn := menu.Data("🗑 Delete Challenge", "delete_challenge")
	mainBtn := menu.Data("🏠 Main Menu", "back_to_main")

	menu.Inline(
		menu.Row(addTaskBtn, editTasksBtn),
		menu.Row(editNameBtn, editDescBtn),
		menu.Row(deleteBtn, mainBtn),
	)
	return menu
}

// AddTaskDone creates the keyboard after adding a task
func AddTaskDone() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	addAnotherBtn := menu.Data("➕ Add Another Task", "add_task")
	doneBtn := menu.Data("✅ Done Adding Tasks", "back_to_admin")
	menu.Inline(menu.Row(addAnotherBtn), menu.Row(doneBtn))
	return menu
}

// EditTasksList creates the edit tasks list keyboard
func EditTasksList(tasks []*domain.Task) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	rows := make([]tele.Row, 0)

	for _, task := range tasks {
		btn := menu.Data(fmt.Sprintf("%d. %s ✏️", task.OrderNum, task.Title), "edit_task", fmt.Sprintf("%d", task.ID))
		rows = append(rows, menu.Row(btn))
	}

	reorderBtn := menu.Data("🔀 Reorder Tasks", "reorder_tasks")
	backBtn := menu.Data("⬅️ Back", "back_to_admin")
	rows = append(rows, menu.Row(reorderBtn, backBtn))

	menu.Inline(rows...)
	return menu
}

// EditTask creates the edit task keyboard
func EditTask(taskID int64) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}

	editTitleBtn := menu.Data("📝 Edit Title", "edit_task_title", fmt.Sprintf("%d", taskID))
	editImageBtn := menu.Data("📷 Change Image", "edit_task_image", fmt.Sprintf("%d", taskID))
	editDescBtn := menu.Data("📄 Edit Description", "edit_task_description", fmt.Sprintf("%d", taskID))
	deleteBtn := menu.Data("🗑 Delete Task", "delete_task", fmt.Sprintf("%d", taskID))
	backBtn := menu.Data("⬅️ Back", "back_to_tasks")

	menu.Inline(
		menu.Row(editTitleBtn, editImageBtn),
		menu.Row(editDescBtn),
		menu.Row(deleteBtn, backBtn),
	)
	return menu
}

// DeleteTaskConfirm creates delete task confirmation keyboard
func DeleteTaskConfirm(taskID int64) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	confirmBtn := menu.Data("✅ Yes, delete", "confirm_delete_task", fmt.Sprintf("%d", taskID))
	cancelBtn := menu.Data("❌ Cancel", "cancel_delete_task")
	menu.Inline(menu.Row(confirmBtn, cancelBtn))
	return menu
}

// DeleteChallengeConfirm creates delete challenge confirmation keyboard
func DeleteChallengeConfirm() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	confirmBtn := menu.Data("🗑 Yes, delete everything", "confirm_delete_challenge")
	cancelBtn := menu.Data("❌ Cancel", "cancel_delete_challenge")
	menu.Inline(menu.Row(confirmBtn, cancelBtn))
	return menu
}

// ReorderTasksList creates the reorder tasks list keyboard
func ReorderTasksList(tasks []*domain.Task) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	rows := make([]tele.Row, 0)

	for _, task := range tasks {
		btn := menu.Data(fmt.Sprintf("%d. %s", task.OrderNum, task.Title), "reorder_select", fmt.Sprintf("%d", task.ID))
		rows = append(rows, menu.Row(btn))
	}

	randomizeBtn := menu.Data("🎲 Randomize", "randomize_tasks")
	backBtn := menu.Data("⬅️ Back", "back_to_tasks")
	rows = append(rows, menu.Row(randomizeBtn, backBtn))

	menu.Inline(rows...)
	return menu
}

// ReorderPositions creates position selection keyboard for reordering
func ReorderPositions(taskID int64, totalTasks, currentPos int) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	rows := make([]tele.Row, 0)

	for i := 1; i <= totalTasks; i++ {
		var btn tele.Btn
		if i == currentPos {
			btn = menu.Data(fmt.Sprintf("   Current position: %d", i), "noop")
		} else if i < currentPos {
			btn = menu.Data(fmt.Sprintf("⬆️ Move to position %d", i), "reorder_move", fmt.Sprintf("%d", taskID), fmt.Sprintf("%d", i))
		} else {
			btn = menu.Data(fmt.Sprintf("⬇️ Move to position %d", i), "reorder_move", fmt.Sprintf("%d", taskID), fmt.Sprintf("%d", i))
		}
		rows = append(rows, menu.Row(btn))
	}

	cancelBtn := menu.Data("❌ Cancel", "reorder_cancel")
	rows = append(rows, menu.Row(cancelBtn))

	menu.Inline(rows...)
	return menu
}

// ReorderDone creates the keyboard after successful reorder
func ReorderDone() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	moveAnotherBtn := menu.Data("🔀 Move Another", "reorder_tasks")
	doneBtn := menu.Data("⬅️ Done", "back_to_tasks")
	menu.Inline(menu.Row(moveAnotherBtn, doneBtn))
	return menu
}

// Settings creates the settings keyboard
func Settings(notifyEnabled bool, isAdmin bool) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}

	var notifyText string
	if notifyEnabled {
		notifyText = "🔔 Notifications: ON"
	} else {
		notifyText = "🔕 Notifications: OFF"
	}
	notifyBtn := menu.Data(notifyText, "toggle_notifications")

	changeNameBtn := menu.Data("✏️ Change Name", "change_name")
	changeEmojiBtn := menu.Data("😀 Change Emoji", "change_emoji")
	shareBtn := menu.Data("🔗 Share the Challenge", "share_id")
	backBtn := menu.Data("⬅️ Back", "back_to_main")

	rows := []tele.Row{
		menu.Row(notifyBtn),
		menu.Row(changeNameBtn, changeEmojiBtn),
		menu.Row(shareBtn),
	}

	if !isAdmin {
		leaveBtn := menu.Data("🚫 Leave Challenge", "leave_challenge")
		rows = append(rows, menu.Row(leaveBtn, backBtn))
	} else {
		rows = append(rows, menu.Row(backBtn))
	}

	menu.Inline(rows...)
	return menu
}

// LeaveConfirm creates leave challenge confirmation keyboard
func LeaveConfirm() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	confirmBtn := menu.Data("✅ Yes, leave", "confirm_leave")
	cancelBtn := menu.Data("❌ Cancel", "cancel_leave")
	menu.Inline(menu.Row(confirmBtn, cancelBtn))
	return menu
}

// JoinWelcome creates the welcome keyboard after joining
func JoinWelcome(challengeID string) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	startBtn := menu.Data("🚀 Start Challenge", "start_challenge", challengeID)
	menu.Inline(menu.Row(startBtn))
	return menu
}

// Celebration creates the celebration keyboard
func Celebration() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	teamBtn := menu.Data("👥 View Squad", "team_progress")
	mainBtn := menu.Data("🏠 Main Menu", "exit_challenge")
	menu.Inline(menu.Row(teamBtn, mainBtn))
	return menu
}

// BackToAdmin creates a back to admin keyboard
func BackToAdmin() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	backBtn := menu.Data("⬅️ Back to Admin", "back_to_admin")
	menu.Inline(menu.Row(backBtn))
	return menu
}
