package sqlite

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDB(t *testing.T) *SQLiteRepository {
	t.Helper()

	// Create temp directory for test DB
	tmpDir, err := os.MkdirTemp("", "squadbot-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	repo, err := New(dbPath)
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

func TestNew(t *testing.T) {
	repo := setupTestDB(t)
	if repo == nil {
		t.Fatal("New() returned nil")
	}
}
