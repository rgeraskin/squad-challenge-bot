package repository

import (
	"time"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

// ChallengeRepository defines methods for challenge data access
type ChallengeRepository interface {
	Create(challenge *domain.Challenge) error
	GetByID(id string) (*domain.Challenge, error)
	GetByUserID(telegramID int64) ([]*domain.Challenge, error)
	GetAll() ([]*domain.Challenge, error)
	Update(challenge *domain.Challenge) error
	UpdateDailyLimit(id string, limit int) error
	UpdateHideFutureTasks(id string, hide bool) error
	Delete(id string) error
	Exists(id string) (bool, error)
}

// TaskRepository defines methods for task data access
type TaskRepository interface {
	Create(task *domain.Task) error
	GetByID(id int64) (*domain.Task, error)
	GetByChallengeID(challengeID string) ([]*domain.Task, error)
	Update(task *domain.Task) error
	Delete(id int64) error
	GetMaxOrderNum(challengeID string) (int, error)
	UpdateOrderNums(challengeID string, updates map[int64]int) error
	CountByChallengeID(challengeID string) (int, error)
}

// ParticipantRepository defines methods for participant data access
type ParticipantRepository interface {
	Create(participant *domain.Participant) error
	GetByID(id int64) (*domain.Participant, error)
	GetByChallengeAndUser(challengeID string, telegramID int64) (*domain.Participant, error)
	GetByChallengeID(challengeID string) ([]*domain.Participant, error)
	Update(participant *domain.Participant) error
	UpdateTimeOffset(id int64, offsetMinutes int) error
	Delete(id int64) error
	CountByChallengeID(challengeID string) (int, error)
	GetUsedEmojis(challengeID string) ([]string, error)
}

// CompletionRepository defines methods for task completion data access
type CompletionRepository interface {
	Create(completion *domain.TaskCompletion) error
	Delete(taskID, participantID int64) error
	GetByTaskID(taskID int64) ([]*domain.TaskCompletion, error)
	GetByParticipantID(participantID int64) ([]*domain.TaskCompletion, error)
	GetByTaskAndParticipant(taskID, participantID int64) (*domain.TaskCompletion, error)
	CountByParticipantID(participantID int64) (int, error)
	CountCompletionsInRange(participantID int64, from, to time.Time) (int, error)
	GetCompletedTaskIDs(participantID int64) ([]int64, error)
}

// StateRepository defines methods for user state data access
type StateRepository interface {
	Get(telegramID int64) (*domain.UserState, error)
	Set(state *domain.UserState) error
	Reset(telegramID int64) error
	ResetByChallenge(challengeID string) error
}

// SuperAdminRepository defines methods for super admin data access
type SuperAdminRepository interface {
	Create(telegramID int64) error
	Delete(telegramID int64) error
	IsSuperAdmin(telegramID int64) (bool, error)
	GetAll() ([]*domain.SuperAdmin, error)
	Exists(telegramID int64) (bool, error)
}

// TemplateRepository defines methods for template data access
type TemplateRepository interface {
	Create(template *domain.Template) error
	GetByID(id int64) (*domain.Template, error)
	GetAll() ([]*domain.Template, error)
	Delete(id int64) error
	Count() (int, error)
	ExistsByName(name string) (bool, error)
	UpdateName(id int64, name string) error
	UpdateDescription(id int64, description string) error
	UpdateDailyLimit(id int64, limit int) error
	UpdateHideFutureTasks(id int64, hide bool) error
}

// TemplateTaskRepository defines methods for template task data access
type TemplateTaskRepository interface {
	Create(task *domain.TemplateTask) error
	GetByID(id int64) (*domain.TemplateTask, error)
	GetByTemplateID(templateID int64) ([]*domain.TemplateTask, error)
	DeleteByTemplateID(templateID int64) error
	CountByTemplateID(templateID int64) (int, error)
	Delete(id int64) error
	UpdateTitle(id int64, title string) error
	UpdateDescription(id int64, description string) error
	UpdateImage(id int64, imageFileID string) error
	GetMaxOrderNum(templateID int64) (int, error)
	UpdateOrderNum(id int64, orderNum int) error
	UpdateOrderNums(templateID int64, updates map[int64]int) error
}

// Repository combines all repositories
type Repository interface {
	Challenge() ChallengeRepository
	Task() TaskRepository
	Participant() ParticipantRepository
	Completion() CompletionRepository
	State() StateRepository
	SuperAdmin() SuperAdminRepository
	Template() TemplateRepository
	TemplateTask() TemplateTaskRepository
	Close() error
}
