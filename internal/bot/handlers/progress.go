package handlers

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/rgeraskin/squad-challenge-bot/internal/bot/assets"
	"github.com/rgeraskin/squad-challenge-bot/internal/bot/keyboards"
	"github.com/rgeraskin/squad-challenge-bot/internal/bot/views"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/logger"
	"github.com/rgeraskin/squad-challenge-bot/internal/service"
	tele "gopkg.in/telebot.v3"
)

// formatDuration formats a duration as HH:MM:SS or shorter
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %02dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// handleCompleteTask completes a task
func (h *Handler) handleCompleteTask(c tele.Context, taskIDStr string) error {
	userID := c.Sender().ID
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	participant, err := h.participant.GetByChallengeAndUser(challengeID, userID)
	if err != nil || participant == nil {
		return h.sendError(c, "😕 You're not in this challenge.")
	}

	// Check daily limit
	challenge, err := h.challenge.GetByID(challengeID)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	logger.Debug("handleCompleteTask",
		"challenge_id", challengeID,
		"task_id", taskID,
		"participant_id", participant.ID,
		"daily_task_limit", challenge.DailyTaskLimit,
		"time_offset_minutes", participant.TimeOffsetMinutes,
	)

	if challenge.DailyTaskLimit > 0 {
		limitInfo, err := h.completion.CheckDailyLimit(participant, challenge.DailyTaskLimit)
		if err != nil {
			logger.Debug("CheckDailyLimit error", "error", err)
			return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
		}

		logger.Debug("PRE-completion limitInfo",
			"allowed", limitInfo.Allowed,
			"completed", limitInfo.Completed,
			"limit", limitInfo.Limit,
		)

		if !limitInfo.Allowed {
			logger.Debug("Daily limit reached BEFORE completion, blocking")
			return h.showDailyLimitReached(c, limitInfo)
		}
	}

	task, err := h.task.GetByID(taskID)
	if err != nil {
		return h.sendError(c, "🤔 Can't find that task.")
	}

	// Check if task is hidden (cannot complete hidden tasks)
	if challenge.HideFutureTasks {
		tasks, _ := h.task.GetByChallengeID(challengeID)
		currentTaskNum := h.completion.GetCurrentTaskNum(participant.ID, tasks)
		// Only check if there's a current task (currentTaskNum > 0 means not all completed)
		if currentTaskNum > 0 && task.OrderNum > currentTaskNum {
			return c.Send(
				"🔒 This task is locked.\n\nComplete your previous tasks first.",
				keyboards.BackToMain(),
			)
		}
	}

	logger.Debug("About to call Complete", "task_id", taskID, "participant_id", participant.ID)
	completion, err := h.completion.Complete(taskID, participant.ID)
	if err != nil {
		logger.Debug("Complete() error", "error", err)
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}
	logger.Debug("Complete() returned", "completion_id", completion.ID, "completed_at", completion.CompletedAt)

	// Post-completion check for daily limit (handles race conditions)
	// If limit exceeded after completion, uncomplete and show limit message
	if challenge.DailyTaskLimit > 0 {
		limitInfo, err := h.completion.CheckDailyLimit(participant, challenge.DailyTaskLimit)
		logger.Debug("POST-completion limitInfo",
			"error", err,
			"allowed", limitInfo.Allowed,
			"completed", limitInfo.Completed,
			"limit", limitInfo.Limit,
		)
		if err == nil && limitInfo != nil && limitInfo.Completed > challenge.DailyTaskLimit {
			// Limit exceeded due to race condition - rollback
			logger.Debug("Race condition detected! Rolling back completion")
			h.completion.Uncomplete(taskID, participant.ID)
			limitInfo.Completed = challenge.DailyTaskLimit // Show correct count
			limitInfo.Allowed = false
			return h.showDailyLimitReached(c, limitInfo)
		}
	}

	// Check if all tasks completed
	tasks, _ := h.task.GetByChallengeID(challengeID)
	allCompleted, _ := h.completion.IsAllCompleted(participant.ID, len(tasks))

	// Notify others
	go h.notification.NotifyTaskCompleted(
		challengeID,
		participant.Emoji,
		participant.DisplayName,
		task.Title,
		userID,
	)

	if allCompleted {
		// Notify challenge completion
		go h.notification.NotifyChallengeCompleted(
			challengeID,
			participant.Emoji,
			participant.DisplayName,
			userID,
		)
		return h.showCelebration(c, challengeID, participant)
	}

	// Show completion feedback with daily progress if limit is set
	if challenge.DailyTaskLimit > 0 {
		limitInfo, _ := h.completion.CheckDailyLimit(participant, challenge.DailyTaskLimit)
		if limitInfo != nil {
			msg := fmt.Sprintf("✅ Task completed! (%d/%d today, resets in %s)",
				limitInfo.Completed, limitInfo.Limit, formatDuration(limitInfo.TimeToReset))
			c.Send(msg)
		}
	}

	return h.showMainChallengeView(c, challengeID)
}

// showDailyLimitReached shows the daily limit reached message
func (h *Handler) showDailyLimitReached(c tele.Context, info *service.DailyLimitInfo) error {
	msg := "🕓 <i>Daily Limit Reached!</i>\n\n"
	msg += fmt.Sprintf("You've completed <b>%d/%d</b> tasks today.\n\n", info.Completed, info.Limit)
	msg += fmt.Sprintf("New day starts in: <b>%s</b>\n\n", formatDuration(info.TimeToReset))
	msg += "🙌 <i>Come back tomorrow to continue!</i>"

	return c.Send(msg, keyboards.BackToMain(), tele.ModeHTML)
}

// handleCompleteCurrent completes the current task
func (h *Handler) handleCompleteCurrent(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	participant, err := h.participant.GetByChallengeAndUser(challengeID, userID)
	if err != nil || participant == nil {
		return h.sendError(c, "😕 You're not in this challenge.")
	}

	tasks, err := h.task.GetByChallengeID(challengeID)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	currentTaskNum := h.completion.GetCurrentTaskNum(participant.ID, tasks)
	if currentTaskNum == 0 {
		return h.sendError(c, "🎉 You've already crushed all the tasks!")
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
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	participant, err := h.participant.GetByChallengeAndUser(challengeID, userID)
	if err != nil || participant == nil {
		return h.sendError(c, "😕 You're not in this challenge.")
	}

	err = h.completion.Uncomplete(taskID, participant.ID)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	return h.showTaskDetail(c, taskIDStr)
}

// showTaskDetail shows the task detail view
func (h *Handler) showTaskDetail(c tele.Context, taskIDStr string) error {
	userID := c.Sender().ID
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	task, err := h.task.GetByID(taskID)
	if err != nil {
		return h.sendError(c, "🤔 Can't find that task.")
	}

	participant, err := h.participant.GetByChallengeAndUser(challengeID, userID)
	if err != nil || participant == nil {
		return h.sendError(c, "😕 You're not in this challenge.")
	}

	// Check if task is hidden
	challenge, err := h.challenge.GetByID(challengeID)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	tasks, _ := h.task.GetByChallengeID(challengeID)
	currentTaskNum := h.completion.GetCurrentTaskNum(participant.ID, tasks)

	// Only show hidden view if there's a current task to work on (currentTaskNum > 0)
	if challenge.HideFutureTasks && currentTaskNum > 0 && task.OrderNum > currentTaskNum {
		// Show hidden task view
		data := views.HiddenTaskDetailData{
			TaskOrderNum:   task.OrderNum,
			CurrentTaskNum: currentTaskNum,
		}
		text := views.RenderHiddenTaskDetail(data)
		return c.Send(text, keyboards.HiddenTaskBack(), tele.ModeHTML)
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

	return c.Send(text, keyboards.TaskDetail(taskID, isCompleted), tele.ModeHTML)
}

// showTeamProgress shows the team progress view
func (h *Handler) showTeamProgress(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	challenge, err := h.challenge.GetByID(challengeID)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	tasks, _ := h.task.GetByChallengeID(challengeID)
	totalTasks := len(tasks)

	participants, err := h.participant.GetByChallengeID(challengeID)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
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
	return c.Send(text, keyboards.TeamProgress(), tele.ModeHTML)
}

// showAllTasks shows the full list of all tasks
func (h *Handler) showAllTasks(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	challenge, err := h.challenge.GetByID(challengeID)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	tasks, err := h.task.GetByChallengeID(challengeID)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	participant, err := h.participant.GetByChallengeAndUser(challengeID, userID)
	if err != nil || participant == nil {
		return h.sendError(c, "😕 You're not in this challenge.")
	}

	// Get completion data
	completedIDs, _ := h.completion.GetCompletedTaskIDs(participant.ID)
	completedSet := make(map[int64]bool)
	for _, id := range completedIDs {
		completedSet[id] = true
	}

	currentTaskNum := h.completion.GetCurrentTaskNum(participant.ID, tasks)

	data := views.AllTasksData{
		ChallengeName:    challenge.Name,
		Tasks:            tasks,
		CompletedTaskIDs: completedSet,
		HideFutureTasks:  challenge.HideFutureTasks,
		CurrentTaskNum:   currentTaskNum,
	}

	text := views.RenderAllTasks(data)
	return c.Send(text, keyboards.TeamProgress(), tele.ModeHTML) // reuse back button
}

// showCelebration shows the celebration view
func (h *Handler) showCelebration(
	c tele.Context,
	challengeID string,
	participant *domain.Participant,
) error {
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

	animation := &tele.Animation{
		File:     tele.FromReader(bytes.NewReader(assets.ChallengeCompletedGIF)),
		FileName: "challenge-completed.gif",
		MIME:     "image/gif",
		Caption:  text,
	}
	return c.Send(animation, keyboards.Celebration(), tele.ModeHTML)
}
