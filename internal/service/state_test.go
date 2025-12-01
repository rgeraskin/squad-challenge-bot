package service

import (
	"testing"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

func TestStateService_Get_DefaultIdle(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewStateService(repo)

	// Get state for user that doesn't exist
	state, err := svc.Get(12345)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if state.State != domain.StateIdle {
		t.Errorf("Get() State = %q, want %q", state.State, domain.StateIdle)
	}
	if state.TelegramID != 12345 {
		t.Errorf("Get() TelegramID = %d, want %d", state.TelegramID, 12345)
	}
}

func TestStateService_SetState(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewStateService(repo)

	err := svc.SetState(12345, domain.StateAwaitingChallengeName)
	if err != nil {
		t.Fatalf("SetState() error = %v", err)
	}

	state, _ := svc.Get(12345)
	if state.State != domain.StateAwaitingChallengeName {
		t.Errorf("After SetState(), State = %q, want %q", state.State, domain.StateAwaitingChallengeName)
	}
}

func TestStateService_SetStateWithData(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewStateService(repo)

	data := map[string]string{"key": "value"}
	err := svc.SetStateWithData(12345, domain.StateAwaitingCreatorName, data)
	if err != nil {
		t.Fatalf("SetStateWithData() error = %v", err)
	}

	state, _ := svc.Get(12345)
	if state.State != domain.StateAwaitingCreatorName {
		t.Errorf("After SetStateWithData(), State = %q, want %q", state.State, domain.StateAwaitingCreatorName)
	}
	if state.TempData == "" {
		t.Error("After SetStateWithData(), TempData should not be empty")
	}
}

func TestStateService_GetTempData(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewStateService(repo)

	// Set state with data
	data := map[string]interface{}{
		"challenge_name": "Test Challenge",
		"display_name":   "John",
	}
	svc.SetStateWithData(12345, domain.StateAwaitingCreatorEmoji, data)

	// Get temp data
	var retrieved map[string]interface{}
	err := svc.GetTempData(12345, &retrieved)
	if err != nil {
		t.Fatalf("GetTempData() error = %v", err)
	}

	if retrieved["challenge_name"] != "Test Challenge" {
		t.Errorf("GetTempData() challenge_name = %v, want %v", retrieved["challenge_name"], "Test Challenge")
	}
}

func TestStateService_Reset(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewStateService(repo)

	// Set some state
	svc.SetStateWithData(12345, domain.StateAwaitingTaskTitle, map[string]string{"key": "value"})

	// Reset
	err := svc.Reset(12345)
	if err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	state, _ := svc.Get(12345)
	if state.State != domain.StateIdle {
		t.Errorf("After Reset(), State = %q, want %q", state.State, domain.StateIdle)
	}
}

func TestStateService_SetCurrentChallenge(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewStateService(repo)

	err := svc.SetCurrentChallenge(12345, "ABC12345")
	if err != nil {
		t.Fatalf("SetCurrentChallenge() error = %v", err)
	}

	state, _ := svc.Get(12345)
	if state.CurrentChallenge != "ABC12345" {
		t.Errorf("After SetCurrentChallenge(), CurrentChallenge = %q, want %q", state.CurrentChallenge, "ABC12345")
	}
}

func TestStateService_ResetKeepChallenge(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewStateService(repo)

	// Set state with challenge
	svc.SetState(12345, domain.StateAwaitingTaskTitle)
	svc.SetCurrentChallenge(12345, "ABC12345")

	// Reset but keep challenge
	err := svc.ResetKeepChallenge(12345)
	if err != nil {
		t.Fatalf("ResetKeepChallenge() error = %v", err)
	}

	state, _ := svc.Get(12345)
	if state.State != domain.StateIdle {
		t.Errorf("After ResetKeepChallenge(), State = %q, want %q", state.State, domain.StateIdle)
	}
	if state.CurrentChallenge != "ABC12345" {
		t.Errorf("After ResetKeepChallenge(), CurrentChallenge = %q, want %q", state.CurrentChallenge, "ABC12345")
	}
}

func TestStateService_ResetByChallenge(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewStateService(repo)

	challengeID := "TESTCH01"
	otherChallengeID := "OTHERCH1"

	// Set up users with different challenges
	user1 := int64(11111)
	user2 := int64(22222)
	user3 := int64(33333)

	// User 1 and 2 are viewing the target challenge
	svc.SetCurrentChallenge(user1, challengeID)
	svc.SetStateWithData(user1, domain.StateAwaitingTaskTitle, map[string]string{"task": "test"})

	svc.SetCurrentChallenge(user2, challengeID)
	svc.SetState(user2, domain.StateIdle)

	// User 3 is viewing a different challenge
	svc.SetCurrentChallenge(user3, otherChallengeID)
	svc.SetState(user3, domain.StateAwaitingEditTitle)

	// Reset all users viewing the target challenge
	err := svc.ResetByChallenge(challengeID)
	if err != nil {
		t.Fatalf("ResetByChallenge() error = %v", err)
	}

	// User 1 should be reset (state and currentChallenge cleared)
	state1, _ := svc.Get(user1)
	if state1.State != domain.StateIdle {
		t.Errorf("User1 State = %q, want %q", state1.State, domain.StateIdle)
	}
	if state1.CurrentChallenge != "" {
		t.Errorf("User1 CurrentChallenge = %q, want empty", state1.CurrentChallenge)
	}
	if state1.TempData != "" {
		t.Errorf("User1 TempData = %q, want empty", state1.TempData)
	}

	// User 2 should also be reset
	state2, _ := svc.Get(user2)
	if state2.State != domain.StateIdle {
		t.Errorf("User2 State = %q, want %q", state2.State, domain.StateIdle)
	}
	if state2.CurrentChallenge != "" {
		t.Errorf("User2 CurrentChallenge = %q, want empty", state2.CurrentChallenge)
	}

	// User 3 should NOT be affected (different challenge)
	state3, _ := svc.Get(user3)
	if state3.State != domain.StateAwaitingEditTitle {
		t.Errorf("User3 State = %q, want %q", state3.State, domain.StateAwaitingEditTitle)
	}
	if state3.CurrentChallenge != otherChallengeID {
		t.Errorf("User3 CurrentChallenge = %q, want %q", state3.CurrentChallenge, otherChallengeID)
	}
}

// State Transition Tests - ensure all state transitions work correctly

func TestStateTransitions_ChallengeCreation(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewStateService(repo)
	userID := int64(12345)

	transitions := []struct {
		name      string
		fromState string
		toState   string
	}{
		{"idle -> awaiting_challenge_name", domain.StateIdle, domain.StateAwaitingChallengeName},
		{"awaiting_challenge_name -> awaiting_creator_name", domain.StateAwaitingChallengeName, domain.StateAwaitingCreatorName},
		{"awaiting_creator_name -> awaiting_creator_emoji", domain.StateAwaitingCreatorName, domain.StateAwaitingCreatorEmoji},
		{"awaiting_creator_emoji -> idle", domain.StateAwaitingCreatorEmoji, domain.StateIdle},
	}

	for _, tt := range transitions {
		t.Run(tt.name, func(t *testing.T) {
			svc.SetState(userID, tt.fromState)
			svc.SetState(userID, tt.toState)

			state, _ := svc.Get(userID)
			if state.State != tt.toState {
				t.Errorf("Transition %s failed: got %s, want %s", tt.name, state.State, tt.toState)
			}
		})
	}
}

func TestStateTransitions_TaskManagement(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewStateService(repo)
	userID := int64(12345)

	transitions := []struct {
		name      string
		fromState string
		toState   string
	}{
		{"idle -> awaiting_task_title", domain.StateIdle, domain.StateAwaitingTaskTitle},
		{"awaiting_task_title -> awaiting_task_image", domain.StateAwaitingTaskTitle, domain.StateAwaitingTaskImage},
		{"awaiting_task_image -> awaiting_task_description", domain.StateAwaitingTaskImage, domain.StateAwaitingTaskDescription},
		{"awaiting_task_description -> idle", domain.StateAwaitingTaskDescription, domain.StateIdle},
		{"idle -> awaiting_edit_title", domain.StateIdle, domain.StateAwaitingEditTitle},
		{"idle -> awaiting_edit_description", domain.StateIdle, domain.StateAwaitingEditDescription},
		{"idle -> awaiting_edit_image", domain.StateIdle, domain.StateAwaitingEditImage},
		{"idle -> reorder_select_task", domain.StateIdle, domain.StateReorderSelectTask},
		{"reorder_select_task -> reorder_select_position", domain.StateReorderSelectTask, domain.StateReorderSelectPosition},
	}

	for _, tt := range transitions {
		t.Run(tt.name, func(t *testing.T) {
			svc.SetState(userID, tt.fromState)
			svc.SetState(userID, tt.toState)

			state, _ := svc.Get(userID)
			if state.State != tt.toState {
				t.Errorf("Transition %s failed: got %s, want %s", tt.name, state.State, tt.toState)
			}
		})
	}
}

func TestStateTransitions_JoinChallenge(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewStateService(repo)
	userID := int64(12345)

	transitions := []struct {
		name      string
		fromState string
		toState   string
	}{
		{"idle -> awaiting_challenge_id", domain.StateIdle, domain.StateAwaitingChallengeID},
		{"awaiting_challenge_id -> awaiting_participant_name", domain.StateAwaitingChallengeID, domain.StateAwaitingParticipantName},
		{"awaiting_participant_name -> awaiting_participant_emoji", domain.StateAwaitingParticipantName, domain.StateAwaitingParticipantEmoji},
		{"awaiting_participant_emoji -> idle", domain.StateAwaitingParticipantEmoji, domain.StateIdle},
	}

	for _, tt := range transitions {
		t.Run(tt.name, func(t *testing.T) {
			svc.SetState(userID, tt.fromState)
			svc.SetState(userID, tt.toState)

			state, _ := svc.Get(userID)
			if state.State != tt.toState {
				t.Errorf("Transition %s failed: got %s, want %s", tt.name, state.State, tt.toState)
			}
		})
	}
}

func TestStateTransitions_Settings(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewStateService(repo)
	userID := int64(12345)

	transitions := []struct {
		name      string
		fromState string
		toState   string
	}{
		{"idle -> awaiting_new_name", domain.StateIdle, domain.StateAwaitingNewName},
		{"awaiting_new_name -> idle", domain.StateAwaitingNewName, domain.StateIdle},
		{"idle -> awaiting_new_emoji", domain.StateIdle, domain.StateAwaitingNewEmoji},
		{"awaiting_new_emoji -> idle", domain.StateAwaitingNewEmoji, domain.StateIdle},
		{"idle -> awaiting_new_challenge_name", domain.StateIdle, domain.StateAwaitingNewChallengeName},
		{"awaiting_new_challenge_name -> idle", domain.StateAwaitingNewChallengeName, domain.StateIdle},
	}

	for _, tt := range transitions {
		t.Run(tt.name, func(t *testing.T) {
			svc.SetState(userID, tt.fromState)
			svc.SetState(userID, tt.toState)

			state, _ := svc.Get(userID)
			if state.State != tt.toState {
				t.Errorf("Transition %s failed: got %s, want %s", tt.name, state.State, tt.toState)
			}
		})
	}
}

func TestStateTransitions_CancelFromAnyState(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewStateService(repo)
	userID := int64(12345)

	// All non-idle states should be able to transition back to idle (cancel)
	allStates := []string{
		domain.StateAwaitingChallengeName,
		domain.StateAwaitingCreatorName,
		domain.StateAwaitingCreatorEmoji,
		domain.StateAwaitingTaskTitle,
		domain.StateAwaitingTaskImage,
		domain.StateAwaitingTaskDescription,
		domain.StateAwaitingEditTitle,
		domain.StateAwaitingEditDescription,
		domain.StateAwaitingEditImage,
		domain.StateReorderSelectTask,
		domain.StateReorderSelectPosition,
		domain.StateAwaitingChallengeID,
		domain.StateAwaitingParticipantName,
		domain.StateAwaitingParticipantEmoji,
		domain.StateAwaitingNewChallengeName,
		domain.StateAwaitingNewName,
		domain.StateAwaitingNewEmoji,
	}

	for _, state := range allStates {
		t.Run("cancel from "+state, func(t *testing.T) {
			svc.SetState(userID, state)
			svc.Reset(userID)

			currentState, _ := svc.Get(userID)
			if currentState.State != domain.StateIdle {
				t.Errorf("Cancel from %s failed: got %s, want %s", state, currentState.State, domain.StateIdle)
			}
		})
	}
}

func TestAllStatesAreDefined(t *testing.T) {
	// Verify all state constants are unique and properly defined
	allStates := map[string]bool{
		domain.StateIdle:                      true,
		domain.StateAwaitingChallengeName:     true,
		domain.StateAwaitingCreatorName:       true,
		domain.StateAwaitingCreatorEmoji:      true,
		domain.StateAwaitingTaskTitle:         true,
		domain.StateAwaitingTaskImage:         true,
		domain.StateAwaitingTaskDescription:   true,
		domain.StateAwaitingEditTitle:         true,
		domain.StateAwaitingEditDescription:   true,
		domain.StateAwaitingEditImage:         true,
		domain.StateReorderSelectTask:         true,
		domain.StateReorderSelectPosition:     true,
		domain.StateAwaitingChallengeID:       true,
		domain.StateAwaitingParticipantName:   true,
		domain.StateAwaitingParticipantEmoji:  true,
		domain.StateAwaitingNewChallengeName:  true,
		domain.StateAwaitingNewName:           true,
		domain.StateAwaitingNewEmoji:          true,
	}

	if len(allStates) != 18 {
		t.Errorf("Expected 18 unique states, got %d", len(allStates))
	}

	// Verify StateIdle is "idle"
	if domain.StateIdle != "idle" {
		t.Errorf("StateIdle = %q, want %q", domain.StateIdle, "idle")
	}
}
