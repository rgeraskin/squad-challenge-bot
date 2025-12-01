package sqlite

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
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
	superAdmin  *SuperAdminRepo
}

// New creates a new SQLite repository
func New(dbPath string) (*SQLiteRepository, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sqlx.Connect("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable foreign keys (must be set per connection)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}

	repo := &SQLiteRepository{
		db:          db,
		challenge:   &ChallengeRepo{db: db},
		task:        &TaskRepo{db: db},
		participant: &ParticipantRepo{db: db},
		completion:  &CompletionRepo{db: db},
		state:       &StateRepo{db: db},
		superAdmin:  &SuperAdminRepo{db: db},
	}

	if err := repo.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	return repo, nil
}

func (r *SQLiteRepository) migrate() error {
	migrations := []string{
		"migrations/001_initial.sql",
		"migrations/002_super_admins.sql",
	}

	for _, m := range migrations {
		migration, err := migrationsFS.ReadFile(m)
		if err != nil {
			return err
		}
		if _, err = r.db.Exec(string(migration)); err != nil {
			return err
		}
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

func (r *SQLiteRepository) SuperAdmin() repository.SuperAdminRepository {
	return r.superAdmin
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
