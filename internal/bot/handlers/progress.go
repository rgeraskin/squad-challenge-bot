package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/rgeraskin/squad-challenge-bot/internal/bot/keyboards"
	"github.com/rgeraskin/squad-challenge-bot/internal/bot/views"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	tele "gopkg.in/telebot.v3"
)

// handleCompleteTask completes a task
func (h *Handler) handleCompleteTask(c tele.Context, taskIDStr string) error {
	userID := c.Sender().ID
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	participant, err := h.participant.GetByChallengeAndUser(challengeID, userID)
	if err != nil || participant == nil {
		return h.sendError(c, "❌ You're not a participant of this challenge.")
	}

	task, err := h.task.GetByID(taskID)
	if err != nil {
		return h.sendError(c, "⚠️ Task not found.")
	}

	_, err = h.completion.Complete(taskID, participant.ID)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	// Check if all tasks completed
	tasks, _ := h.task.GetByChallengeID(challengeID)
	allCompleted, _ := h.completion.IsAllCompleted(participant.ID, len(tasks))

	// Notify others
	go h.notification.NotifyTaskCompleted(challengeID, participant.Emoji, participant.DisplayName, task.Title, userID)

	if allCompleted {
		// Notify challenge completion
		go h.notification.NotifyChallengeCompleted(challengeID, participant.Emoji, participant.DisplayName, userID)
		return h.showCelebration(c, challengeID, participant)
	}

	return h.showMainChallengeView(c, challengeID)
}

// handleCompleteCurrent completes the current task
func (h *Handler) handleCompleteCurrent(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	participant, err := h.participant.GetByChallengeAndUser(challengeID, userID)
	if err != nil || participant == nil {
		return h.sendError(c, "❌ You're not a participant of this challenge.")
	}

	tasks, err := h.task.GetByChallengeID(challengeID)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	currentTaskNum := h.completion.GetCurrentTaskNum(participant.ID, tasks)
	if currentTaskNum == 0 {
		return h.sendError(c, "ℹ️ You've already completed all tasks!")
	}

	// Find task ID for current task number
	var taskID int64
	for _, t := range tasks {
		if t.OrderNum == currentTaskNum {
			taskID = t.ID
			break
		}
	}

	return h.handleCompleteTask(c, fmt.Sprintf("%d", taskID))
}

// handleUncompleteTask uncompletes a task
func (h *Handler) handleUncompleteTask(c tele.Context, taskIDStr string) error {
	userID := c.Sender().ID
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	participant, err := h.participant.GetByChallengeAndUser(challengeID, userID)
	if err != nil || participant == nil {
		return h.sendError(c, "❌ You're not a participant of this challenge.")
	}

	err = h.completion.Uncomplete(taskID, participant.ID)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	return h.showTaskDetail(c, taskIDStr)
}

// showTaskDetail shows the task detail view
func (h *Handler) showTaskDetail(c tele.Context, taskIDStr string) error {
	userID := c.Sender().ID
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	task, err := h.task.GetByID(taskID)
	if err != nil {
		return h.sendError(c, "⚠️ Task not found.")
	}

	participant, err := h.participant.GetByChallengeAndUser(challengeID, userID)
	if err != nil || participant == nil {
		return h.sendError(c, "❌ You're not a participant of this challenge.")
	}

	isCompleted, _ := h.completion.IsCompleted(taskID, participant.ID)

	// Get completion status for all participants
	participants, _ := h.participant.GetByChallengeID(challengeID)
	completions, _ := h.completion.GetCompletionsByTaskID(taskID)

	completedSet := make(map[int64]bool)
	for _, comp := range completions {
		completedSet[comp.ParticipantID] = true
	}

	var completedBy, notYet []*views.ParticipantStatus
	for _, p := range participants {
		status := &views.ParticipantStatus{
			Emoji: p.Emoji,
			Name:  p.DisplayName,
		}
		if completedSet[p.ID] {
			completedBy = append(completedBy, status)
		} else {
			notYet = append(notYet, status)
		}
	}

	data := views.TaskDetailData{
		Task:        task,
		IsCompleted: isCompleted,
		CompletedBy: completedBy,
		NotYet:      notYet,
	}

	text := views.RenderTaskDetail(data)

	// Send image if exists
	if task.ImageFileID != "" {
		photo := &tele.Photo{File: tele.File{FileID: task.ImageFileID}}
		c.Send(photo)
	}

	return c.Send(text, keyboards.TaskDetail(taskID, isCompleted))
}

// showTeamProgress shows the team progress view
func (h *Handler) showTeamProgress(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	challenge, err := h.challenge.GetByID(challengeID)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	tasks, _ := h.task.GetByChallengeID(challengeID)
	totalTasks := len(tasks)

	participants, err := h.participant.GetByChallengeID(challengeID)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	var progressList []*views.ParticipantProgress
	for _, p := range participants {
		completed, _ := h.completion.CountByParticipantID(p.ID)
		progressList = append(progressList, &views.ParticipantProgress{
			Emoji:          p.Emoji,
			Name:           p.DisplayName,
			IsAdmin:        challenge.CreatorID == p.TelegramID,
			CompletedTasks: completed,
			TotalTasks:     totalTasks,
		})
	}

	data := views.TeamProgressData{
		ChallengeName: challenge.Name,
		Participants:  progressList,
	}

	text := views.RenderTeamProgress(data)
	return c.Send(text, keyboards.TeamProgress())
}

// showAllTasks shows the full list of all tasks
func (h *Handler) showAllTasks(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	challenge, err := h.challenge.GetByID(challengeID)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	tasks, err := h.task.GetByChallengeID(challengeID)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	participant, err := h.participant.GetByChallengeAndUser(challengeID, userID)
	if err != nil || participant == nil {
		return h.sendError(c, "❌ You're not a participant of this challenge.")
	}

	// Get completion data
	completedIDs, _ := h.completion.GetCompletedTaskIDs(participant.ID)
	completedSet := make(map[int64]bool)
	for _, id := range completedIDs {
		completedSet[id] = true
	}

	data := views.AllTasksData{
		ChallengeName:    challenge.Name,
		Tasks:            tasks,
		CompletedTaskIDs: completedSet,
	}

	text := views.RenderAllTasks(data)
	return c.Send(text, keyboards.TeamProgress()) // reuse back button
}

// showCelebration shows the celebration view
func (h *Handler) showCelebration(c tele.Context, challengeID string, participant *domain.Participant) error {
	challenge, _ := h.challenge.GetByID(challengeID)
	tasks, _ := h.task.GetByChallengeID(challengeID)
	totalTasks := len(tasks)

	participants, _ := h.participant.GetByChallengeID(challengeID)

	var teamStatus []*views.TeamMemberStatus
	for _, p := range participants {
		completed, _ := h.completion.CountByParticipantID(p.ID)
		isCompleted := completed >= totalTasks && totalTasks > 0
		teamStatus = append(teamStatus, &views.TeamMemberStatus{
			Emoji:          p.Emoji,
			Name:           p.DisplayName,
			IsCompleted:    isCompleted,
			CompletedTasks: completed,
			TotalTasks:     totalTasks,
		})
	}

	timeTaken := time.Since(participant.JoinedAt)

	data := views.CelebrationData{
		ChallengeName:  challenge.Name,
		TotalTasks:     totalTasks,
		CompletedTasks: totalTasks,
		TimeTaken:      timeTaken,
		TeamStatus:     teamStatus,
	}

	text := views.RenderCelebration(data)
	return c.Send(text, keyboards.Celebration())
}

