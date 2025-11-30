package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rgeraskin/squad-challenge-bot/internal/repository/sqlite"
)

func TestParticipantService_Join(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	// Create a challenge first
	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)

	// Join challenge
	participant, err := participantSvc.Join(challenge.ID, 12345, "John", "💪")
	if err != nil {
		t.Fatalf("Join failed: %v", err)
	}
	if participant.DisplayName != "John" {
		t.Errorf("DisplayName = %q, want %q", participant.DisplayName, "John")
	}
	if participant.Emoji != "💪" {
		t.Errorf("Emoji = %q, want %q", participant.Emoji, "💪")
	}
	if !participant.NotifyEnabled {
		t.Error("NotifyEnabled should be true by default")
	}
}

func TestParticipantService_Join_EmptyName(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)

	_, err := participantSvc.Join(challenge.ID, 12345, "", "💪")
	if err != ErrEmptyName {
		t.Errorf("Join with empty name: error = %v, want ErrEmptyName", err)
	}
}

func TestParticipantService_Join_NameTooLong(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)

	longName := strings.Repeat("a", 31)
	_, err := participantSvc.Join(challenge.ID, 12345, longName, "💪")
	if err != ErrNameTooLong {
		t.Errorf("Join with long name: error = %v, want ErrNameTooLong", err)
	}
}

func TestParticipantService_Join_EmojiTaken(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)

	// First participant takes the emoji
	participantSvc.Join(challenge.ID, 12345, "John", "💪")

	// Second participant tries to use the same emoji
	_, err := participantSvc.Join(challenge.ID, 67890, "Jane", "💪")
	if err != ErrEmojiTaken {
		t.Errorf("Join with taken emoji: error = %v, want ErrEmojiTaken", err)
	}
}

func TestParticipantService_GetByID(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)
	created, _ := participantSvc.Join(challenge.ID, 12345, "John", "💪")

	participant, err := participantSvc.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if participant.DisplayName != "John" {
		t.Errorf("DisplayName = %q, want %q", participant.DisplayName, "John")
	}
}

func TestParticipantService_GetByID_NotFound(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	participantSvc := NewParticipantService(repo)

	_, err := participantSvc.GetByID(99999)
	if err != ErrParticipantNotFound {
		t.Errorf("GetByID for non-existent: error = %v, want ErrParticipantNotFound", err)
	}
}

func TestParticipantService_UpdateName(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)
	participant, _ := participantSvc.Join(challenge.ID, 12345, "John", "💪")

	err := participantSvc.UpdateName(participant.ID, "Johnny")
	if err != nil {
		t.Fatalf("UpdateName failed: %v", err)
	}

	updated, _ := participantSvc.GetByID(participant.ID)
	if updated.DisplayName != "Johnny" {
		t.Errorf("DisplayName = %q, want %q", updated.DisplayName, "Johnny")
	}
}

func TestParticipantService_UpdateName_Empty(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)
	participant, _ := participantSvc.Join(challenge.ID, 12345, "John", "💪")

	err := participantSvc.UpdateName(participant.ID, "")
	if err != ErrEmptyName {
		t.Errorf("UpdateName with empty: error = %v, want ErrEmptyName", err)
	}
}

func TestParticipantService_UpdateEmoji(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)
	participant, _ := participantSvc.Join(challenge.ID, 12345, "John", "💪")

	err := participantSvc.UpdateEmoji(participant.ID, "🔥", challenge.ID)
	if err != nil {
		t.Fatalf("UpdateEmoji failed: %v", err)
	}

	updated, _ := participantSvc.GetByID(participant.ID)
	if updated.Emoji != "🔥" {
		t.Errorf("Emoji = %q, want %q", updated.Emoji, "🔥")
	}
}

func TestParticipantService_UpdateEmoji_Taken(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)
	participantSvc.Join(challenge.ID, 12345, "John", "💪")
	participant2, _ := participantSvc.Join(challenge.ID, 67890, "Jane", "🔥")

	err := participantSvc.UpdateEmoji(participant2.ID, "💪", challenge.ID)
	if err != ErrEmojiTaken {
		t.Errorf("UpdateEmoji to taken: error = %v, want ErrEmojiTaken", err)
	}
}

func TestParticipantService_ToggleNotifications(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)
	participant, _ := participantSvc.Join(challenge.ID, 12345, "John", "💪")

	// Initially enabled
	if !participant.NotifyEnabled {
		t.Error("NotifyEnabled should be true initially")
	}

	// Toggle off
	enabled, err := participantSvc.ToggleNotifications(participant.ID)
	if err != nil {
		t.Fatalf("ToggleNotifications failed: %v", err)
	}
	if enabled {
		t.Error("NotifyEnabled should be false after toggle")
	}

	// Toggle on
	enabled, err = participantSvc.ToggleNotifications(participant.ID)
	if err != nil {
		t.Fatalf("ToggleNotifications failed: %v", err)
	}
	if !enabled {
		t.Error("NotifyEnabled should be true after second toggle")
	}
}

func TestParticipantService_Leave(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)
	participant, _ := participantSvc.Join(challenge.ID, 12345, "John", "💪")

	err := participantSvc.Leave(participant.ID)
	if err != nil {
		t.Fatalf("Leave failed: %v", err)
	}

	_, err = participantSvc.GetByID(participant.ID)
	if err != ErrParticipantNotFound {
		t.Errorf("GetByID after leave: error = %v, want ErrParticipantNotFound", err)
	}
}

func TestParticipantService_CountByChallengeID(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)

	count, _ := participantSvc.CountByChallengeID(challenge.ID)
	if count != 0 {
		t.Errorf("Count = %d, want 0", count)
	}

	participantSvc.Join(challenge.ID, 12345, "John", "💪")
	participantSvc.Join(challenge.ID, 67890, "Jane", "🔥")

	count, _ = participantSvc.CountByChallengeID(challenge.ID)
	if count != 2 {
		t.Errorf("Count = %d, want 2", count)
	}
}

func TestParticipantService_GetUsedEmojis(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)

	participantSvc.Join(challenge.ID, 12345, "John", "💪")
	participantSvc.Join(challenge.ID, 67890, "Jane", "🔥")

	emojis, err := participantSvc.GetUsedEmojis(challenge.ID)
	if err != nil {
		t.Fatalf("GetUsedEmojis failed: %v", err)
	}
	if len(emojis) != 2 {
		t.Errorf("Emoji count = %d, want 2", len(emojis))
	}
}

func TestParticipantService_GetByChallengeID(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)

	participantSvc.Join(challenge.ID, 12345, "John", "💪")
	participantSvc.Join(challenge.ID, 67890, "Jane", "🔥")

	participants, err := participantSvc.GetByChallengeID(challenge.ID)
	if err != nil {
		t.Fatalf("GetByChallengeID failed: %v", err)
	}
	if len(participants) != 2 {
		t.Errorf("Participant count = %d, want 2", len(participants))
	}
}

func TestParticipantService_GetByChallengeAndUser(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "participant-test-*")
	defer os.RemoveAll(tmpDir)

	repo, _ := sqlite.New(filepath.Join(tmpDir, "test.db"))
	defer repo.Close()

	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)
	participantSvc.Join(challenge.ID, 12345, "John", "💪")

	// Find existing
	participant, err := participantSvc.GetByChallengeAndUser(challenge.ID, 12345)
	if err != nil {
		t.Fatalf("GetByChallengeAndUser failed: %v", err)
	}
	if participant == nil {
		t.Error("Expected participant, got nil")
	}
	if participant.DisplayName != "John" {
		t.Errorf("DisplayName = %q, want %q", participant.DisplayName, "John")
	}

	// Find non-existing
	participant, err = participantSvc.GetByChallengeAndUser(challenge.ID, 99999)
	if err != nil {
		t.Fatalf("GetByChallengeAndUser failed: %v", err)
	}
	if participant != nil {
		t.Error("Expected nil for non-existing user")
	}
}
