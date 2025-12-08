package service

import (
	"testing"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

func TestTaskService_Create(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)

	// Create challenge first
	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345, 0, false)

	// Create task
	task, err := taskSvc.Create(challenge.ID, "Task 1", "Description", "")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if task.Title != "Task 1" {
		t.Errorf("Create() Title = %q, want %q", task.Title, "Task 1")
	}
	if task.OrderNum != 1 {
		t.Errorf("Create() OrderNum = %d, want 1", task.OrderNum)
	}

	// Create second task
	task2, _ := taskSvc.Create(challenge.ID, "Task 2", "", "")
	if task2.OrderNum != 2 {
		t.Errorf("Second task OrderNum = %d, want 2", task2.OrderNum)
	}
}

func TestTaskService_Create_EmptyTitle(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345, 0, false)

	_, err := taskSvc.Create(challenge.ID, "", "Description", "")
	if err != ErrEmptyTaskTitle {
		t.Errorf("Create() error = %v, want ErrEmptyTaskTitle", err)
	}
}

func TestTaskService_Delete_Renumber(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345, 0, false)

	// Create 3 tasks
	task1, _ := taskSvc.Create(challenge.ID, "Task 1", "", "")
	task2, _ := taskSvc.Create(challenge.ID, "Task 2", "", "")
	taskSvc.Create(challenge.ID, "Task 3", "", "")

	// Delete task 2
	err := taskSvc.Delete(task2.ID, challenge.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Get remaining tasks
	tasks, _ := taskSvc.GetByChallengeID(challenge.ID)
	if len(tasks) != 2 {
		t.Errorf("After delete, got %d tasks, want 2", len(tasks))
	}

	// Check renumbering
	if tasks[0].ID != task1.ID || tasks[0].OrderNum != 1 {
		t.Errorf("Task 1 should have OrderNum 1, got %d", tasks[0].OrderNum)
	}
	if tasks[1].OrderNum != 2 {
		t.Errorf("Task 3 should be renumbered to 2, got %d", tasks[1].OrderNum)
	}
}

func TestTaskService_MoveTask(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345, 0, false)

	// Create 3 tasks: 1, 2, 3
	task1, _ := taskSvc.Create(challenge.ID, "Task 1", "", "")
	taskSvc.Create(challenge.ID, "Task 2", "", "")
	task3, _ := taskSvc.Create(challenge.ID, "Task 3", "", "")

	// Move task 3 to position 1
	err := taskSvc.MoveTask(task3.ID, challenge.ID, 1)
	if err != nil {
		t.Fatalf("MoveTask() error = %v", err)
	}

	// Check new order
	tasks, _ := taskSvc.GetByChallengeID(challenge.ID)

	// New order should be: Task 3, Task 1, Task 2
	if tasks[0].ID != task3.ID {
		t.Errorf("Position 1 should be Task 3")
	}
	if tasks[1].ID != task1.ID {
		t.Errorf("Position 2 should be Task 1")
	}
}

func TestTaskService_MaxTasks(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345, 0, false)

	// Create 50 tasks (max)
	for i := 0; i < domain.MaxTasksPerChallenge; i++ {
		_, err := taskSvc.Create(challenge.ID, "Task", "", "")
		if err != nil {
			t.Fatalf("Create() %d error = %v", i, err)
		}
	}

	// 51st should fail
	_, err := taskSvc.Create(challenge.ID, "One More", "", "")
	if err != ErrMaxTasksReached {
		t.Errorf("Create() error = %v, want ErrMaxTasksReached", err)
	}
}

func TestTaskService_Update(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345, 0, false)
	task, _ := taskSvc.Create(challenge.ID, "Original Title", "", "")

	// Update task
	task.Title = "Updated Title"
	task.Description = "New Description"
	err := taskSvc.Update(task)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update
	updated, _ := taskSvc.GetByID(task.ID)
	if updated.Title != "Updated Title" {
		t.Errorf("Title = %q, want %q", updated.Title, "Updated Title")
	}
	if updated.Description != "New Description" {
		t.Errorf("Description = %q, want %q", updated.Description, "New Description")
	}
}

func TestTaskService_CountByChallengeID(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345, 0, false)

	// No tasks yet
	count, err := taskSvc.CountByChallengeID(challenge.ID)
	if err != nil {
		t.Fatalf("CountByChallengeID() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Count = %d, want 0", count)
	}

	// Create some tasks
	taskSvc.Create(challenge.ID, "Task 1", "", "")
	taskSvc.Create(challenge.ID, "Task 2", "", "")
	taskSvc.Create(challenge.ID, "Task 3", "", "")

	count, err = taskSvc.CountByChallengeID(challenge.ID)
	if err != nil {
		t.Fatalf("CountByChallengeID() error = %v", err)
	}
	if count != 3 {
		t.Errorf("Count = %d, want 3", count)
	}
}

func TestTaskService_GetByID(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)

	challenge, _ := challengeSvc.Create("Test Challenge", "", 12345, 0, false)
	task, _ := taskSvc.Create(challenge.ID, "Test Task", "Description", "")

	// Get existing task
	found, err := taskSvc.GetByID(task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found.Title != "Test Task" {
		t.Errorf("Title = %q, want %q", found.Title, "Test Task")
	}

	// Get non-existing task
	_, err = taskSvc.GetByID(99999)
	if err != ErrTaskNotFound {
		t.Errorf("GetByID() for non-existing: error = %v, want ErrTaskNotFound", err)
	}
}
