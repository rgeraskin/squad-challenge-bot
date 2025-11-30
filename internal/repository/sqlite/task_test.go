package sqlite

import (
	"testing"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

func TestTaskRepo_Create(t *testing.T) {
	repo := setupTestDB(t)

	// Create challenge first
	challenge := &domain.Challenge{ID: "TEST1234", Name: "Test", CreatorID: 12345}
	repo.Challenge().Create(challenge)

	task := &domain.Task{
		ChallengeID: "TEST1234",
		OrderNum:    1,
		Title:       "Test Task",
		Description: "Test Description",
	}

	err := repo.Task().Create(task)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if task.ID == 0 {
		t.Error("Create() should set task ID")
	}
}

func TestTaskRepo_GetByChallengeID(t *testing.T) {
	repo := setupTestDB(t)

	// Create challenge and tasks
	challenge := &domain.Challenge{ID: "TEST1234", Name: "Test", CreatorID: 12345}
	repo.Challenge().Create(challenge)

	task1 := &domain.Task{ChallengeID: "TEST1234", OrderNum: 1, Title: "Task 1"}
	task2 := &domain.Task{ChallengeID: "TEST1234", OrderNum: 2, Title: "Task 2"}
	task3 := &domain.Task{ChallengeID: "TEST1234", OrderNum: 3, Title: "Task 3"}
	repo.Task().Create(task1)
	repo.Task().Create(task2)
	repo.Task().Create(task3)

	tasks, err := repo.Task().GetByChallengeID("TEST1234")
	if err != nil {
		t.Fatalf("GetByChallengeID() error = %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("GetByChallengeID() returned %d tasks, want 3", len(tasks))
	}

	// Verify order
	for i, task := range tasks {
		if task.OrderNum != i+1 {
			t.Errorf("Task at index %d has OrderNum %d, want %d", i, task.OrderNum, i+1)
		}
	}
}

func TestTaskRepo_GetMaxOrderNum(t *testing.T) {
	repo := setupTestDB(t)

	// Create challenge
	challenge := &domain.Challenge{ID: "TEST1234", Name: "Test", CreatorID: 12345}
	repo.Challenge().Create(challenge)

	// No tasks yet
	max, err := repo.Task().GetMaxOrderNum("TEST1234")
	if err != nil {
		t.Fatalf("GetMaxOrderNum() error = %v", err)
	}
	if max != 0 {
		t.Errorf("GetMaxOrderNum() = %d, want 0 for empty challenge", max)
	}

	// Add tasks
	task1 := &domain.Task{ChallengeID: "TEST1234", OrderNum: 1, Title: "Task 1"}
	task2 := &domain.Task{ChallengeID: "TEST1234", OrderNum: 5, Title: "Task 5"}
	repo.Task().Create(task1)
	repo.Task().Create(task2)

	max, err = repo.Task().GetMaxOrderNum("TEST1234")
	if err != nil {
		t.Fatalf("GetMaxOrderNum() error = %v", err)
	}
	if max != 5 {
		t.Errorf("GetMaxOrderNum() = %d, want 5", max)
	}
}

func TestTaskRepo_Delete(t *testing.T) {
	repo := setupTestDB(t)

	// Create challenge and task
	challenge := &domain.Challenge{ID: "TEST1234", Name: "Test", CreatorID: 12345}
	repo.Challenge().Create(challenge)

	task := &domain.Task{ChallengeID: "TEST1234", OrderNum: 1, Title: "Task 1"}
	repo.Task().Create(task)
	taskID := task.ID

	err := repo.Task().Delete(taskID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	got, _ := repo.Task().GetByID(taskID)
	if got != nil {
		t.Error("GetByID() should return nil after Delete()")
	}
}

func TestTaskRepo_CascadeDelete(t *testing.T) {
	repo := setupTestDB(t)

	// Create challenge and task
	challenge := &domain.Challenge{ID: "TEST1234", Name: "Test", CreatorID: 12345}
	repo.Challenge().Create(challenge)

	task := &domain.Task{ChallengeID: "TEST1234", OrderNum: 1, Title: "Task 1"}
	repo.Task().Create(task)

	// Delete challenge (should cascade)
	repo.Challenge().Delete("TEST1234")

	// Verify tasks deleted
	tasks, _ := repo.Task().GetByChallengeID("TEST1234")
	if len(tasks) != 0 {
		t.Error("Tasks should be deleted when challenge is deleted")
	}
}
