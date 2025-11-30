package service

import (
	"testing"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

func TestCompletionService_Complete(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	participantSvc := NewParticipantService(repo)
	completionSvc := NewCompletionService(repo)

	// Setup
	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)
	task, _ := taskSvc.Create(challenge.ID, "Task 1", "", "")
	participant, _ := participantSvc.Join(challenge.ID, 12345, "User", "💪")

	// Complete task
	completion, err := completionSvc.Complete(task.ID, participant.ID)
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if completion.TaskID != task.ID {
		t.Errorf("Complete() TaskID = %d, want %d", completion.TaskID, task.ID)
	}

	// Complete again (should be idempotent)
	completion2, err := completionSvc.Complete(task.ID, participant.ID)
	if err != nil {
		t.Fatalf("Complete() second time error = %v", err)
	}
	if completion2.ID != completion.ID {
		t.Error("Complete() should return existing completion")
	}
}

func TestCompletionService_Uncomplete(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	participantSvc := NewParticipantService(repo)
	completionSvc := NewCompletionService(repo)

	// Setup
	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)
	task, _ := taskSvc.Create(challenge.ID, "Task 1", "", "")
	participant, _ := participantSvc.Join(challenge.ID, 12345, "User", "💪")

	// Complete then uncomplete
	completionSvc.Complete(task.ID, participant.ID)
	err := completionSvc.Uncomplete(task.ID, participant.ID)
	if err != nil {
		t.Fatalf("Uncomplete() error = %v", err)
	}

	// Verify uncompleted
	isCompleted, _ := completionSvc.IsCompleted(task.ID, participant.ID)
	if isCompleted {
		t.Error("IsCompleted() = true after Uncomplete(), want false")
	}
}

func TestCompletionService_GetCurrentTaskNum(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	participantSvc := NewParticipantService(repo)
	completionSvc := NewCompletionService(repo)

	// Setup
	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)
	task1, _ := taskSvc.Create(challenge.ID, "Task 1", "", "")
	taskSvc.Create(challenge.ID, "Task 2", "", "")
	task3, _ := taskSvc.Create(challenge.ID, "Task 3", "", "")
	participant, _ := participantSvc.Join(challenge.ID, 12345, "User", "💪")

	tasks, _ := taskSvc.GetByChallengeID(challenge.ID)

	// No completions - current should be 1
	current := completionSvc.GetCurrentTaskNum(participant.ID, tasks)
	if current != 1 {
		t.Errorf("GetCurrentTaskNum() = %d, want 1", current)
	}

	// Complete task 1 - current should be 2
	completionSvc.Complete(task1.ID, participant.ID)
	current = completionSvc.GetCurrentTaskNum(participant.ID, tasks)
	if current != 2 {
		t.Errorf("GetCurrentTaskNum() = %d, want 2", current)
	}

	// Complete task 1 and 3 (gap at 2) - current should be 2
	completionSvc.Complete(task3.ID, participant.ID)
	current = completionSvc.GetCurrentTaskNum(participant.ID, tasks)
	if current != 2 {
		t.Errorf("GetCurrentTaskNum() with gap = %d, want 2", current)
	}
}

func TestCompletionService_IsAllCompleted(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	participantSvc := NewParticipantService(repo)
	completionSvc := NewCompletionService(repo)

	// Setup
	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)
	task1, _ := taskSvc.Create(challenge.ID, "Task 1", "", "")
	task2, _ := taskSvc.Create(challenge.ID, "Task 2", "", "")
	participant, _ := participantSvc.Join(challenge.ID, 12345, "User", "💪")

	// Not all completed
	completionSvc.Complete(task1.ID, participant.ID)
	allDone, _ := completionSvc.IsAllCompleted(participant.ID, 2)
	if allDone {
		t.Error("IsAllCompleted() = true with 1/2 completed, want false")
	}

	// All completed
	completionSvc.Complete(task2.ID, participant.ID)
	allDone, _ = completionSvc.IsAllCompleted(participant.ID, 2)
	if !allDone {
		t.Error("IsAllCompleted() = false with 2/2 completed, want true")
	}
}

func TestCompletionService_EmptyTasks(t *testing.T) {
	repo := setupTestRepo(t)
	completionSvc := NewCompletionService(repo)

	// Empty tasks slice
	tasks := []*domain.Task{}
	current := completionSvc.GetCurrentTaskNum(1, tasks)
	if current != 0 {
		t.Errorf("GetCurrentTaskNum() with empty tasks = %d, want 0", current)
	}

	// IsAllCompleted with 0 tasks
	allDone, _ := completionSvc.IsAllCompleted(1, 0)
	if allDone {
		t.Error("IsAllCompleted() = true with 0 tasks, want false")
	}
}

func TestCompletionService_GetCompletionsByTaskID(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	participantSvc := NewParticipantService(repo)
	completionSvc := NewCompletionService(repo)

	// Setup
	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)
	task, _ := taskSvc.Create(challenge.ID, "Task 1", "", "")
	p1, _ := participantSvc.Join(challenge.ID, 12345, "User1", "💪")
	p2, _ := participantSvc.Join(challenge.ID, 67890, "User2", "🔥")

	// Both complete the task
	completionSvc.Complete(task.ID, p1.ID)
	completionSvc.Complete(task.ID, p2.ID)

	completions, err := completionSvc.GetCompletionsByTaskID(task.ID)
	if err != nil {
		t.Fatalf("GetCompletionsByTaskID() error = %v", err)
	}
	if len(completions) != 2 {
		t.Errorf("GetCompletionsByTaskID() count = %d, want 2", len(completions))
	}
}

func TestCompletionService_GetCompletedTaskIDs(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	participantSvc := NewParticipantService(repo)
	completionSvc := NewCompletionService(repo)

	// Setup
	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345)
	task1, _ := taskSvc.Create(challenge.ID, "Task 1", "", "")
	task2, _ := taskSvc.Create(challenge.ID, "Task 2", "", "")
	taskSvc.Create(challenge.ID, "Task 3", "", "")
	participant, _ := participantSvc.Join(challenge.ID, 12345, "User", "💪")

	// Complete tasks 1 and 2
	completionSvc.Complete(task1.ID, participant.ID)
	completionSvc.Complete(task2.ID, participant.ID)

	taskIDs, err := completionSvc.GetCompletedTaskIDs(participant.ID)
	if err != nil {
		t.Fatalf("GetCompletedTaskIDs() error = %v", err)
	}
	if len(taskIDs) != 2 {
		t.Errorf("GetCompletedTaskIDs() count = %d, want 2", len(taskIDs))
	}
}
