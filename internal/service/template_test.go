package service

import (
	"testing"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

// Note: This file uses domain.TemplateTask but not domain.Task
// The taskSvc.Create uses the service method signature (challengeID, title, description, imageFileID string)

func TestTemplateService_CreateFromChallenge(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	templateSvc := NewTemplateService(repo)

	// Create a challenge with tasks
	challenge, err := challengeSvc.Create("Test Challenge", "Description", 12345, 5, true)
	if err != nil {
		t.Fatalf("Create challenge error = %v", err)
	}

	// Add tasks
	taskSvc.Create(challenge.ID, "Task 1", "Desc 1", "")
	taskSvc.Create(challenge.ID, "Task 2", "", "img123")

	// Create template from challenge
	template, err := templateSvc.CreateFromChallenge(challenge.ID)
	if err != nil {
		t.Fatalf("CreateFromChallenge() error = %v", err)
	}

	if template.Name != challenge.Name {
		t.Errorf("Template Name = %q, want %q", template.Name, challenge.Name)
	}
	if template.Description != challenge.Description {
		t.Errorf("Template Description = %q, want %q", template.Description, challenge.Description)
	}
	if template.DailyTaskLimit != challenge.DailyTaskLimit {
		t.Errorf("Template DailyTaskLimit = %d, want %d", template.DailyTaskLimit, challenge.DailyTaskLimit)
	}
	if template.HideFutureTasks != challenge.HideFutureTasks {
		t.Errorf("Template HideFutureTasks = %v, want %v", template.HideFutureTasks, challenge.HideFutureTasks)
	}

	// Verify tasks were copied
	tasks, err := templateSvc.GetTasks(template.ID)
	if err != nil {
		t.Fatalf("GetTasks() error = %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("GetTasks() count = %d, want 2", len(tasks))
	}

	// Verify task details
	if tasks[0].Title != "Task 1" {
		t.Errorf("Task[0] Title = %q, want %q", tasks[0].Title, "Task 1")
	}
	if tasks[1].ImageFileID != "img123" {
		t.Errorf("Task[1] ImageFileID = %q, want %q", tasks[1].ImageFileID, "img123")
	}
}

func TestTemplateService_CreateFromChallenge_DuplicateName(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Duplicate Name", "", 12345, 0, false)

	// Create first template
	_, err := templateSvc.CreateFromChallenge(challenge.ID)
	if err != nil {
		t.Fatalf("First CreateFromChallenge() error = %v", err)
	}

	// Try to create duplicate
	_, err = templateSvc.CreateFromChallenge(challenge.ID)
	if err != ErrTemplateNameExists {
		t.Errorf("Second CreateFromChallenge() error = %v, want ErrTemplateNameExists", err)
	}
}

func TestTemplateService_CreateFromChallenge_NotFound(t *testing.T) {
	repo := setupTestRepo(t)
	templateSvc := NewTemplateService(repo)

	_, err := templateSvc.CreateFromChallenge("NOTEXIST")
	if err != ErrChallengeNotFound {
		t.Errorf("CreateFromChallenge(non-existent) error = %v, want ErrChallengeNotFound", err)
	}
}

func TestTemplateService_GetByID(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Test", "", 12345, 0, false)
	created, _ := templateSvc.CreateFromChallenge(challenge.ID)

	got, err := templateSvc.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "Test" {
		t.Errorf("GetByID() Name = %q, want %q", got.Name, "Test")
	}

	// Non-existent
	_, err = templateSvc.GetByID(99999)
	if err != ErrTemplateNotFound {
		t.Errorf("GetByID(non-existent) error = %v, want ErrTemplateNotFound", err)
	}
}

func TestTemplateService_GetAll(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	templateSvc := NewTemplateService(repo)

	// No templates yet
	templates, err := templateSvc.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(templates) != 0 {
		t.Errorf("GetAll() count = %d, want 0", len(templates))
	}

	// Create templates
	ch1, _ := challengeSvc.Create("Template 1", "", 12345, 0, false)
	ch2, _ := challengeSvc.Create("Template 2", "", 12345, 0, false)
	templateSvc.CreateFromChallenge(ch1.ID)
	templateSvc.CreateFromChallenge(ch2.ID)

	templates, err = templateSvc.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(templates) != 2 {
		t.Errorf("GetAll() count = %d, want 2", len(templates))
	}
}

func TestTemplateService_Delete(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Test", "", 12345, 0, false)
	template, _ := templateSvc.CreateFromChallenge(challenge.ID)

	err := templateSvc.Delete(template.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = templateSvc.GetByID(template.ID)
	if err != ErrTemplateNotFound {
		t.Errorf("GetByID() after delete: error = %v, want ErrTemplateNotFound", err)
	}
}

func TestTemplateService_UpdateName(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Original", "", 12345, 0, false)
	template, _ := templateSvc.CreateFromChallenge(challenge.ID)

	err := templateSvc.UpdateName(template.ID, "Updated Name")
	if err != nil {
		t.Fatalf("UpdateName() error = %v", err)
	}

	updated, _ := templateSvc.GetByID(template.ID)
	if updated.Name != "Updated Name" {
		t.Errorf("UpdateName() Name = %q, want %q", updated.Name, "Updated Name")
	}
}

func TestTemplateService_UpdateDescription(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Test", "", 12345, 0, false)
	template, _ := templateSvc.CreateFromChallenge(challenge.ID)

	err := templateSvc.UpdateDescription(template.ID, "New Description")
	if err != nil {
		t.Fatalf("UpdateDescription() error = %v", err)
	}

	updated, _ := templateSvc.GetByID(template.ID)
	if updated.Description != "New Description" {
		t.Errorf("UpdateDescription() Description = %q, want %q", updated.Description, "New Description")
	}
}

func TestTemplateService_UpdateDailyLimit(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Test", "", 12345, 0, false)
	template, _ := templateSvc.CreateFromChallenge(challenge.ID)

	err := templateSvc.UpdateDailyLimit(template.ID, 10)
	if err != nil {
		t.Fatalf("UpdateDailyLimit() error = %v", err)
	}

	updated, _ := templateSvc.GetByID(template.ID)
	if updated.DailyTaskLimit != 10 {
		t.Errorf("UpdateDailyLimit() DailyTaskLimit = %d, want %d", updated.DailyTaskLimit, 10)
	}
}

func TestTemplateService_UpdateHideFutureTasks(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Test", "", 12345, 0, false)
	template, _ := templateSvc.CreateFromChallenge(challenge.ID)

	err := templateSvc.UpdateHideFutureTasks(template.ID, true)
	if err != nil {
		t.Fatalf("UpdateHideFutureTasks() error = %v", err)
	}

	updated, _ := templateSvc.GetByID(template.ID)
	if !updated.HideFutureTasks {
		t.Error("UpdateHideFutureTasks() HideFutureTasks should be true")
	}
}

func TestTemplateService_CreateTask(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Test", "", 12345, 0, false)
	template, _ := templateSvc.CreateFromChallenge(challenge.ID)

	task := &domain.TemplateTask{
		TemplateID:  template.ID,
		Title:       "New Task",
		Description: "Task Description",
	}
	err := templateSvc.CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if task.OrderNum != 1 {
		t.Errorf("CreateTask() OrderNum = %d, want 1", task.OrderNum)
	}

	// Create another task
	task2 := &domain.TemplateTask{
		TemplateID: template.ID,
		Title:      "Second Task",
	}
	templateSvc.CreateTask(task2)
	if task2.OrderNum != 2 {
		t.Errorf("CreateTask() second task OrderNum = %d, want 2", task2.OrderNum)
	}
}

func TestTemplateService_DeleteTask(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Test", "", 12345, 0, false)
	taskSvc.Create(challenge.ID, "Task 1", "", "")
	taskSvc.Create(challenge.ID, "Task 2", "", "")
	taskSvc.Create(challenge.ID, "Task 3", "", "")

	template, _ := templateSvc.CreateFromChallenge(challenge.ID)

	tasks, _ := templateSvc.GetTasks(template.ID)
	if len(tasks) != 3 {
		t.Fatalf("Initial task count = %d, want 3", len(tasks))
	}

	// Delete middle task
	err := templateSvc.DeleteTask(tasks[1].ID, template.ID)
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	// Verify renumbering
	remaining, _ := templateSvc.GetTasks(template.ID)
	if len(remaining) != 2 {
		t.Errorf("After delete count = %d, want 2", len(remaining))
	}
	if remaining[0].OrderNum != 1 {
		t.Errorf("Task[0] OrderNum = %d, want 1", remaining[0].OrderNum)
	}
	if remaining[1].OrderNum != 2 {
		t.Errorf("Task[1] OrderNum = %d, want 2", remaining[1].OrderNum)
	}
}

func TestTemplateService_MoveTask(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Test", "", 12345, 0, false)
	taskSvc.Create(challenge.ID, "Task A", "", "")
	taskSvc.Create(challenge.ID, "Task B", "", "")
	taskSvc.Create(challenge.ID, "Task C", "", "")

	template, _ := templateSvc.CreateFromChallenge(challenge.ID)

	tasks, _ := templateSvc.GetTasks(template.ID)
	// Move Task C (position 3) to position 1
	err := templateSvc.MoveTask(tasks[2].ID, template.ID, 1)
	if err != nil {
		t.Fatalf("MoveTask() error = %v", err)
	}

	reordered, _ := templateSvc.GetTasks(template.ID)
	// Should now be: C, A, B
	if reordered[0].Title != "Task C" {
		t.Errorf("After move, Task[0] = %q, want %q", reordered[0].Title, "Task C")
	}
	if reordered[1].Title != "Task A" {
		t.Errorf("After move, Task[1] = %q, want %q", reordered[1].Title, "Task A")
	}
	if reordered[2].Title != "Task B" {
		t.Errorf("After move, Task[2] = %q, want %q", reordered[2].Title, "Task B")
	}
}

func TestTemplateService_MoveTask_InvalidPosition(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Test", "", 12345, 0, false)
	taskSvc.Create(challenge.ID, "Task 1", "", "")

	template, _ := templateSvc.CreateFromChallenge(challenge.ID)
	tasks, _ := templateSvc.GetTasks(template.ID)

	err := templateSvc.MoveTask(tasks[0].ID, template.ID, 5)
	if err == nil {
		t.Error("MoveTask() with invalid position should return error")
	}

	err = templateSvc.MoveTask(tasks[0].ID, template.ID, 0)
	if err == nil {
		t.Error("MoveTask() with position 0 should return error")
	}
}

func TestTemplateService_RandomizeTaskOrder(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Test", "", 12345, 0, false)
	for i := 0; i < 10; i++ {
		taskSvc.Create(challenge.ID, "Task", "", "")
	}

	template, _ := templateSvc.CreateFromChallenge(challenge.ID)

	// Randomize multiple times to ensure it works
	for i := 0; i < 5; i++ {
		err := templateSvc.RandomizeTaskOrder(template.ID)
		if err != nil {
			t.Fatalf("RandomizeTaskOrder() error = %v", err)
		}
	}

	// Verify all tasks still exist with valid order numbers
	tasks, _ := templateSvc.GetTasks(template.ID)
	if len(tasks) != 10 {
		t.Errorf("After randomize, task count = %d, want 10", len(tasks))
	}

	// Verify order numbers are 1-10 (no duplicates or gaps)
	orderNums := make(map[int]bool)
	for _, task := range tasks {
		if task.OrderNum < 1 || task.OrderNum > 10 {
			t.Errorf("Invalid OrderNum = %d", task.OrderNum)
		}
		if orderNums[task.OrderNum] {
			t.Errorf("Duplicate OrderNum = %d", task.OrderNum)
		}
		orderNums[task.OrderNum] = true
	}
}

func TestTemplateService_RandomizeTaskOrder_SingleTask(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Test", "", 12345, 0, false)
	taskSvc.Create(challenge.ID, "Only Task", "", "")

	template, _ := templateSvc.CreateFromChallenge(challenge.ID)

	// Should not error with single task
	err := templateSvc.RandomizeTaskOrder(template.ID)
	if err != nil {
		t.Fatalf("RandomizeTaskOrder() with single task error = %v", err)
	}
}

func TestTemplateService_Count(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	templateSvc := NewTemplateService(repo)

	count, err := templateSvc.Count()
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Count() = %d, want 0", count)
	}

	ch1, _ := challengeSvc.Create("Template 1", "", 12345, 0, false)
	ch2, _ := challengeSvc.Create("Template 2", "", 12345, 0, false)
	templateSvc.CreateFromChallenge(ch1.ID)
	templateSvc.CreateFromChallenge(ch2.ID)

	count, err = templateSvc.Count()
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 2 {
		t.Errorf("Count() = %d, want 2", count)
	}
}

func TestTemplateService_GetTaskCount(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Test", "", 12345, 0, false)
	taskSvc.Create(challenge.ID, "Task 1", "", "")
	taskSvc.Create(challenge.ID, "Task 2", "", "")
	taskSvc.Create(challenge.ID, "Task 3", "", "")

	template, _ := templateSvc.CreateFromChallenge(challenge.ID)

	count, err := templateSvc.GetTaskCount(template.ID)
	if err != nil {
		t.Fatalf("GetTaskCount() error = %v", err)
	}
	if count != 3 {
		t.Errorf("GetTaskCount() = %d, want 3", count)
	}
}

func TestTemplateService_UpdateTaskTitle(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Test", "", 12345, 0, false)
	taskSvc.Create(challenge.ID, "Original", "", "")

	template, _ := templateSvc.CreateFromChallenge(challenge.ID)
	tasks, _ := templateSvc.GetTasks(template.ID)

	err := templateSvc.UpdateTaskTitle(tasks[0].ID, "Updated Title")
	if err != nil {
		t.Fatalf("UpdateTaskTitle() error = %v", err)
	}

	updated, _ := templateSvc.GetTaskByID(tasks[0].ID)
	if updated.Title != "Updated Title" {
		t.Errorf("UpdateTaskTitle() Title = %q, want %q", updated.Title, "Updated Title")
	}
}

func TestTemplateService_UpdateTaskDescription(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Test", "", 12345, 0, false)
	taskSvc.Create(challenge.ID, "Task", "", "")

	template, _ := templateSvc.CreateFromChallenge(challenge.ID)
	tasks, _ := templateSvc.GetTasks(template.ID)

	err := templateSvc.UpdateTaskDescription(tasks[0].ID, "New Description")
	if err != nil {
		t.Fatalf("UpdateTaskDescription() error = %v", err)
	}

	updated, _ := templateSvc.GetTaskByID(tasks[0].ID)
	if updated.Description != "New Description" {
		t.Errorf("UpdateTaskDescription() Description = %q, want %q", updated.Description, "New Description")
	}
}

func TestTemplateService_UpdateTaskImage(t *testing.T) {
	repo := setupTestRepo(t)
	challengeSvc := NewChallengeService(repo)
	taskSvc := NewTaskService(repo)
	templateSvc := NewTemplateService(repo)

	challenge, _ := challengeSvc.Create("Test", "", 12345, 0, false)
	taskSvc.Create(challenge.ID, "Task", "", "")

	template, _ := templateSvc.CreateFromChallenge(challenge.ID)
	tasks, _ := templateSvc.GetTasks(template.ID)

	err := templateSvc.UpdateTaskImage(tasks[0].ID, "new_image_id")
	if err != nil {
		t.Fatalf("UpdateTaskImage() error = %v", err)
	}

	updated, _ := templateSvc.GetTaskByID(tasks[0].ID)
	if updated.ImageFileID != "new_image_id" {
		t.Errorf("UpdateTaskImage() ImageFileID = %q, want %q", updated.ImageFileID, "new_image_id")
	}
}
