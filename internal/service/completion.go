package service

import (
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
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
