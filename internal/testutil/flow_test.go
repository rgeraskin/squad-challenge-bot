package testutil

import (
	"testing"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/service"
)

// Flow Integration Tests - test complete user scenarios

func TestFlow_CreateChallengeAndAddTasks(t *testing.T) {
	f := NewFlowRunner(t)

	creatorID := int64(12345)

	// Step 1: Create challenge
	challenge, err := f.Challenge.Create("30-Day Fitness", "", creatorID, 0)
	if err != nil {
		t.Fatalf("Create challenge failed: %v", err)
	}
	if len(challenge.ID) != 8 {
		t.Errorf("Challenge ID length = %d, want 8", len(challenge.ID))
	}

	// Step 2: Creator joins as participant
	participant, err := f.Participant.Join(challenge.ID, creatorID, "John", "💪", 0)
	if err != nil {
		t.Fatalf("Join challenge failed: %v", err)
	}

	// Step 3: Add tasks
	task1, _ := f.Task.Create(challenge.ID, "Morning Stretch", "Stretch for 10 minutes", "")
	task2, _ := f.Task.Create(challenge.ID, "50 Squats", "", "")
	task3, _ := f.Task.Create(challenge.ID, "100 Push-ups", "", "")

	// Verify task order
	tasks, _ := f.Task.GetByChallengeID(challenge.ID)
	if len(tasks) != 3 {
		t.Errorf("Task count = %d, want 3", len(tasks))
	}
	if tasks[0].Title != "Morning Stretch" {
		t.Errorf("First task = %q, want %q", tasks[0].Title, "Morning Stretch")
	}

	// Step 4: Verify current task
	currentTask := f.Completion.GetCurrentTaskNum(participant.ID, tasks)
	if currentTask != 1 {
		t.Errorf("Current task = %d, want 1", currentTask)
	}

	// Use task IDs to avoid "declared but not used"
	_ = task1
	_ = task2
	_ = task3
}

func TestFlow_JoinChallengeByID(t *testing.T) {
	f := NewFlowRunner(t)

	creatorID := int64(12345)
	joinerID := int64(67890)

	// Creator sets up challenge
	challengeID := f.CreateChallengeWithTasks(creatorID, "Reading Challenge", 5)

	// Joiner validates challenge exists
	err := f.Challenge.CanJoin(challengeID, joinerID)
	if err != nil {
		t.Fatalf("CanJoin failed: %v", err)
	}

	// Joiner joins
	participant, err := f.Participant.Join(challengeID, joinerID, "Sarah", "🔥", 0)
	if err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	// Verify participant count
	count, _ := f.Participant.CountByChallengeID(challengeID)
	if count != 2 { // Creator + Joiner
		t.Errorf("Participant count = %d, want 2", count)
	}

	// Verify joiner starts at task 1
	tasks, _ := f.Task.GetByChallengeID(challengeID)
	currentTask := f.Completion.GetCurrentTaskNum(participant.ID, tasks)
	if currentTask != 1 {
		t.Errorf("Current task = %d, want 1", currentTask)
	}
}

func TestFlow_CompleteTasksProgressively(t *testing.T) {
	f := NewFlowRunner(t)

	userID := int64(12345)
	challengeID := f.CreateChallengeWithTasks(userID, "Test Challenge", 5)

	participant, _ := f.Participant.GetByChallengeAndUser(challengeID, userID)
	tasks, _ := f.Task.GetByChallengeID(challengeID)

	// Complete tasks 1, 2, 3 in order
	for i := 0; i < 3; i++ {
		f.Completion.Complete(tasks[i].ID, participant.ID)
		f.AssertProgress(challengeID, participant.ID, i+1, 5)

		// Verify current task advances
		currentTask := f.Completion.GetCurrentTaskNum(participant.ID, tasks)
		expectedCurrent := i + 2
		if i == 2 {
			expectedCurrent = 4 // After completing 3, current is 4
		}
		if currentTask != expectedCurrent {
			t.Errorf("After completing task %d, current = %d, want %d", i+1, currentTask, expectedCurrent)
		}
	}
}

func TestFlow_CompleteTasksOutOfOrder(t *testing.T) {
	f := NewFlowRunner(t)

	userID := int64(12345)
	challengeID := f.CreateChallengeWithTasks(userID, "Test Challenge", 5)

	participant, _ := f.Participant.GetByChallengeAndUser(challengeID, userID)
	tasks, _ := f.Task.GetByChallengeID(challengeID)

	// Complete tasks 1 and 3 (skip 2)
	f.Completion.Complete(tasks[0].ID, participant.ID)
	f.Completion.Complete(tasks[2].ID, participant.ID)

	// Current task should be 4 (next after last completed, which is 3)
	currentTask := f.Completion.GetCurrentTaskNum(participant.ID, tasks)
	if currentTask != 4 {
		t.Errorf("Current task with gap = %d, want 4", currentTask)
	}

	// Complete task 2 (filling the gap doesn't change current)
	f.Completion.Complete(tasks[1].ID, participant.ID)

	// Current should still be 4
	currentTask = f.Completion.GetCurrentTaskNum(participant.ID, tasks)
	if currentTask != 4 {
		t.Errorf("Current task after filling gap = %d, want 4", currentTask)
	}
}

func TestFlow_ChallengeCompletion(t *testing.T) {
	f := NewFlowRunner(t)

	userID := int64(12345)
	challengeID := f.CreateChallengeWithTasks(userID, "Test Challenge", 3)

	participant, _ := f.Participant.GetByChallengeAndUser(challengeID, userID)
	tasks, _ := f.Task.GetByChallengeID(challengeID)

	// Not completed yet
	allDone, _ := f.Completion.IsAllCompleted(participant.ID, len(tasks))
	if allDone {
		t.Error("Should not be completed yet")
	}

	// Complete all tasks
	for _, task := range tasks {
		f.Completion.Complete(task.ID, participant.ID)
	}

	// Now should be completed
	allDone, _ = f.Completion.IsAllCompleted(participant.ID, len(tasks))
	if !allDone {
		t.Error("Should be completed after all tasks done")
	}

	// Current task should be 0 (all done)
	currentTask := f.Completion.GetCurrentTaskNum(participant.ID, tasks)
	if currentTask != 0 {
		t.Errorf("Current task when all done = %d, want 0", currentTask)
	}
}

func TestFlow_UncompleteTask(t *testing.T) {
	f := NewFlowRunner(t)

	userID := int64(12345)
	challengeID := f.CreateChallengeWithTasks(userID, "Test Challenge", 3)

	participant, _ := f.Participant.GetByChallengeAndUser(challengeID, userID)
	tasks, _ := f.Task.GetByChallengeID(challengeID)

	// Complete all tasks
	for _, task := range tasks {
		f.Completion.Complete(task.ID, participant.ID)
	}

	// Verify all completed
	f.AssertProgress(challengeID, participant.ID, 3, 3)

	// Uncomplete task 2
	f.Completion.Uncomplete(tasks[1].ID, participant.ID)

	// Verify progress decreased
	f.AssertProgress(challengeID, participant.ID, 2, 3)

	// Current task should be 2 now
	currentTask := f.Completion.GetCurrentTaskNum(participant.ID, tasks)
	if currentTask != 2 {
		t.Errorf("Current task after uncomplete = %d, want 2", currentTask)
	}
}

func TestFlow_MultipleParticipants(t *testing.T) {
	f := NewFlowRunner(t)

	user1ID := int64(11111)
	user2ID := int64(22222)
	user3ID := int64(33333)

	challengeID := f.CreateChallengeWithTasks(user1ID, "Team Challenge", 5)

	// Others join
	p2 := f.JoinChallenge(challengeID, user2ID, "User2", "🔥")
	p3 := f.JoinChallenge(challengeID, user3ID, "User3", "⭐")

	participant1, _ := f.Participant.GetByChallengeAndUser(challengeID, user1ID)

	// Each user completes different amounts
	f.CompleteTasks(challengeID, participant1.ID, 1, 2, 3, 4) // 4 tasks
	f.CompleteTasks(challengeID, p2, 1, 2, 3)                 // 3 tasks
	f.CompleteTasks(challengeID, p3, 1, 2)                    // 2 tasks

	// Verify each user's progress
	f.AssertProgress(challengeID, participant1.ID, 4, 5)
	f.AssertProgress(challengeID, p2, 3, 5)
	f.AssertProgress(challengeID, p3, 2, 5)

	// Verify participant count
	count, _ := f.Participant.CountByChallengeID(challengeID)
	if count != 3 {
		t.Errorf("Participant count = %d, want 3", count)
	}
}

func TestFlow_LeaveChallenge(t *testing.T) {
	f := NewFlowRunner(t)

	creatorID := int64(12345)
	leaverID := int64(67890)

	challengeID := f.CreateChallengeWithTasks(creatorID, "Test Challenge", 3)
	leaverParticipantID := f.JoinChallenge(challengeID, leaverID, "Leaver", "🚪")

	// Complete some tasks
	f.CompleteTasks(challengeID, leaverParticipantID, 1, 2)

	// Leave challenge
	err := f.Participant.Leave(leaverParticipantID)
	if err != nil {
		t.Fatalf("Leave failed: %v", err)
	}

	// Verify participant is gone
	participant, _ := f.Participant.GetByChallengeAndUser(challengeID, leaverID)
	if participant != nil {
		t.Error("Participant should be nil after leaving")
	}

	// Verify count decreased
	count, _ := f.Participant.CountByChallengeID(challengeID)
	if count != 1 {
		t.Errorf("Participant count = %d, want 1", count)
	}
}

func TestFlow_DeleteChallenge(t *testing.T) {
	f := NewFlowRunner(t)

	creatorID := int64(12345)
	memberID := int64(67890)

	challengeID := f.CreateChallengeWithTasks(creatorID, "Doomed Challenge", 5)
	f.JoinChallenge(challengeID, memberID, "Member", "🔥")

	// Delete challenge
	err := f.Challenge.Delete(challengeID, creatorID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify challenge is gone
	_, err = f.Challenge.GetByID(challengeID)
	if err != service.ErrChallengeNotFound {
		t.Errorf("GetByID after delete: error = %v, want ErrChallengeNotFound", err)
	}

	// Verify tasks cascaded
	tasks, _ := f.Task.GetByChallengeID(challengeID)
	if len(tasks) != 0 {
		t.Error("Tasks should be deleted with challenge")
	}

	// Verify participants cascaded
	count, _ := f.Participant.CountByChallengeID(challengeID)
	if count != 0 {
		t.Error("Participants should be deleted with challenge")
	}
}

func TestFlow_ReorderTasks(t *testing.T) {
	f := NewFlowRunner(t)

	userID := int64(12345)
	challengeID := f.CreateChallengeWithTasks(userID, "Test Challenge", 5)

	tasks, _ := f.Task.GetByChallengeID(challengeID)
	task5ID := tasks[4].ID // Task at position 5

	// Move task 5 to position 1
	err := f.Task.MoveTask(task5ID, challengeID, 1)
	if err != nil {
		t.Fatalf("MoveTask failed: %v", err)
	}

	// Verify new order
	tasks, _ = f.Task.GetByChallengeID(challengeID)
	if tasks[0].ID != task5ID {
		t.Error("Task 5 should now be at position 1")
	}
}

func TestFlow_StateManagementDuringFlow(t *testing.T) {
	f := NewFlowRunner(t)

	userID := int64(12345)

	// Simulate create challenge flow with state transitions
	// Step 1: Start create flow
	f.State.SetState(userID, domain.StateAwaitingChallengeName)
	state, _ := f.State.Get(userID)
	if state.State != domain.StateAwaitingChallengeName {
		t.Errorf("State = %q, want %q", state.State, domain.StateAwaitingChallengeName)
	}

	// Step 2: Store challenge name, move to creator name
	f.State.SetStateWithData(userID, domain.StateAwaitingCreatorName, map[string]string{
		"challenge_name": "My Challenge",
	})

	// Step 3: Store creator name, move to emoji
	var tempData map[string]interface{}
	f.State.GetTempData(userID, &tempData)
	tempData["display_name"] = "John"
	f.State.SetStateWithData(userID, domain.StateAwaitingCreatorEmoji, tempData)

	// Step 4: Complete flow
	challenge, _ := f.Challenge.Create("My Challenge", "", userID, 0)
	f.Participant.Join(challenge.ID, userID, "John", "💪", 0)
	f.State.SetCurrentChallenge(userID, challenge.ID)
	f.State.ResetKeepChallenge(userID)

	// Verify final state
	state, _ = f.State.Get(userID)
	if state.State != domain.StateIdle {
		t.Errorf("Final state = %q, want %q", state.State, domain.StateIdle)
	}
	if state.CurrentChallenge != challenge.ID {
		t.Errorf("Current challenge = %q, want %q", state.CurrentChallenge, challenge.ID)
	}
}

func TestFlow_ChallengeFull(t *testing.T) {
	f := NewFlowRunner(t)

	creatorID := int64(10000)
	challengeID := f.CreateChallengeWithTasks(creatorID, "Full Challenge", 3)

	// Different emojis for each participant (emoji uniqueness per challenge)
	emojis := []string{"🔥", "⭐", "🎯", "🏆", "🎨", "🎪", "🎭", "🎮", "🎲"}

	// Add 9 more participants (total 10 - max)
	for i := 1; i < 10; i++ {
		userID := int64(10000 + i)
		_, err := f.Participant.Join(challengeID, userID, "User", emojis[i-1], 0)
		if err != nil {
			t.Fatalf("Failed to add participant %d: %v", i, err)
		}
	}

	// Verify count is 10
	count, _ := f.Participant.CountByChallengeID(challengeID)
	if count != 10 {
		t.Errorf("Participant count = %d, want 10", count)
	}

	// 11th participant should fail
	err := f.Challenge.CanJoin(challengeID, 99999)
	if err != service.ErrChallengeFull {
		t.Errorf("CanJoin for 11th participant: error = %v, want ErrChallengeFull", err)
	}
}
