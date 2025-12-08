package service

import (
	"errors"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/repository"
	"github.com/rgeraskin/squad-challenge-bot/internal/util"
)

const (
	MaxChallengesPerUser = 10
	MaxParticipants      = 50
)

var (
	ErrChallengeNotFound    = errors.New("challenge not found")
	ErrChallengeFull        = errors.New("challenge is full")
	ErrAlreadyMember        = errors.New("already a member of this challenge")
	ErrMaxChallengesReached = errors.New("maximum challenges reached")
	ErrNotAdmin             = errors.New("not an admin of this challenge")
)

// ChallengeService handles challenge business logic
type ChallengeService struct {
	repo repository.Repository
}

// NewChallengeService creates a new ChallengeService
func NewChallengeService(repo repository.Repository) *ChallengeService {
	return &ChallengeService{repo: repo}
}

// Create creates a new challenge
func (s *ChallengeService) Create(
	name, description string,
	creatorID int64,
	dailyTaskLimit int,
	hideFutureTasks bool,
) (*domain.Challenge, error) {
	// Check max challenges for user
	challenges, err := s.repo.Challenge().GetByUserID(creatorID)
	if err != nil {
		return nil, err
	}
	if len(challenges) >= MaxChallengesPerUser {
		return nil, ErrMaxChallengesReached
	}

	// Generate unique ID
	var id string
	for i := 0; i < 10; i++ {
		id, err = util.GenerateID()
		if err != nil {
			return nil, err
		}
		exists, err := s.repo.Challenge().Exists(id)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}
		if i == 9 {
			return nil, errors.New("failed to generate unique challenge ID")
		}
	}

	challenge := &domain.Challenge{
		ID:              id,
		Name:            name,
		Description:     description,
		CreatorID:       creatorID,
		DailyTaskLimit:  dailyTaskLimit,
		HideFutureTasks: hideFutureTasks,
	}

	if err := s.repo.Challenge().Create(challenge); err != nil {
		return nil, err
	}

	return challenge, nil
}

// GetByID retrieves a challenge by ID
func (s *ChallengeService) GetByID(id string) (*domain.Challenge, error) {
	challenge, err := s.repo.Challenge().GetByID(id)
	if err != nil {
		return nil, err
	}
	if challenge == nil {
		return nil, ErrChallengeNotFound
	}
	return challenge, nil
}

// GetByUserID retrieves all challenges a user participates in
func (s *ChallengeService) GetByUserID(telegramID int64) ([]*domain.Challenge, error) {
	return s.repo.Challenge().GetByUserID(telegramID)
}

// UpdateName updates a challenge's name
func (s *ChallengeService) UpdateName(
	id string,
	name string,
	userID int64,
	isSuperAdmin bool,
) error {
	challenge, err := s.GetByID(id)
	if err != nil {
		return err
	}

	if challenge.CreatorID != userID && !isSuperAdmin {
		return ErrNotAdmin
	}

	challenge.Name = name
	return s.repo.Challenge().Update(challenge)
}

// UpdateDescription updates a challenge's description (admin only)
func (s *ChallengeService) UpdateDescription(
	id string,
	description string,
	userID int64,
	isSuperAdmin bool,
) error {
	challenge, err := s.GetByID(id)
	if err != nil {
		return err
	}

	if challenge.CreatorID != userID && !isSuperAdmin {
		return ErrNotAdmin
	}

	challenge.Description = description
	return s.repo.Challenge().Update(challenge)
}

// UpdateDailyLimit updates a challenge's daily task limit (admin only)
func (s *ChallengeService) UpdateDailyLimit(
	id string,
	limit int,
	userID int64,
	isSuperAdmin bool,
) error {
	challenge, err := s.GetByID(id)
	if err != nil {
		return err
	}

	if challenge.CreatorID != userID && !isSuperAdmin {
		return ErrNotAdmin
	}

	return s.repo.Challenge().UpdateDailyLimit(id, limit)
}

// UpdateHideFutureTasks updates the hide future tasks setting (admin only)
func (s *ChallengeService) UpdateHideFutureTasks(
	id string,
	hide bool,
	userID int64,
	isSuperAdmin bool,
) error {
	challenge, err := s.GetByID(id)
	if err != nil {
		return err
	}

	if challenge.CreatorID != userID && !isSuperAdmin {
		return ErrNotAdmin
	}

	return s.repo.Challenge().UpdateHideFutureTasks(id, hide)
}

// ToggleHideFutureTasks toggles the hide future tasks setting and returns new value (admin only)
func (s *ChallengeService) ToggleHideFutureTasks(
	id string,
	userID int64,
	isSuperAdmin bool,
) (bool, error) {
	challenge, err := s.GetByID(id)
	if err != nil {
		return false, err
	}

	if challenge.CreatorID != userID && !isSuperAdmin {
		return false, ErrNotAdmin
	}

	newValue := !challenge.HideFutureTasks
	err = s.repo.Challenge().UpdateHideFutureTasks(id, newValue)
	return newValue, err
}

// Delete deletes a challenge (admin only)
func (s *ChallengeService) Delete(id string, userID int64, isSuperAdmin bool) error {
	challenge, err := s.GetByID(id)
	if err != nil {
		return err
	}

	if challenge.CreatorID != userID && !isSuperAdmin {
		return ErrNotAdmin
	}

	return s.repo.Challenge().Delete(id)
}

// IsAdmin checks if a user is the admin of a challenge
func (s *ChallengeService) IsAdmin(challengeID string, userID int64) (bool, error) {
	challenge, err := s.repo.Challenge().GetByID(challengeID)
	if err != nil {
		return false, err
	}
	if challenge == nil {
		return false, nil
	}
	return challenge.CreatorID == userID, nil
}

// CanJoin checks if a user can join a challenge
func (s *ChallengeService) CanJoin(challengeID string, userID int64) error {
	challenge, err := s.repo.Challenge().GetByID(challengeID)
	if err != nil {
		return err
	}
	if challenge == nil {
		return ErrChallengeNotFound
	}

	// Check if already a member
	participant, err := s.repo.Participant().GetByChallengeAndUser(challengeID, userID)
	if err != nil {
		return err
	}
	if participant != nil {
		return ErrAlreadyMember
	}

	// Check participant limit
	count, err := s.repo.Participant().CountByChallengeID(challengeID)
	if err != nil {
		return err
	}
	if count >= MaxParticipants {
		return ErrChallengeFull
	}

	return nil
}

// CreateFromTemplate creates a challenge from a template with provided name and creator
func (s *ChallengeService) CreateFromTemplate(
	template *domain.Template,
	templateTasks []*domain.TemplateTask,
	name string,
	creatorID int64,
) (*domain.Challenge, error) {
	// Check max challenges for user
	challenges, err := s.repo.Challenge().GetByUserID(creatorID)
	if err != nil {
		return nil, err
	}
	if len(challenges) >= MaxChallengesPerUser {
		return nil, ErrMaxChallengesReached
	}

	// Generate unique ID
	var id string
	for i := 0; i < 10; i++ {
		id, err = util.GenerateID()
		if err != nil {
			return nil, err
		}
		exists, err := s.repo.Challenge().Exists(id)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}
		if i == 9 {
			return nil, errors.New("failed to generate unique challenge ID")
		}
	}

	// Create challenge with template settings
	challenge := &domain.Challenge{
		ID:              id,
		Name:            name,
		Description:     template.Description,
		CreatorID:       creatorID,
		DailyTaskLimit:  template.DailyTaskLimit,
		HideFutureTasks: template.HideFutureTasks,
	}

	if err := s.repo.Challenge().Create(challenge); err != nil {
		return nil, err
	}

	// Copy tasks from template
	for _, tt := range templateTasks {
		task := &domain.Task{
			ChallengeID: challenge.ID,
			OrderNum:    tt.OrderNum,
			Title:       tt.Title,
			Description: tt.Description,
			ImageFileID: tt.ImageFileID,
		}
		if err := s.repo.Task().Create(task); err != nil {
			// Rollback: delete challenge (tasks cascade)
			s.repo.Challenge().Delete(challenge.ID)
			return nil, err
		}
	}

	return challenge, nil
}
