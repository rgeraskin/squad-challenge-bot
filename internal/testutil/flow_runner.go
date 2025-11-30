package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rgeraskin/squad-challenge-bot/internal/repository/sqlite"
	"github.com/rgeraskin/squad-challenge-bot/internal/service"
)

// FlowRunner helps test complete user flows
type FlowRunner struct {
	T            *testing.T
	Repo         *sqlite.SQLiteRepository
	Challenge    *service.ChallengeService
	Task         *service.TaskService
	Participant  *service.ParticipantService
	Completion   *service.CompletionService
	State        *service.StateService
}

// NewFlowRunner creates a new flow runner with all services
func NewFlowRunner(t *testing.T) *FlowRunner {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "squadbot-flow-test-*")
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

	return &FlowRunner{
		T:           t,
		Repo:        repo,
		Challenge:   service.NewChallengeService(repo),
		Task:        service.NewTaskService(repo),
		Participant: service.NewParticipantService(repo),
		Completion:  service.NewCompletionService(repo),
		State:       service.NewStateService(repo),
	}
}

// CreateChallengeWithTasks creates a challenge with the specified number of tasks
func (f *FlowRunner) CreateChallengeWithTasks(creatorID int64, challengeName string, numTasks int) string {
	// Create challenge
	challenge, err := f.Challenge.Create(challengeName, "", creatorID)
	if err != nil {
		f.T.Fatalf("Failed to create challenge: %v", err)
	}

	// Add creator as participant
	_, err = f.Participant.Join(challenge.ID, creatorID, "Creator", "💪")
	if err != nil {
		f.T.Fatalf("Failed to add creator as participant: %v", err)
	}

	// Create tasks
	for i := 0; i < numTasks; i++ {
		_, err := f.Task.Create(challenge.ID, "Task "+string(rune('A'+i)), "", "")
		if err != nil {
			f.T.Fatalf("Failed to create task: %v", err)
		}
	}

	return challenge.ID
}

// JoinChallenge adds a user to a challenge
func (f *FlowRunner) JoinChallenge(challengeID string, userID int64, name, emoji string) int64 {
	participant, err := f.Participant.Join(challengeID, userID, name, emoji)
	if err != nil {
		f.T.Fatalf("Failed to join challenge: %v", err)
	}
	return participant.ID
}

// CompleteTasks completes the specified task numbers for a participant
func (f *FlowRunner) CompleteTasks(challengeID string, participantID int64, taskNums ...int) {
	tasks, err := f.Task.GetByChallengeID(challengeID)
	if err != nil {
		f.T.Fatalf("Failed to get tasks: %v", err)
	}

	for _, num := range taskNums {
		for _, task := range tasks {
			if task.OrderNum == num {
				_, err := f.Completion.Complete(task.ID, participantID)
				if err != nil {
					f.T.Fatalf("Failed to complete task %d: %v", num, err)
				}
				break
			}
		}
	}
}

// GetProgress returns completed/total tasks for a participant
func (f *FlowRunner) GetProgress(challengeID string, participantID int64) (completed, total int) {
	tasks, _ := f.Task.GetByChallengeID(challengeID)
	total = len(tasks)
	completed, _ = f.Completion.CountByParticipantID(participantID)
	return
}

// AssertProgress verifies the participant's progress
func (f *FlowRunner) AssertProgress(challengeID string, participantID int64, expectedCompleted, expectedTotal int) {
	completed, total := f.GetProgress(challengeID, participantID)
	if completed != expectedCompleted || total != expectedTotal {
		f.T.Errorf("Progress = %d/%d, want %d/%d", completed, total, expectedCompleted, expectedTotal)
	}
}
