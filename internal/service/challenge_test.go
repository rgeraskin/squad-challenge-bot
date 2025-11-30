package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rgeraskin/squad-challenge-bot/internal/repository/sqlite"
)

func setupTestRepo(t *testing.T) *sqlite.SQLiteRepository {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "squadbot-service-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	repo, err := sqlite.New(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create repository: %v", err)
	}

	t.Cleanup(func() {
		repo.Close()
		os.RemoveAll(tmpDir)
	})

	return repo
}

func TestChallengeService_Create(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewChallengeService(repo)

	challenge, err := svc.Create("Test Challenge", "", 12345, 0, false)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if len(challenge.ID) != 8 {
		t.Errorf("Create() ID length = %d, want 8", len(challenge.ID))
	}
	if challenge.Name != "Test Challenge" {
		t.Errorf("Create() Name = %q, want %q", challenge.Name, "Test Challenge")
	}
	if challenge.CreatorID != 12345 {
		t.Errorf("Create() CreatorID = %d, want %d", challenge.CreatorID, 12345)
	}
}

func TestChallengeService_GetByID(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewChallengeService(repo)

	// Create challenge
	created, _ := svc.Create("Test Challenge", "", 12345, 0, false)

	// Get by ID
	got, err := svc.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "Test Challenge" {
		t.Errorf("GetByID() Name = %q, want %q", got.Name, "Test Challenge")
	}

	// Get non-existent
	_, err = svc.GetByID("NOTEXIST")
	if err != ErrChallengeNotFound {
		t.Errorf("GetByID(non-existent) error = %v, want ErrChallengeNotFound", err)
	}
}

func TestChallengeService_IsAdmin(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewChallengeService(repo)

	challenge, _ := svc.Create("Test Challenge", "", 12345, 0, false)

	// Creator is admin
	isAdmin, err := svc.IsAdmin(challenge.ID, 12345)
	if err != nil {
		t.Fatalf("IsAdmin() error = %v", err)
	}
	if !isAdmin {
		t.Error("IsAdmin() = false for creator, want true")
	}

	// Other user is not admin
	isAdmin, err = svc.IsAdmin(challenge.ID, 99999)
	if err != nil {
		t.Fatalf("IsAdmin() error = %v", err)
	}
	if isAdmin {
		t.Error("IsAdmin() = true for non-creator, want false")
	}
}

func TestChallengeService_MaxChallenges(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	// Create 10 challenges (max)
	for i := 0; i < MaxChallengesPerUser; i++ {
		challenge, err := challengeSvc.Create("Challenge", "", 12345, 0, false)
		if err != nil {
			t.Fatalf("Create() %d error = %v", i, err)
		}
		// Join as participant
		participantSvc.Join(challenge.ID, 12345, "User", "💪", 0)
	}

	// 11th should fail
	_, err := challengeSvc.Create("One More", "", 12345, 0, false)
	if err != ErrMaxChallengesReached {
		t.Errorf("Create() error = %v, want ErrMaxChallengesReached", err)
	}
}

func TestChallengeService_GetByUserID(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	userID := int64(12345)

	// No challenges yet
	challenges, err := challengeSvc.GetByUserID(userID)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}
	if len(challenges) != 0 {
		t.Errorf("GetByUserID() count = %d, want 0", len(challenges))
	}

	// Create and join challenges
	ch1, _ := challengeSvc.Create("Challenge 1", "", userID, 0, false)
	participantSvc.Join(ch1.ID, userID, "User", "💪", 0)

	ch2, _ := challengeSvc.Create("Challenge 2", "", userID, 0, false)
	participantSvc.Join(ch2.ID, userID, "User", "🔥", 0)

	challenges, err = challengeSvc.GetByUserID(userID)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}
	if len(challenges) != 2 {
		t.Errorf("GetByUserID() count = %d, want 2", len(challenges))
	}
}

func TestChallengeService_UpdateName(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewChallengeService(repo)

	userID := int64(12345)
	challenge, _ := svc.Create("Original Name", "", userID, 0, false)

	err := svc.UpdateName(challenge.ID, "New Name", userID)
	if err != nil {
		t.Fatalf("UpdateName() error = %v", err)
	}

	updated, _ := svc.GetByID(challenge.ID)
	if updated.Name != "New Name" {
		t.Errorf("UpdateName() Name = %q, want %q", updated.Name, "New Name")
	}
}

func TestChallengeService_UpdateName_NotAdmin(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewChallengeService(repo)

	challenge, _ := svc.Create("Original Name", "", 12345, 0, false)

	err := svc.UpdateName(challenge.ID, "New Name", 99999)
	if err != ErrNotAdmin {
		t.Errorf("UpdateName() by non-admin: error = %v, want ErrNotAdmin", err)
	}
}

func TestChallengeService_Delete(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewChallengeService(repo)

	challenge, _ := svc.Create("Test Challenge", "", 12345, 0, false)

	err := svc.Delete(challenge.ID, 12345)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = svc.GetByID(challenge.ID)
	if err != ErrChallengeNotFound {
		t.Errorf("GetByID() after delete: error = %v, want ErrChallengeNotFound", err)
	}
}

func TestChallengeService_Delete_NotAdmin(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewChallengeService(repo)

	challenge, _ := svc.Create("Test Challenge", "", 12345, 0, false)

	err := svc.Delete(challenge.ID, 99999)
	if err != ErrNotAdmin {
		t.Errorf("Delete() by non-admin: error = %v, want ErrNotAdmin", err)
	}
}

func TestChallengeService_CanJoin(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345, 0, false)

	// Should be able to join
	err := challengeSvc.CanJoin(challenge.ID, 67890)
	if err != nil {
		t.Errorf("CanJoin() for new user: error = %v, want nil", err)
	}

	// Join and try again
	participantSvc.Join(challenge.ID, 67890, "Jane", "🔥", 0)
	err = challengeSvc.CanJoin(challenge.ID, 67890)
	if err != ErrAlreadyMember {
		t.Errorf("CanJoin() for existing member: error = %v, want ErrAlreadyMember", err)
	}
}

func TestChallengeService_CanJoin_Full(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	participantSvc := NewParticipantService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 10000, 0, false)

	emojis := []string{"💪", "🔥", "⭐", "🎯", "🏆", "🎨", "🎪", "🎭", "🎮", "🎲"}

	// Add 10 participants (max)
	for i := 0; i < MaxParticipants; i++ {
		userID := int64(10000 + i)
		participantSvc.Join(challenge.ID, userID, "User", emojis[i], 0)
	}

	// 11th should fail
	err := challengeSvc.CanJoin(challenge.ID, 99999)
	if err != ErrChallengeFull {
		t.Errorf("CanJoin() for 11th user: error = %v, want ErrChallengeFull", err)
	}
}
