package sqlite

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rgeraskin/squad-challenge-bot/internal/repository"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// SQLiteRepository implements repository.Repository for SQLite
type SQLiteRepository struct {
	db          *sqlx.DB
	challenge   *ChallengeRepo
	task        *TaskRepo
	participant *ParticipantRepo
	completion  *CompletionRepo
	state       *StateRepo
}

// New creates a new SQLite repository
func New(dbPath string) (*SQLiteRepository, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sqlx.Connect("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	repo := &SQLiteRepository{
		db:          db,
		challenge:   &ChallengeRepo{db: db},
		task:        &TaskRepo{db: db},
		participant: &ParticipantRepo{db: db},
		completion:  &CompletionRepo{db: db},
		state:       &StateRepo{db: db},
	}

	if err := repo.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	return repo, nil
}

func (r *SQLiteRepository) migrate() error {
	// Run migrations in order
	migrations := []string{
		"migrations/001_initial.sql",
		"migrations/002_add_challenge_description.sql",
		"migrations/003_daily_limit.sql",
		"migrations/004_hide_future_tasks.sql",
	}

	for _, migrationFile := range migrations {
		migration, err := migrationsFS.ReadFile(migrationFile)
		if err != nil {
			return err
		}
		// Ignore errors for idempotent migrations (e.g., column already exists)
		r.db.Exec(string(migration))
	}

	return nil
}

func (r *SQLiteRepository) Challenge() repository.ChallengeRepository {
	return r.challenge
}

func (r *SQLiteRepository) Task() repository.TaskRepository {
	return r.task
}

func (r *SQLiteRepository) Participant() repository.ParticipantRepository {
	return r.participant
}

func (r *SQLiteRepository) Completion() repository.CompletionRepository {
	return r.completion
}

func (r *SQLiteRepository) State() repository.StateRepository {
	return r.state
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
