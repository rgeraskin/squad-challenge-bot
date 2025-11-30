package sqlite

import (
	"testing"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

func TestParticipantRepo_Create(t *testing.T) {
	repo := setupTestDB(t)

	// Create challenge first
	challenge := &domain.Challenge{ID: "TEST1234", Name: "Test", CreatorID: 12345}
	repo.Challenge().Create(challenge)

	participant := &domain.Participant{
		ChallengeID:   "TEST1234",
		TelegramID:    12345,
		DisplayName:   "Test User",
		Emoji:         "💪",
		NotifyEnabled: true,
	}

	err := repo.Participant().Create(participant)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if participant.ID == 0 {
		t.Error("Create() should set participant ID")
	}
}

func TestParticipantRepo_GetByChallengeAndUser(t *testing.T) {
	repo := setupTestDB(t)

	// Create challenge and participant
	challenge := &domain.Challenge{ID: "TEST1234", Name: "Test", CreatorID: 12345}
	repo.Challenge().Create(challenge)

	participant := &domain.Participant{
		ChallengeID:   "TEST1234",
		TelegramID:    12345,
		DisplayName:   "Test User",
		Emoji:         "💪",
		NotifyEnabled: true,
	}
	repo.Participant().Create(participant)

	// Get by challenge and user
	got, err := repo.Participant().GetByChallengeAndUser("TEST1234", 12345)
	if err != nil {
		t.Fatalf("GetByChallengeAndUser() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByChallengeAndUser() returned nil")
	}
	if got.DisplayName != "Test User" {
		t.Errorf("DisplayName = %q, want %q", got.DisplayName, "Test User")
	}

	// Non-existent user
	got, _ = repo.Participant().GetByChallengeAndUser("TEST1234", 99999)
	if got != nil {
		t.Error("GetByChallengeAndUser() should return nil for non-existent user")
	}
}

func TestParticipantRepo_GetUsedEmojis(t *testing.T) {
	repo := setupTestDB(t)

	// Create challenge and participants
	challenge := &domain.Challenge{ID: "TEST1234", Name: "Test", CreatorID: 12345}
	repo.Challenge().Create(challenge)

	p1 := &domain.Participant{ChallengeID: "TEST1234", TelegramID: 1, DisplayName: "User1", Emoji: "💪", NotifyEnabled: true}
	p2 := &domain.Participant{ChallengeID: "TEST1234", TelegramID: 2, DisplayName: "User2", Emoji: "🔥", NotifyEnabled: true}
	repo.Participant().Create(p1)
	repo.Participant().Create(p2)

	emojis, err := repo.Participant().GetUsedEmojis("TEST1234")
	if err != nil {
		t.Fatalf("GetUsedEmojis() error = %v", err)
	}
	if len(emojis) != 2 {
		t.Errorf("GetUsedEmojis() returned %d emojis, want 2", len(emojis))
	}
}

func TestParticipantRepo_UniqueConstraint(t *testing.T) {
	repo := setupTestDB(t)

	// Create challenge
	challenge := &domain.Challenge{ID: "TEST1234", Name: "Test", CreatorID: 12345}
	repo.Challenge().Create(challenge)

	// Create first participant
	p1 := &domain.Participant{ChallengeID: "TEST1234", TelegramID: 12345, DisplayName: "User", Emoji: "💪", NotifyEnabled: true}
	err := repo.Participant().Create(p1)
	if err != nil {
		t.Fatalf("First Create() error = %v", err)
	}

	// Try to create duplicate (same challenge + user)
	p2 := &domain.Participant{ChallengeID: "TEST1234", TelegramID: 12345, DisplayName: "User2", Emoji: "🔥", NotifyEnabled: true}
	err = repo.Participant().Create(p2)
	if err == nil {
		t.Error("Create() should fail for duplicate challenge+user")
	}
}
