package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/logger"
	"github.com/rgeraskin/squad-challenge-bot/internal/repository/sqlite"
	"github.com/rgeraskin/squad-challenge-bot/internal/service"
	"github.com/rgeraskin/squad-challenge-bot/internal/testutil"
)

func init() {
	// Initialize logger for tests
	logger.Init("error")
}

// testHandler creates a handler with real services for testing
func testHandler(t *testing.T) (*Handler, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "handler-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	repo, err := sqlite.New(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create repository: %v", err)
	}

	cleanup := func() {
		repo.Close()
		os.RemoveAll(tmpDir)
	}

	h := NewHandler(
		repo,
		service.NewChallengeService(repo),
		service.NewTaskService(repo),
		service.NewParticipantService(repo),
		service.NewCompletionService(repo),
		service.NewStateService(repo),
		nil, // notification service not needed for tests
		nil, // bot not needed for tests
	)

	return h, cleanup
}

func TestHandleStart_NoPayload(t *testing.T) {
	h, cleanup := testHandler(t)
	defer cleanup()

	ctx := testutil.NewMockContext(12345)

	err := h.HandleStart(ctx)
	if err != nil {
		t.Fatalf("HandleStart failed: %v", err)
	}

	// Should send welcome message
	if ctx.MessageCount() == 0 {
		t.Error("Expected a message to be sent")
	}

	msg := ctx.LastMessage()
	if !strings.Contains(msg, "Hey there") {
		t.Errorf("Expected welcome message, got: %s", msg)
	}
}

func TestHandleStart_ResetsState(t *testing.T) {
	h, cleanup := testHandler(t)
	defer cleanup()

	userID := int64(12345)

	// Set some state first
	h.state.SetState(userID, domain.StateAwaitingChallengeName)

	ctx := testutil.NewMockContext(userID)
	err := h.HandleStart(ctx)
	if err != nil {
		t.Fatalf("HandleStart failed: %v", err)
	}

	// Verify state was reset
	state, _ := h.state.Get(userID)
	if state.State != domain.StateIdle {
		t.Errorf("State = %q, want %q", state.State, domain.StateIdle)
	}
}

func TestHandleText_IdleState(t *testing.T) {
	h, cleanup := testHandler(t)
	defer cleanup()

	ctx := testutil.NewMockContext(12345).WithMessage("random text")

	err := h.HandleText(ctx)
	if err != nil {
		t.Fatalf("HandleText failed: %v", err)
	}

	// In idle state, text should be ignored (no message sent)
	if ctx.MessageCount() != 0 {
		t.Error("Expected no message in idle state")
	}
}

func TestHandleText_ChallengeNameFlow(t *testing.T) {
	h, cleanup := testHandler(t)
	defer cleanup()

	userID := int64(12345)

	// Start create challenge flow
	h.state.SetState(userID, domain.StateAwaitingChallengeName)

	ctx := testutil.NewMockContext(userID).WithMessage("30-Day Fitness")

	err := h.HandleText(ctx)
	if err != nil {
		t.Fatalf("HandleText failed: %v", err)
	}

	// Should advance to next state (description)
	state, _ := h.state.Get(userID)
	if state.State != domain.StateAwaitingChallengeDescription {
		t.Errorf("State = %q, want %q", state.State, domain.StateAwaitingChallengeDescription)
	}

	// Should have sent a message asking for description
	if ctx.MessageCount() == 0 {
		t.Error("Expected a message to be sent")
	}
}

func TestHandleText_InvalidEmoji(t *testing.T) {
	h, cleanup := testHandler(t)
	defer cleanup()

	userID := int64(12345)

	// Set state to await emoji
	h.state.SetState(userID, domain.StateAwaitingCreatorEmoji)

	// Send non-emoji text
	ctx := testutil.NewMockContext(userID).WithMessage("not an emoji")

	err := h.HandleText(ctx)
	if err != nil {
		t.Fatalf("HandleText failed: %v", err)
	}

	// Should get error message
	msg := ctx.LastMessage()
	if !strings.Contains(msg, "emoji") {
		t.Errorf("Expected emoji error message, got: %s", msg)
	}

	// State should not change
	state, _ := h.state.Get(userID)
	if state.State != domain.StateAwaitingCreatorEmoji {
		t.Errorf("State = %q, want %q", state.State, domain.StateAwaitingCreatorEmoji)
	}
}

func TestHandleText_CompleteChallengeCreation(t *testing.T) {
	h, cleanup := testHandler(t)
	defer cleanup()

	userID := int64(12345)

	// Step 1: Challenge name
	h.state.SetState(userID, domain.StateAwaitingChallengeName)
	ctx := testutil.NewMockContext(userID).WithMessage("Test Challenge")
	h.HandleText(ctx)

	// Step 2: Challenge description
	ctx = testutil.NewMockContext(userID).WithMessage("A test challenge")
	h.HandleText(ctx)

	// Step 3: Creator name
	ctx = testutil.NewMockContext(userID).WithMessage("John")
	h.HandleText(ctx)

	// Step 4: Creator emoji
	h.state.SetStateWithData(userID, domain.StateAwaitingCreatorEmoji, map[string]interface{}{
		"challenge_name":        "Test Challenge",
		"challenge_description": "A test challenge",
		"display_name":          "John",
	})
	ctx = testutil.NewMockContext(userID).WithMessage("💪")
	h.HandleText(ctx)

	// Step 5: Daily limit (use valid number 1-50)
	ctx = testutil.NewMockContext(userID).WithMessage("5")
	h.HandleText(ctx)

	// Step 6: Hide future tasks (select "no")
	ctx = testutil.NewMockContext(userID).WithCallback("hide_future_no")
	h.HandleCallback(ctx)

	// Step 7: Creator sync time (14:30)
	ctx = testutil.NewMockContext(userID).WithMessage("14:30")
	err := h.HandleText(ctx)
	if err != nil {
		t.Fatalf("HandleText failed: %v", err)
	}

	// Verify challenge was created
	challenges, _ := h.challenge.GetByUserID(userID)
	if len(challenges) != 1 {
		t.Errorf("Challenge count = %d, want 1", len(challenges))
	}
	if challenges[0].Name != "Test Challenge" {
		t.Errorf("Challenge name = %q, want %q", challenges[0].Name, "Test Challenge")
	}
}

func TestHandleCallback_Cancel_ResetsState(t *testing.T) {
	h, cleanup := testHandler(t)
	defer cleanup()

	userID := int64(12345)

	// Set some state
	h.state.SetState(userID, domain.StateAwaitingChallengeName)

	ctx := testutil.NewMockContext(userID).WithCallback("cancel")

	err := h.HandleCallback(ctx)
	if err != nil {
		t.Fatalf("HandleCallback failed: %v", err)
	}

	// State should be reset
	state, _ := h.state.Get(userID)
	if state.State != domain.StateIdle {
		t.Errorf("State = %q, want %q", state.State, domain.StateIdle)
	}
}

func TestHandleCallback_CreateChallenge_StartsFlow(t *testing.T) {
	h, cleanup := testHandler(t)
	defer cleanup()

	userID := int64(12345)
	ctx := testutil.NewMockContext(userID).WithCallback("create_challenge")

	err := h.HandleCallback(ctx)
	if err != nil {
		t.Fatalf("HandleCallback failed: %v", err)
	}

	// Should be in awaiting challenge name state
	state, _ := h.state.Get(userID)
	if state.State != domain.StateAwaitingChallengeName {
		t.Errorf("State = %q, want %q", state.State, domain.StateAwaitingChallengeName)
	}

	// Should have sent message asking for name
	if ctx.MessageCount() == 0 {
		t.Error("Expected a message to be sent")
	}
}

func TestHandleCallback_JoinChallenge_StartsFlow(t *testing.T) {
	h, cleanup := testHandler(t)
	defer cleanup()

	userID := int64(12345)
	ctx := testutil.NewMockContext(userID).WithCallback("join_challenge")

	err := h.HandleCallback(ctx)
	if err != nil {
		t.Fatalf("HandleCallback failed: %v", err)
	}

	// Should be in awaiting challenge ID state
	state, _ := h.state.Get(userID)
	if state.State != domain.StateAwaitingChallengeID {
		t.Errorf("State = %q, want %q", state.State, domain.StateAwaitingChallengeID)
	}
}

func TestHandleText_TaskCreationFlow(t *testing.T) {
	h, cleanup := testHandler(t)
	defer cleanup()

	userID := int64(12345)

	// First create a challenge and participant
	challenge, _ := h.challenge.Create("Test", "", userID, 0, false)
	h.participant.Join(challenge.ID, userID, "Test", "💪", 0)
	h.state.SetCurrentChallenge(userID, challenge.ID)

	// Start task creation
	h.state.SetStateWithData(userID, domain.StateAwaitingTaskTitle, map[string]string{
		"challenge_id": challenge.ID,
	})

	ctx := testutil.NewMockContext(userID).WithMessage("Morning Exercise")
	err := h.HandleText(ctx)
	if err != nil {
		t.Fatalf("HandleText failed: %v", err)
	}

	// Should advance to image state
	state, _ := h.state.Get(userID)
	if state.State != domain.StateAwaitingTaskImage {
		t.Errorf("State = %q, want %q", state.State, domain.StateAwaitingTaskImage)
	}
}

func TestSendError(t *testing.T) {
	h, cleanup := testHandler(t)
	defer cleanup()

	ctx := testutil.NewMockContext(12345)

	err := h.sendError(ctx, "Test error message")
	if err != nil {
		t.Fatalf("sendError failed: %v", err)
	}

	msg := ctx.LastMessage()
	if msg != "Test error message" {
		t.Errorf("Message = %q, want %q", msg, "Test error message")
	}
}
