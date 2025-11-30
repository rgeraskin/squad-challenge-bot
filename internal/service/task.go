package service

import (
	"errors"
	"math/rand"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/repository"
)

const MaxTasksPerChallenge = 50

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrMaxTasksReached   = errors.New("maximum tasks reached")
	ErrEmptyTaskTitle    = errors.New("task title cannot be empty")
)

// TaskService handles task business logic
type TaskService struct {
	repo repository.Repository
}

// NewTaskService creates a new TaskService
func NewTaskService(repo repository.Repository) *TaskService {
	return &TaskService{repo: repo}
}

// Create creates a new task
func (s *TaskService) Create(challengeID, title, description, imageFileID string) (*domain.Task, error) {
	if title == "" {
		return nil, ErrEmptyTaskTitle
	}

	// Check max tasks
	count, err := s.repo.Task().CountByChallengeID(challengeID)
	if err != nil {
		return nil, err
	}
	if count >= MaxTasksPerChallenge {
		return nil, ErrMaxTasksReached
	}

	// Get next order number
	maxOrder, err := s.repo.Task().GetMaxOrderNum(challengeID)
	if err != nil {
		return nil, err
	}

	task := &domain.Task{
		ChallengeID: challengeID,
		OrderNum:    maxOrder + 1,
		Title:       title,
		Description: description,
		ImageFileID: imageFileID,
	}

	if err := s.repo.Task().Create(task); err != nil {
		return nil, err
	}

	return task, nil
}

// GetByID retrieves a task by ID
func (s *TaskService) GetByID(id int64) (*domain.Task, error) {
	task, err := s.repo.Task().GetByID(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

// GetByChallengeID retrieves all tasks for a challenge
func (s *TaskService) GetByChallengeID(challengeID string) ([]*domain.Task, error) {
	return s.repo.Task().GetByChallengeID(challengeID)
}

// Update updates a task
func (s *TaskService) Update(task *domain.Task) error {
	if task.Title == "" {
		return ErrEmptyTaskTitle
	}
	return s.repo.Task().Update(task)
}

// Delete deletes a task and renumbers remaining tasks
func (s *TaskService) Delete(taskID int64, challengeID string) error {
	task, err := s.GetByID(taskID)
	if err != nil {
		return err
	}

	if err := s.repo.Task().Delete(taskID); err != nil {
		return err
	}

	// Renumber remaining tasks
	tasks, err := s.repo.Task().GetByChallengeID(challengeID)
	if err != nil {
		return err
	}

	updates := make(map[int64]int)
	for i, t := range tasks {
		if t.OrderNum != i+1 {
			updates[t.ID] = i + 1
		}
	}

	if len(updates) > 0 {
		return s.repo.Task().UpdateOrderNums(challengeID, updates)
	}

	_ = task // silence unused warning
	return nil
}

// MoveTask moves a task to a new position
func (s *TaskService) MoveTask(taskID int64, challengeID string, newPosition int) error {
	tasks, err := s.repo.Task().GetByChallengeID(challengeID)
	if err != nil {
		return err
	}

	if newPosition < 1 || newPosition > len(tasks) {
		return errors.New("invalid position")
	}

	// Find the task to move
	var movingTask *domain.Task
	var oldPosition int
	for i, t := range tasks {
		if t.ID == taskID {
			movingTask = t
			oldPosition = i + 1
			break
		}
	}

	if movingTask == nil {
		return ErrTaskNotFound
	}

	if oldPosition == newPosition {
		return nil // No change needed
	}

	// Calculate new order numbers
	updates := make(map[int64]int)

	if newPosition < oldPosition {
		// Moving up: shift tasks between newPosition and oldPosition down
		for _, t := range tasks {
			if t.OrderNum >= newPosition && t.OrderNum < oldPosition {
				updates[t.ID] = t.OrderNum + 1
			}
		}
	} else {
		// Moving down: shift tasks between oldPosition and newPosition up
		for _, t := range tasks {
			if t.OrderNum > oldPosition && t.OrderNum <= newPosition {
				updates[t.ID] = t.OrderNum - 1
			}
		}
	}

	updates[taskID] = newPosition

	return s.repo.Task().UpdateOrderNums(challengeID, updates)
}

// CountByChallengeID returns the number of tasks in a challenge
func (s *TaskService) CountByChallengeID(challengeID string) (int, error) {
	return s.repo.Task().CountByChallengeID(challengeID)
}

// RandomizeOrder randomizes the order of tasks in a challenge
func (s *TaskService) RandomizeOrder(challengeID string) error {
	tasks, err := s.repo.Task().GetByChallengeID(challengeID)
	if err != nil {
		return err
	}

	if len(tasks) < 2 {
		return nil // Nothing to randomize
	}

	// Create shuffled positions
	positions := make([]int, len(tasks))
	for i := range positions {
		positions[i] = i + 1
	}

	// Fisher-Yates shuffle
	for i := len(positions) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		positions[i], positions[j] = positions[j], positions[i]
	}

	// Build updates map
	updates := make(map[int64]int)
	for i, t := range tasks {
		updates[t.ID] = positions[i]
	}

	return s.repo.Task().UpdateOrderNums(challengeID, updates)
}
