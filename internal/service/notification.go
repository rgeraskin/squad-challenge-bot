package service

import (
	"fmt"

	"github.com/rgeraskin/squad-challenge-bot/internal/logger"
	"github.com/rgeraskin/squad-challenge-bot/internal/repository"
	tele "gopkg.in/telebot.v3"
)

// Notifier interface for sending notifications
type Notifier interface {
	Send(to tele.Recipient, what interface{}, opts ...interface{}) (*tele.Message, error)
}

// NotificationService handles sending notifications to participants
type NotificationService struct {
	repo repository.Repository
	bot  Notifier
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(repo repository.Repository, bot Notifier) *NotificationService {
	return &NotificationService{
		repo: repo,
		bot:  bot,
	}
}

// TelegramUser implements tele.Recipient for a telegram ID
type TelegramUser struct {
	ID int64
}

func (u TelegramUser) Recipient() string {
	return fmt.Sprintf("%d", u.ID)
}

// NotifyJoin notifies all participants that someone joined
func (s *NotificationService) NotifyJoin(challengeID string, joinerEmoji, joinerName string, excludeUserID int64) {
	participants, err := s.repo.Participant().GetByChallengeID(challengeID)
	if err != nil {
		logger.Error("NotifyJoin: failed to get participants", "challenge_id", challengeID, "error", err)
		return
	}

	message := fmt.Sprintf("🎉 %s %s joined the challenge!", joinerEmoji, joinerName)

	for _, p := range participants {
		if p.TelegramID == excludeUserID || !p.NotifyEnabled {
			continue
		}
		if _, err := s.bot.Send(TelegramUser{ID: p.TelegramID}, message); err != nil {
			logger.Warn("NotifyJoin: failed to send", "telegram_id", p.TelegramID, "error", err)
		}
	}
}

// NotifyTaskCompleted notifies all participants that someone completed a task
func (s *NotificationService) NotifyTaskCompleted(challengeID string, completerEmoji, completerName, taskTitle string, excludeUserID int64) {
	participants, err := s.repo.Participant().GetByChallengeID(challengeID)
	if err != nil {
		logger.Error("NotifyTaskCompleted: failed to get participants", "challenge_id", challengeID, "error", err)
		return
	}

	message := fmt.Sprintf("✅ %s %s completed \"%s\"!", completerEmoji, completerName, taskTitle)

	for _, p := range participants {
		if p.TelegramID == excludeUserID || !p.NotifyEnabled {
			continue
		}
		if _, err := s.bot.Send(TelegramUser{ID: p.TelegramID}, message); err != nil {
			logger.Warn("NotifyTaskCompleted: failed to send", "telegram_id", p.TelegramID, "error", err)
		}
	}
}

// NotifyChallengeCompleted notifies all participants that someone finished the challenge
func (s *NotificationService) NotifyChallengeCompleted(challengeID string, completerEmoji, completerName string, excludeUserID int64) {
	participants, err := s.repo.Participant().GetByChallengeID(challengeID)
	if err != nil {
		logger.Error("NotifyChallengeCompleted: failed to get participants", "challenge_id", challengeID, "error", err)
		return
	}

	message := fmt.Sprintf("🏆 %s %s finished the challenge!", completerEmoji, completerName)

	for _, p := range participants {
		if p.TelegramID == excludeUserID || !p.NotifyEnabled {
			continue
		}
		if _, err := s.bot.Send(TelegramUser{ID: p.TelegramID}, message); err != nil {
			logger.Warn("NotifyChallengeCompleted: failed to send", "telegram_id", p.TelegramID, "error", err)
		}
	}
}

// NotifyUserChallengeCompleted sends a celebration message directly to a user who completed
func (s *NotificationService) NotifyUserChallengeCompleted(userID int64, challengeName string) {
	message := fmt.Sprintf("🎉🏆 Congratulations! You've completed \"%s\"!", challengeName)
	if _, err := s.bot.Send(TelegramUser{ID: userID}, message); err != nil {
		logger.Warn("NotifyUserChallengeCompleted: failed to send", "user_id", userID, "error", err)
	}
}

// NotifyLeave notifies all participants that someone left the challenge
func (s *NotificationService) NotifyLeave(challengeID string, leaverEmoji, leaverName string, excludeUserID int64) {
	participants, err := s.repo.Participant().GetByChallengeID(challengeID)
	if err != nil {
		logger.Error("NotifyLeave: failed to get participants", "challenge_id", challengeID, "error", err)
		return
	}

	message := fmt.Sprintf("👋 %s %s left the challenge", leaverEmoji, leaverName)

	for _, p := range participants {
		if p.TelegramID == excludeUserID || !p.NotifyEnabled {
			continue
		}
		if _, err := s.bot.Send(TelegramUser{ID: p.TelegramID}, message); err != nil {
			logger.Warn("NotifyLeave: failed to send", "telegram_id", p.TelegramID, "error", err)
		}
	}
}

// GetParticipantsForDeletion returns the list of participants before a challenge is deleted
// This must be called BEFORE the challenge is deleted due to CASCADE deletes
func (s *NotificationService) GetParticipantsForDeletion(challengeID string) []int64 {
	participants, err := s.repo.Participant().GetByChallengeID(challengeID)
	if err != nil {
		logger.Error("GetParticipantsForDeletion: failed to get participants", "challenge_id", challengeID, "error", err)
		return nil
	}
	ids := make([]int64, len(participants))
	for i, p := range participants {
		ids[i] = p.TelegramID
	}
	return ids
}

// NotifyChallengeDeletedAsync sends deletion notifications to participants
// This should be called AFTER the challenge is deleted, with participant IDs fetched beforehand
func (s *NotificationService) NotifyChallengeDeletedAsync(challengeID, challengeName string, participantIDs []int64, excludeUserID int64) {
	// Reset state for all users viewing this challenge (moves them to /start menu)
	if err := s.repo.State().ResetByChallenge(challengeID); err != nil {
		logger.Error("NotifyChallengeDeletedAsync: failed to reset user states", "challenge_id", challengeID, "error", err)
	}

	message := fmt.Sprintf("❌ Challenge \"%s\" has been deleted by admin.\n\nUse /start to return to main menu.", challengeName)

	for _, telegramID := range participantIDs {
		if telegramID == excludeUserID {
			continue
		}
		// Always send deletion notification regardless of notify_enabled
		if _, err := s.bot.Send(TelegramUser{ID: telegramID}, message); err != nil {
			logger.Warn("NotifyChallengeDeletedAsync: failed to send", "telegram_id", telegramID, "error", err)
		}
	}
}
