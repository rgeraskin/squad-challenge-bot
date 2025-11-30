package service

import (
	"errors"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/repository"
)

var (
	ErrParticipantNotFound = errors.New("participant not found")
	ErrEmojiTaken          = errors.New("emoji is already taken")
	ErrInvalidEmoji        = errors.New("invalid emoji")
	ErrEmptyName           = errors.New("name cannot be empty")
	ErrNameTooLong         = errors.New("name is too long (max 30 characters)")
)

// ParticipantService handles participant business logic
type ParticipantService struct {
	repo repository.Repository
}

// NewParticipantService creates a new ParticipantService
func NewParticipantService(repo repository.Repository) *ParticipantService {
	return &ParticipantService{repo: repo}
}

// Join adds a user to a challenge
func (s *ParticipantService) Join(challengeID string, telegramID int64, displayName, emoji string, timeOffsetMinutes int) (*domain.Participant, error) {
	if displayName == "" {
		return nil, ErrEmptyName
	}
	if len(displayName) > 30 {
		return nil, ErrNameTooLong
	}

	// Check if emoji is taken
	usedEmojis, err := s.repo.Participant().GetUsedEmojis(challengeID)
	if err != nil {
		return nil, err
	}
	for _, e := range usedEmojis {
		if e == emoji {
			return nil, ErrEmojiTaken
		}
	}

	participant := &domain.Participant{
		ChallengeID:       challengeID,
		TelegramID:        telegramID,
		DisplayName:       displayName,
		Emoji:             emoji,
		NotifyEnabled:     true,
		TimeOffsetMinutes: timeOffsetMinutes,
	}

	if err := s.repo.Participant().Create(participant); err != nil {
		return nil, err
	}

	return participant, nil
}

// GetByID retrieves a participant by ID
func (s *ParticipantService) GetByID(id int64) (*domain.Participant, error) {
	participant, err := s.repo.Participant().GetByID(id)
	if err != nil {
		return nil, err
	}
	if participant == nil {
		return nil, ErrParticipantNotFound
	}
	return participant, nil
}

// GetByChallengeAndUser retrieves a participant by challenge and user
func (s *ParticipantService) GetByChallengeAndUser(challengeID string, telegramID int64) (*domain.Participant, error) {
	return s.repo.Participant().GetByChallengeAndUser(challengeID, telegramID)
}

// GetByChallengeID retrieves all participants for a challenge
func (s *ParticipantService) GetByChallengeID(challengeID string) ([]*domain.Participant, error) {
	return s.repo.Participant().GetByChallengeID(challengeID)
}

// UpdateName updates a participant's display name
func (s *ParticipantService) UpdateName(participantID int64, name string) error {
	if name == "" {
		return ErrEmptyName
	}
	if len(name) > 30 {
		return ErrNameTooLong
	}

	participant, err := s.GetByID(participantID)
	if err != nil {
		return err
	}

	participant.DisplayName = name
	return s.repo.Participant().Update(participant)
}

// UpdateEmoji updates a participant's emoji
func (s *ParticipantService) UpdateEmoji(participantID int64, emoji string, challengeID string) error {
	// Check if emoji is taken by others
	participants, err := s.repo.Participant().GetByChallengeID(challengeID)
	if err != nil {
		return err
	}
	for _, p := range participants {
		if p.ID != participantID && p.Emoji == emoji {
			return ErrEmojiTaken
		}
	}

	participant, err := s.GetByID(participantID)
	if err != nil {
		return err
	}

	participant.Emoji = emoji
	return s.repo.Participant().Update(participant)
}

// ToggleNotifications toggles notifications for a participant
func (s *ParticipantService) ToggleNotifications(participantID int64) (bool, error) {
	participant, err := s.GetByID(participantID)
	if err != nil {
		return false, err
	}

	participant.NotifyEnabled = !participant.NotifyEnabled
	if err := s.repo.Participant().Update(participant); err != nil {
		return false, err
	}

	return participant.NotifyEnabled, nil
}

// Leave removes a participant from a challenge
func (s *ParticipantService) Leave(participantID int64) error {
	return s.repo.Participant().Delete(participantID)
}

// CountByChallengeID returns the number of participants in a challenge
func (s *ParticipantService) CountByChallengeID(challengeID string) (int, error) {
	return s.repo.Participant().CountByChallengeID(challengeID)
}

// GetUsedEmojis returns emojis already used in a challenge
func (s *ParticipantService) GetUsedEmojis(challengeID string) ([]string, error) {
	return s.repo.Participant().GetUsedEmojis(challengeID)
}

// UpdateTimeOffset updates a participant's time offset
func (s *ParticipantService) UpdateTimeOffset(participantID int64, offsetMinutes int) error {
	return s.repo.Participant().UpdateTimeOffset(participantID, offsetMinutes)
}
