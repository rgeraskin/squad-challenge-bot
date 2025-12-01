package handlers

import (
	"fmt"
	"strconv"

	"github.com/rgeraskin/squad-challenge-bot/internal/bot/keyboards"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/service"
	tele "gopkg.in/telebot.v3"
)

// handleAddTask starts the add task flow
func (h *Handler) handleAddTask(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	// Check task limit
	count, _ := h.task.CountByChallengeID(challengeID)
	if count >= service.MaxTasksPerChallenge {
		return h.sendError(c, "📋 Maxed out at 50 tasks!")
	}

	h.state.SetState(userID, domain.StateAwaitingTaskTitle)
	return c.Send("📝 What's the task called?", keyboards.CancelOnly())
}

// processTaskTitle processes task title input
func (h *Handler) processTaskTitle(c tele.Context, title string) error {
	userID := c.Sender().ID

	if len(title) == 0 || len(title) > 100 {
		return c.Send("😅 Keep it between 1-100 characters:", keyboards.CancelOnly())
	}

	tempData := map[string]interface{}{
		"task_title": title,
	}
	h.state.SetStateWithData(userID, domain.StateAwaitingTaskImage, tempData)

	return c.Send("🖼 Got a picture for this task? (or skip it)", keyboards.SkipCancel())
}

// processTaskImage processes task image upload
func (h *Handler) processTaskImage(c tele.Context, fileID string) error {
	userID := c.Sender().ID

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	tempData["image_file_id"] = fileID
	h.state.SetStateWithData(userID, domain.StateAwaitingTaskDescription, tempData)

	return c.Send("📝 Add some details? (or skip it)", keyboards.SkipCancel())
}

// skipTaskImage skips the task image
func (h *Handler) skipTaskImage(c tele.Context) error {
	userID := c.Sender().ID

	h.state.SetState(userID, domain.StateAwaitingTaskDescription)
	return c.Send("📝 Add some details? (or skip it)", keyboards.SkipCancel())
}

// processTaskDescription processes task description input
func (h *Handler) processTaskDescription(c tele.Context, description string) error {
	if len(description) > 800 {
		return c.Send("😅 That's a bit long! Keep it under 800 characters:", keyboards.SkipCancel())
	}

	return h.createTask(c, description)
}

// skipTaskDescription skips the task description and creates the task
func (h *Handler) skipTaskDescription(c tele.Context) error {
	return h.createTask(c, "")
}

// createTask creates the task with collected data
func (h *Handler) createTask(c tele.Context, description string) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)

	title := tempData["task_title"].(string)
	var imageFileID string
	if img, ok := tempData["image_file_id"]; ok {
		imageFileID = img.(string)
	}

	task, err := h.task.Create(challengeID, title, description, imageFileID)
	if err != nil {
		h.state.ResetKeepChallenge(userID)
		if err == service.ErrMaxTasksReached {
			return h.sendError(c, "📋 Maxed out at 50 tasks!")
		}
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	h.state.ResetKeepChallenge(userID)

	msg := fmt.Sprintf("✅ Task #%d added: \"%s\"", task.OrderNum, task.Title)
	return c.Send(msg, keyboards.AddTaskDone())
}

// handleEditTasks shows the edit tasks list
func (h *Handler) handleEditTasks(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	tasks, err := h.task.GetByChallengeID(challengeID)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	if len(tasks) == 0 {
		return c.Send("📭 No tasks yet — add some first!", keyboards.BackToAdmin())
	}

	msg := "📋 <i>Edit tasks</i>\n\nTap one to edit"
	return c.Send(msg, keyboards.EditTasksList(tasks), tele.ModeHTML)
}

// handleEditTask shows the edit task menu
func (h *Handler) handleEditTask(c tele.Context, taskIDStr string) error {
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	task, err := h.task.GetByID(taskID)
	if err != nil {
		return h.sendError(c, "🤔 Can't find that task.")
	}

	// Send image first if exists
	if task.ImageFileID != "" {
		photo := &tele.Photo{File: tele.File{FileID: task.ImageFileID}}
		c.Send(photo)
	}

	msg := fmt.Sprintf("✏️ Task #%d: <b>%s</b>", task.OrderNum, task.Title)
	if task.Description != "" {
		msg += fmt.Sprintf("\n\n<i>%s</i>", task.Description)
	}
	return c.Send(msg, keyboards.EditTask(taskID), tele.ModeHTML)
}

// handleEditTaskTitle starts editing task title
func (h *Handler) handleEditTaskTitle(c tele.Context, taskIDStr string) error {
	userID := c.Sender().ID
	taskID, _ := strconv.ParseInt(taskIDStr, 10, 64)

	tempData := map[string]interface{}{
		"task_id": taskID,
	}
	h.state.SetStateWithData(userID, domain.StateAwaitingEditTitle, tempData)

	return c.Send("✏️ What's the new title?", keyboards.CancelOnly())
}

// processEditTitle processes new task title
func (h *Handler) processEditTitle(c tele.Context, title string) error {
	userID := c.Sender().ID

	if len(title) == 0 || len(title) > 100 {
		return c.Send("😅 Keep it between 1-100 characters:", keyboards.CancelOnly())
	}

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	taskID := int64(tempData["task_id"].(float64))

	task, err := h.task.GetByID(taskID)
	if err != nil {
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "🤔 Can't find that task.")
	}

	task.Title = title
	if err := h.task.Update(task); err != nil {
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	h.state.ResetKeepChallenge(userID)
	c.Send(fmt.Sprintf("✅ Done! Now it's \"%s\"", title))
	return h.handleEditTasks(c)
}

// handleEditTaskDescription starts editing task description
func (h *Handler) handleEditTaskDescription(c tele.Context, taskIDStr string) error {
	userID := c.Sender().ID
	taskID, _ := strconv.ParseInt(taskIDStr, 10, 64)

	tempData := map[string]interface{}{
		"task_id": taskID,
	}
	h.state.SetStateWithData(userID, domain.StateAwaitingEditDescription, tempData)

	return c.Send(
		"📝 What's the new description?",
		keyboards.CancelOnly(),
	)
}

// processEditDescription processes new task description
func (h *Handler) processEditDescription(c tele.Context, description string) error {
	userID := c.Sender().ID

	if len(description) > 800 {
		return c.Send("😅 That's a bit long! Keep it under 800 characters:", keyboards.CancelOnly())
	}

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	taskID := int64(tempData["task_id"].(float64))

	task, err := h.task.GetByID(taskID)
	if err != nil {
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "🤔 Can't find that task.")
	}

	task.Description = description
	if err := h.task.Update(task); err != nil {
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	h.state.ResetKeepChallenge(userID)
	c.Send("✅ Description updated!")
	return h.handleEditTasks(c)
}

// handleEditTaskImage starts editing task image
func (h *Handler) handleEditTaskImage(c tele.Context, taskIDStr string) error {
	userID := c.Sender().ID
	taskID, _ := strconv.ParseInt(taskIDStr, 10, 64)

	tempData := map[string]interface{}{
		"task_id": taskID,
	}
	h.state.SetStateWithData(userID, domain.StateAwaitingEditImage, tempData)

	return c.Send("🖼 Send the new image", keyboards.CancelOnly())
}

// processEditImage processes new task image
func (h *Handler) processEditImage(c tele.Context, fileID string) error {
	userID := c.Sender().ID

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	taskID := int64(tempData["task_id"].(float64))

	task, err := h.task.GetByID(taskID)
	if err != nil {
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "🤔 Can't find that task.")
	}

	task.ImageFileID = fileID
	if err := h.task.Update(task); err != nil {
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	h.state.ResetKeepChallenge(userID)
	c.Send("✅ New image saved!")
	return h.handleEditTasks(c)
}

// handleDeleteTask shows delete task confirmation
func (h *Handler) handleDeleteTask(c tele.Context, taskIDStr string) error {
	taskID, _ := strconv.ParseInt(taskIDStr, 10, 64)

	task, err := h.task.GetByID(taskID)
	if err != nil {
		return h.sendError(c, "🤔 Can't find that task.")
	}

	msg := fmt.Sprintf("🗑 Delete \"%s\"?\n\nEveryone's progress on this will be gone!", task.Title)
	return c.Send(msg, keyboards.DeleteTaskConfirm(taskID))
}

// handleConfirmDeleteTask confirms task deletion
func (h *Handler) handleConfirmDeleteTask(c tele.Context, taskIDStr string) error {
	userID := c.Sender().ID
	taskID, _ := strconv.ParseInt(taskIDStr, 10, 64)

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	if err := h.task.Delete(taskID, challengeID); err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	// Check if any participants now completed all tasks due to this deletion
	go h.checkCompletionsAfterTaskDelete(challengeID, userID)

	c.Send("✅ Gone! Task deleted.")
	return h.handleEditTasks(c)
}

// checkCompletionsAfterTaskDelete checks if anyone completed the challenge after task deletion
func (h *Handler) checkCompletionsAfterTaskDelete(challengeID string, adminUserID int64) {
	tasks, err := h.task.GetByChallengeID(challengeID)
	if err != nil || len(tasks) == 0 {
		return
	}

	challenge, err := h.challenge.GetByID(challengeID)
	if err != nil {
		return
	}

	participants, err := h.participant.GetByChallengeID(challengeID)
	if err != nil {
		return
	}

	totalTasks := len(tasks)

	for _, p := range participants {
		completed, _ := h.completion.CountByParticipantID(p.ID)
		if completed >= totalTasks {
			// This participant has now completed all tasks
			// Notify them directly
			h.notification.NotifyUserChallengeCompleted(p.TelegramID, challenge.Name)

			// Notify others (except admin who just deleted the task)
			h.notification.NotifyChallengeCompleted(
				challengeID,
				p.Emoji,
				p.DisplayName,
				adminUserID,
			)
		}
	}
}

// handleReorderTasks shows the reorder tasks list
func (h *Handler) handleReorderTasks(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	tasks, err := h.task.GetByChallengeID(challengeID)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	if len(tasks) < 2 {
		return c.Send("🤷 Need at least 2 tasks to shuffle around!", keyboards.BackToAdmin())
	}

	challenge, _ := h.challenge.GetByID(challengeID)
	msg := fmt.Sprintf("🔀 Reorder — %s\n\nTap the task you want to move:", challenge.Name)
	return c.Send(msg, keyboards.ReorderTasksList(tasks))
}

// handleReorderSelect selects a task to move
func (h *Handler) handleReorderSelect(c tele.Context, taskIDStr string) error {
	userID := c.Sender().ID
	taskID, _ := strconv.ParseInt(taskIDStr, 10, 64)

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	task, err := h.task.GetByID(taskID)
	if err != nil {
		return h.sendError(c, "🤔 Can't find that task.")
	}

	tasks, _ := h.task.GetByChallengeID(challengeID)

	msg := fmt.Sprintf("🔀 Moving \"%s\"\n\nWhere should it go?", task.Title)
	return c.Send(msg, keyboards.ReorderPositions(taskID, len(tasks), task.OrderNum))
}

// handleReorderMove moves task to new position
func (h *Handler) handleReorderMove(c tele.Context, taskIDStr, positionStr string) error {
	userID := c.Sender().ID
	taskID, _ := strconv.ParseInt(taskIDStr, 10, 64)
	newPosition, _ := strconv.Atoi(positionStr)

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	if err := h.task.MoveTask(taskID, challengeID, newPosition); err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	// Show new order
	tasks, _ := h.task.GetByChallengeID(challengeID)
	msg := "✅ Done! Here's the new order:\n\n"
	for _, t := range tasks {
		msg += fmt.Sprintf("%d. %s\n", t.OrderNum, t.Title)
	}

	return c.Send(msg, keyboards.ReorderDone())
}

// handleRandomizeTasks randomizes the order of all tasks
func (h *Handler) handleRandomizeTasks(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	if err := h.task.RandomizeOrder(challengeID); err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	// Show reorder view with updated order
	return h.handleReorderTasks(c)
}
