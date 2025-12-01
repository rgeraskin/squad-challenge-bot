package service

import (
	"encoding/json"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/repository"
)

// StateService handles user state management
type StateService struct {
	repo repository.Repository
}

// NewStateService creates a new StateService
func NewStateService(repo repository.Repository) *StateService {
	return &StateService{repo: repo}
}

// Get retrieves the current state for a user
func (s *StateService) Get(telegramID int64) (*domain.UserState, error) {
	return s.repo.State().Get(telegramID)
}

// SetState sets the user's conversation state
func (s *StateService) SetState(telegramID int64, state string) error {
	userState, err := s.Get(telegramID)
	if err != nil {
		return err
	}

	userState.State = state
	return s.repo.State().Set(userState)
}

// SetStateWithData sets the user's state with temporary data
func (s *StateService) SetStateWithData(telegramID int64, state string, data interface{}) error {
	userState, err := s.Get(telegramID)
	if err != nil {
		return err
	}

	userState.State = state

	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		userState.TempData = string(jsonData)
	} else {
		userState.TempData = ""
	}

	return s.repo.State().Set(userState)
}

// SetCurrentChallenge sets the user's active challenge
func (s *StateService) SetCurrentChallenge(telegramID int64, challengeID string) error {
	userState, err := s.Get(telegramID)
	if err != nil {
		return err
	}

	userState.CurrentChallenge = challengeID
	return s.repo.State().Set(userState)
}

// GetTempData retrieves and unmarshals the user's temporary data
func (s *StateService) GetTempData(telegramID int64, target interface{}) error {
	userState, err := s.Get(telegramID)
	if err != nil {
		return err
	}

	if userState.TempData == "" {
		return nil
	}

	return json.Unmarshal([]byte(userState.TempData), target)
}

// Reset resets the user's state to idle
func (s *StateService) Reset(telegramID int64) error {
	return s.repo.State().Reset(telegramID)
}

// ResetKeepChallenge resets state to idle but keeps current challenge
func (s *StateService) ResetKeepChallenge(telegramID int64) error {
	userState, err := s.Get(telegramID)
	if err != nil {
		return err
	}

	userState.State = domain.StateIdle
	userState.TempData = ""
	return s.repo.State().Set(userState)
}

// ResetByChallenge resets all users who have a specific challenge as their current challenge
func (s *StateService) ResetByChallenge(challengeID string) error {
	return s.repo.State().ResetByChallenge(challengeID)
}
