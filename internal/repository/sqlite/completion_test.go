package sqlite

import (
	"testing"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

func TestCompletionRepo_Create(t *testing.T) {
	repo := setupTestDB(t)

	// Setup: challenge, task, participant
	challenge := &domain.Challenge{ID: "TEST1234", Name: "Test", CreatorID: 12345}
	repo.Challenge().Create(challenge)

	task := &domain.Task{ChallengeID: "TEST1234", OrderNum: 1, Title: "Task 1"}
	repo.Task().Create(task)

	participant := &domain.Participant{ChallengeID: "TEST1234", TelegramID: 12345, DisplayName: "User", Emoji: "💪", NotifyEnabled: true}
	repo.Participant().Create(participant)

	// Create completion
	completion := &domain.TaskCompletion{
		TaskID:        task.ID,
		ParticipantID: participant.ID,
	}

	err := repo.Completion().Create(completion)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if completion.ID == 0 {
		t.Error("Create() should set completion ID")
	}
}

func TestCompletionRepo_GetByTaskAndParticipant(t *testing.T) {
	repo := setupTestDB(t)

	// Setup
	challenge := &domain.Challenge{ID: "TEST1234", Name: "Test", CreatorID: 12345}
	repo.Challenge().Create(challenge)

	task := &domain.Task{ChallengeID: "TEST1234", OrderNum: 1, Title: "Task 1"}
	repo.Task().Create(task)

	participant := &domain.Participant{ChallengeID: "TEST1234", TelegramID: 12345, DisplayName: "User", Emoji: "💪", NotifyEnabled: true}
	repo.Participant().Create(participant)

	// No completion yet
	got, err := repo.Completion().GetByTaskAndParticipant(task.ID, participant.ID)
	if err != nil {
		t.Fatalf("GetByTaskAndParticipant() error = %v", err)
	}
	if got != nil {
		t.Error("GetByTaskAndParticipant() should return nil before completion")
	}

	// Create completion
	completion := &domain.TaskCompletion{TaskID: task.ID, ParticipantID: participant.ID}
	repo.Completion().Create(completion)

	// Now should find it
	got, err = repo.Completion().GetByTaskAndParticipant(task.ID, participant.ID)
	if err != nil {
		t.Fatalf("GetByTaskAndParticipant() error = %v", err)
	}
	if got == nil {
		t.Error("GetByTaskAndParticipant() should return completion after creation")
	}
}

func TestCompletionRepo_Delete(t *testing.T) {
	repo := setupTestDB(t)

	// Setup
	challenge := &domain.Challenge{ID: "TEST1234", Name: "Test", CreatorID: 12345}
	repo.Challenge().Create(challenge)

	task := &domain.Task{ChallengeID: "TEST1234", OrderNum: 1, Title: "Task 1"}
	repo.Task().Create(task)

	participant := &domain.Participant{ChallengeID: "TEST1234", TelegramID: 12345, DisplayName: "User", Emoji: "💪", NotifyEnabled: true}
	repo.Participant().Create(participant)

	completion := &domain.TaskCompletion{TaskID: task.ID, ParticipantID: participant.ID}
	repo.Completion().Create(completion)

	// Delete
	err := repo.Completion().Delete(task.ID, participant.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	got, _ := repo.Completion().GetByTaskAndParticipant(task.ID, participant.ID)
	if got != nil {
		t.Error("GetByTaskAndParticipant() should return nil after Delete()")
	}
}

func TestCompletionRepo_GetCompletedTaskIDs(t *testing.T) {
	repo := setupTestDB(t)

	// Setup
	challenge := &domain.Challenge{ID: "TEST1234", Name: "Test", CreatorID: 12345}
	repo.Challenge().Create(challenge)

	task1 := &domain.Task{ChallengeID: "TEST1234", OrderNum: 1, Title: "Task 1"}
	task2 := &domain.Task{ChallengeID: "TEST1234", OrderNum: 2, Title: "Task 2"}
	task3 := &domain.Task{ChallengeID: "TEST1234", OrderNum: 3, Title: "Task 3"}
	repo.Task().Create(task1)
	repo.Task().Create(task2)
	repo.Task().Create(task3)

	participant := &domain.Participant{ChallengeID: "TEST1234", TelegramID: 12345, DisplayName: "User", Emoji: "💪", NotifyEnabled: true}
	repo.Participant().Create(participant)

	// Complete tasks 1 and 3
	repo.Completion().Create(&domain.TaskCompletion{TaskID: task1.ID, ParticipantID: participant.ID})
	repo.Completion().Create(&domain.TaskCompletion{TaskID: task3.ID, ParticipantID: participant.ID})

	ids, err := repo.Completion().GetCompletedTaskIDs(participant.ID)
	if err != nil {
		t.Fatalf("GetCompletedTaskIDs() error = %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("GetCompletedTaskIDs() returned %d IDs, want 2", len(ids))
	}
}

func TestCompletionRepo_CascadeDeleteOnTask(t *testing.T) {
	repo := setupTestDB(t)

	// Setup
	challenge := &domain.Challenge{ID: "TEST1234", Name: "Test", CreatorID: 12345}
	repo.Challenge().Create(challenge)

	task := &domain.Task{ChallengeID: "TEST1234", OrderNum: 1, Title: "Task 1"}
	repo.Task().Create(task)

	participant := &domain.Participant{ChallengeID: "TEST1234", TelegramID: 12345, DisplayName: "User", Emoji: "💪", NotifyEnabled: true}
	repo.Participant().Create(participant)

	completion := &domain.TaskCompletion{TaskID: task.ID, ParticipantID: participant.ID}
	repo.Completion().Create(completion)

	// Delete task (should cascade to completions)
	repo.Task().Delete(task.ID)

	// Verify completions deleted
	ids, _ := repo.Completion().GetCompletedTaskIDs(participant.ID)
	if len(ids) != 0 {
		t.Error("Completions should be deleted when task is deleted")
	}
}
