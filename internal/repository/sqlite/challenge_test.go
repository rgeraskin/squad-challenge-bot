package sqlite

import (
	"testing"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

func TestChallengeRepo_Create(t *testing.T) {
	repo := setupTestDB(t)

	challenge := &domain.Challenge{
		ID:        "TEST1234",
		Name:      "Test Challenge",
		CreatorID: 12345,
	}

	err := repo.Challenge().Create(challenge)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify it was created
	got, err := repo.Challenge().GetByID("TEST1234")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByID() returned nil")
	}
	if got.Name != "Test Challenge" {
		t.Errorf("GetByID().Name = %q, want %q", got.Name, "Test Challenge")
	}
	if got.CreatorID != 12345 {
		t.Errorf("GetByID().CreatorID = %d, want %d", got.CreatorID, 12345)
	}
}

func TestChallengeRepo_GetByID_NotFound(t *testing.T) {
	repo := setupTestDB(t)

	got, err := repo.Challenge().GetByID("NOTEXIST")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got != nil {
		t.Error("GetByID() should return nil for non-existent ID")
	}
}

func TestChallengeRepo_Exists(t *testing.T) {
	repo := setupTestDB(t)

	// Create a challenge
	challenge := &domain.Challenge{
		ID:        "TEST1234",
		Name:      "Test Challenge",
		CreatorID: 12345,
	}
	repo.Challenge().Create(challenge)

	// Test exists
	exists, err := repo.Challenge().Exists("TEST1234")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() = false, want true")
	}

	// Test not exists
	exists, err = repo.Challenge().Exists("NOTEXIST")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Exists() = true, want false")
	}
}

func TestChallengeRepo_Update(t *testing.T) {
	repo := setupTestDB(t)

	challenge := &domain.Challenge{
		ID:        "TEST1234",
		Name:      "Original Name",
		CreatorID: 12345,
	}
	repo.Challenge().Create(challenge)

	// Update name
	challenge.Name = "Updated Name"
	err := repo.Challenge().Update(challenge)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update
	got, _ := repo.Challenge().GetByID("TEST1234")
	if got.Name != "Updated Name" {
		t.Errorf("After Update(), Name = %q, want %q", got.Name, "Updated Name")
	}
}

func TestChallengeRepo_Delete(t *testing.T) {
	repo := setupTestDB(t)

	challenge := &domain.Challenge{
		ID:        "TEST1234",
		Name:      "Test Challenge",
		CreatorID: 12345,
	}
	repo.Challenge().Create(challenge)

	err := repo.Challenge().Delete("TEST1234")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	got, _ := repo.Challenge().GetByID("TEST1234")
	if got != nil {
		t.Error("GetByID() should return nil after Delete()")
	}
}

func TestChallengeRepo_GetByUserID(t *testing.T) {
	repo := setupTestDB(t)

	// Create challenges
	c1 := &domain.Challenge{ID: "CHAL0001", Name: "Challenge 1", CreatorID: 12345}
	c2 := &domain.Challenge{ID: "CHAL0002", Name: "Challenge 2", CreatorID: 12345}
	repo.Challenge().Create(c1)
	repo.Challenge().Create(c2)

	// Add user as participant
	p1 := &domain.Participant{ChallengeID: "CHAL0001", TelegramID: 12345, DisplayName: "User", Emoji: "💪", NotifyEnabled: true}
	p2 := &domain.Participant{ChallengeID: "CHAL0002", TelegramID: 12345, DisplayName: "User", Emoji: "🔥", NotifyEnabled: true}
	repo.Participant().Create(p1)
	repo.Participant().Create(p2)

	// Get challenges for user
	challenges, err := repo.Challenge().GetByUserID(12345)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}
	if len(challenges) != 2 {
		t.Errorf("GetByUserID() returned %d challenges, want 2", len(challenges))
	}
}
