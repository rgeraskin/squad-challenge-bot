package service

import (
	"errors"
	"math/rand"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/repository"
)

var (
	ErrTemplateNotFound      = errors.New("template not found")
	ErrTemplateNameExists    = errors.New("template with this name already exists")
)

// TemplateService handles template business logic
type TemplateService struct {
	repo repository.Repository
}

// NewTemplateService creates a new TemplateService
func NewTemplateService(repo repository.Repository) *TemplateService {
	return &TemplateService{repo: repo}
}

// CreateFromChallenge creates a template from an existing challenge
func (s *TemplateService) CreateFromChallenge(challengeID string) (*domain.Template, error) {
	// Get challenge
	challenge, err := s.repo.Challenge().GetByID(challengeID)
	if err != nil {
		return nil, err
	}
	if challenge == nil {
		return nil, ErrChallengeNotFound
	}

	// Check if template with this name already exists
	exists, err := s.repo.Template().ExistsByName(challenge.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTemplateNameExists
	}

	// Create template
	template := &domain.Template{
		Name:            challenge.Name,
		Description:     challenge.Description,
		DailyTaskLimit:  challenge.DailyTaskLimit,
		HideFutureTasks: challenge.HideFutureTasks,
	}

	if err := s.repo.Template().Create(template); err != nil {
		return nil, err
	}

	// Copy tasks
	tasks, err := s.repo.Task().GetByChallengeID(challengeID)
	if err != nil {
		// Rollback: delete template if task copying fails
		s.repo.Template().Delete(template.ID)
		return nil, err
	}

	for _, task := range tasks {
		templateTask := &domain.TemplateTask{
			TemplateID:  template.ID,
			OrderNum:    task.OrderNum,
			Title:       task.Title,
			Description: task.Description,
			ImageFileID: task.ImageFileID,
		}
		if err := s.repo.TemplateTask().Create(templateTask); err != nil {
			// Rollback
			s.repo.Template().Delete(template.ID)
			return nil, err
		}
	}

	return template, nil
}

// GetByID retrieves a template by ID
func (s *TemplateService) GetByID(id int64) (*domain.Template, error) {
	template, err := s.repo.Template().GetByID(id)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, ErrTemplateNotFound
	}
	return template, nil
}

// GetAll returns all templates
func (s *TemplateService) GetAll() ([]*domain.Template, error) {
	return s.repo.Template().GetAll()
}

// GetTasks returns all tasks for a template
func (s *TemplateService) GetTasks(templateID int64) ([]*domain.TemplateTask, error) {
	return s.repo.TemplateTask().GetByTemplateID(templateID)
}

// GetTaskByID returns a single template task by ID
func (s *TemplateService) GetTaskByID(id int64) (*domain.TemplateTask, error) {
	return s.repo.TemplateTask().GetByID(id)
}

// GetTaskCount returns the number of tasks in a template
func (s *TemplateService) GetTaskCount(templateID int64) (int, error) {
	return s.repo.TemplateTask().CountByTemplateID(templateID)
}

// Delete deletes a template and its tasks
func (s *TemplateService) Delete(id int64) error {
	// Tasks are automatically deleted via CASCADE
	return s.repo.Template().Delete(id)
}

// Count returns the total number of templates
func (s *TemplateService) Count() (int, error) {
	return s.repo.Template().Count()
}

// UpdateName updates the template name
func (s *TemplateService) UpdateName(id int64, name string) error {
	return s.repo.Template().UpdateName(id, name)
}

// UpdateDescription updates the template description
func (s *TemplateService) UpdateDescription(id int64, description string) error {
	return s.repo.Template().UpdateDescription(id, description)
}

// UpdateDailyLimit updates the template daily limit
func (s *TemplateService) UpdateDailyLimit(id int64, limit int) error {
	return s.repo.Template().UpdateDailyLimit(id, limit)
}

// UpdateHideFutureTasks updates the template hide future tasks setting
func (s *TemplateService) UpdateHideFutureTasks(id int64, hide bool) error {
	return s.repo.Template().UpdateHideFutureTasks(id, hide)
}

// CreateTask creates a new task for a template
func (s *TemplateService) CreateTask(task *domain.TemplateTask) error {
	// Get max order number
	maxOrder, err := s.repo.TemplateTask().GetMaxOrderNum(task.TemplateID)
	if err != nil {
		return err
	}
	task.OrderNum = maxOrder + 1
	return s.repo.TemplateTask().Create(task)
}

// DeleteTask deletes a template task and renumbers remaining tasks
func (s *TemplateService) DeleteTask(taskID int64, templateID int64) error {
	if err := s.repo.TemplateTask().Delete(taskID); err != nil {
		return err
	}

	// Renumber remaining tasks
	tasks, err := s.repo.TemplateTask().GetByTemplateID(templateID)
	if err != nil {
		return err
	}

	for i, t := range tasks {
		if t.OrderNum != i+1 {
			if err := s.repo.TemplateTask().UpdateOrderNum(t.ID, i+1); err != nil {
				return err
			}
		}
	}

	return nil
}

// MoveTask moves a template task to a new position
func (s *TemplateService) MoveTask(taskID int64, templateID int64, newPosition int) error {
	tasks, err := s.repo.TemplateTask().GetByTemplateID(templateID)
	if err != nil {
		return err
	}

	if newPosition < 1 || newPosition > len(tasks) {
		return errors.New("invalid position")
	}

	// Find the task to move
	var movingTask *domain.TemplateTask
	var oldPosition int
	for i, t := range tasks {
		if t.ID == taskID {
			movingTask = t
			oldPosition = i + 1
			break
		}
	}

	if movingTask == nil {
		return errors.New("task not found")
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

	return s.repo.TemplateTask().UpdateOrderNums(templateID, updates)
}

// UpdateTaskTitle updates a template task title
func (s *TemplateService) UpdateTaskTitle(id int64, title string) error {
	return s.repo.TemplateTask().UpdateTitle(id, title)
}

// UpdateTaskDescription updates a template task description
func (s *TemplateService) UpdateTaskDescription(id int64, description string) error {
	return s.repo.TemplateTask().UpdateDescription(id, description)
}

// UpdateTaskImage updates a template task image
func (s *TemplateService) UpdateTaskImage(id int64, imageFileID string) error {
	return s.repo.TemplateTask().UpdateImage(id, imageFileID)
}

// RandomizeTaskOrder randomizes the order of tasks in a template
func (s *TemplateService) RandomizeTaskOrder(templateID int64) error {
	tasks, err := s.repo.TemplateTask().GetByTemplateID(templateID)
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

	return s.repo.TemplateTask().UpdateOrderNums(templateID, updates)
}
