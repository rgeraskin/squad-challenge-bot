package service

import (
	"time"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/logger"
	"github.com/rgeraskin/squad-challenge-bot/internal/repository"
)

// CompletionService handles task completion business logic
type CompletionService struct {
	repo repository.Repository
}

// NewCompletionService creates a new CompletionService
func NewCompletionService(repo repository.Repository) *CompletionService {
	return &CompletionService{repo: repo}
}

// Complete marks a task as completed by a participant
func (s *CompletionService) Complete(taskID, participantID int64) (*domain.TaskCompletion, error) {
	// Check if already completed
	existing, err := s.repo.Completion().GetByTaskAndParticipant(taskID, participantID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil // Already completed
	}

	completion := &domain.TaskCompletion{
		TaskID:        taskID,
		ParticipantID: participantID,
	}

	if err := s.repo.Completion().Create(completion); err != nil {
		return nil, err
	}

	return completion, nil
}

// Uncomplete removes a task completion
func (s *CompletionService) Uncomplete(taskID, participantID int64) error {
	return s.repo.Completion().Delete(taskID, participantID)
}

// IsCompleted checks if a task is completed by a participant
func (s *CompletionService) IsCompleted(taskID, participantID int64) (bool, error) {
	completion, err := s.repo.Completion().GetByTaskAndParticipant(taskID, participantID)
	if err != nil {
		return false, err
	}
	return completion != nil, nil
}

// GetCompletedTaskIDs returns IDs of tasks completed by a participant
func (s *CompletionService) GetCompletedTaskIDs(participantID int64) ([]int64, error) {
	return s.repo.Completion().GetCompletedTaskIDs(participantID)
}

// GetCompletionsByTaskID returns all completions for a task
func (s *CompletionService) GetCompletionsByTaskID(taskID int64) ([]*domain.TaskCompletion, error) {
	return s.repo.Completion().GetByTaskID(taskID)
}

// CountByParticipantID returns the number of completed tasks for a participant
func (s *CompletionService) CountByParticipantID(participantID int64) (int, error) {
	return s.repo.Completion().CountByParticipantID(participantID)
}

// GetCurrentTaskNum calculates the current task number for a participant
// Current task = next uncompleted task after the last completed task (by order)
func (s *CompletionService) GetCurrentTaskNum(participantID int64, tasks []*domain.Task) int {
	if len(tasks) == 0 {
		return 0
	}

	completedIDs, err := s.GetCompletedTaskIDs(participantID)
	if err != nil {
		return 1
	}

	completedSet := make(map[int64]bool)
	for _, id := range completedIDs {
		completedSet[id] = true
	}

	// Find the highest order number among completed tasks
	lastCompletedOrder := 0
	for _, task := range tasks {
		if completedSet[task.ID] && task.OrderNum > lastCompletedOrder {
			lastCompletedOrder = task.OrderNum
		}
	}

	// Find first uncompleted task after the last completed one
	for _, task := range tasks {
		if task.OrderNum > lastCompletedOrder && !completedSet[task.ID] {
			return task.OrderNum
		}
	}

	// All tasks after last completed are done (or no tasks at all)
	// Check if there are any uncompleted tasks before
	for _, task := range tasks {
		if !completedSet[task.ID] {
			return task.OrderNum
		}
	}

	// All tasks completed
	return 0
}

// IsAllCompleted checks if a participant has completed all tasks
func (s *CompletionService) IsAllCompleted(participantID int64, totalTasks int) (bool, error) {
	count, err := s.CountByParticipantID(participantID)
	if err != nil {
		return false, err
	}
	return count >= totalTasks && totalTasks > 0, nil
}

// DailyLimitInfo contains information about daily task completion limits
type DailyLimitInfo struct {
	Allowed       bool
	Completed     int
	Limit         int
	TimeToReset   time.Duration
	UserLocalTime time.Time
}

// GetUserDayBoundaries calculates the start and end of user's current day
func GetUserDayBoundaries(offsetMinutes int) (start, end time.Time) {
	now := time.Now().UTC()
	// Add user's offset to get their local time
	userLocalTime := now.Add(time.Duration(offsetMinutes) * time.Minute)

	// Get start of user's day in their local time
	year, month, day := userLocalTime.Date()
	userDayStart := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	// Convert back to server time by subtracting offset
	serverDayStart := userDayStart.Add(-time.Duration(offsetMinutes) * time.Minute)
	serverDayEnd := serverDayStart.Add(24 * time.Hour)

	return serverDayStart, serverDayEnd
}

// GetUserLocalTime returns current time in user's timezone
func GetUserLocalTime(offsetMinutes int) time.Time {
	return time.Now().UTC().Add(time.Duration(offsetMinutes) * time.Minute)
}

// TimeUntilUserMidnight returns duration until user's next day
func TimeUntilUserMidnight(offsetMinutes int) time.Duration {
	_, dayEnd := GetUserDayBoundaries(offsetMinutes)
	return time.Until(dayEnd)
}

// GetCompletionsToday returns the number of completions for today (user's day)
func (s *CompletionService) GetCompletionsToday(participantID int64, offsetMinutes int) (int, error) {
	dayStart, dayEnd := GetUserDayBoundaries(offsetMinutes)
	logger.Debug("GetCompletionsToday",
		"participant_id", participantID,
		"offset_minutes", offsetMinutes,
		"day_start", dayStart,
		"day_end", dayEnd,
	)
	count, err := s.repo.Completion().CountCompletionsInRange(participantID, dayStart, dayEnd)
	logger.Debug("GetCompletionsToday result", "count", count, "error", err)
	return count, err
}

// CheckDailyLimit checks if a participant can complete more tasks today
func (s *CompletionService) CheckDailyLimit(participant *domain.Participant, dailyLimit int) (*DailyLimitInfo, error) {
	info := &DailyLimitInfo{
		Limit:         dailyLimit,
		UserLocalTime: GetUserLocalTime(participant.TimeOffsetMinutes),
	}

	// If no limit set, always allowed
	if dailyLimit <= 0 {
		info.Allowed = true
		return info, nil
	}

	completed, err := s.GetCompletionsToday(participant.ID, participant.TimeOffsetMinutes)
	if err != nil {
		return nil, err
	}

	info.Completed = completed
	info.TimeToReset = TimeUntilUserMidnight(participant.TimeOffsetMinutes)
	info.Allowed = completed < dailyLimit

	return info, nil
}
