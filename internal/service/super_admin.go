package service

import (
	"errors"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/repository"
)

var (
	ErrNotSuperAdmin      = errors.New("not a super admin")
	ErrSuperAdminNotFound = errors.New("super admin not found")
	ErrAlreadySuperAdmin  = errors.New("user is already a super admin")
	ErrCannotRemoveSelf   = errors.New("cannot remove yourself as super admin")
)

// SuperAdminService handles super admin business logic
type SuperAdminService struct {
	repo repository.Repository
}

// NewSuperAdminService creates a new SuperAdminService
func NewSuperAdminService(repo repository.Repository) *SuperAdminService {
	return &SuperAdminService{repo: repo}
}

// IsSuperAdmin checks if a user is a super admin
func (s *SuperAdminService) IsSuperAdmin(telegramID int64) (bool, error) {
	return s.repo.SuperAdmin().IsSuperAdmin(telegramID)
}

// Grant grants super admin privileges to a user
func (s *SuperAdminService) Grant(grantedByID, targetID int64) error {
	// Verify granter is super admin
	isSuperAdmin, err := s.IsSuperAdmin(grantedByID)
	if err != nil {
		return err
	}
	if !isSuperAdmin {
		return ErrNotSuperAdmin
	}

	// Check if target already is super admin
	exists, err := s.repo.SuperAdmin().Exists(targetID)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadySuperAdmin
	}

	return s.repo.SuperAdmin().Create(targetID)
}

// Revoke removes super admin privileges from a user
func (s *SuperAdminService) Revoke(revokedByID, targetID int64) error {
	// Verify revoker is super admin
	isSuperAdmin, err := s.IsSuperAdmin(revokedByID)
	if err != nil {
		return err
	}
	if !isSuperAdmin {
		return ErrNotSuperAdmin
	}

	// Prevent self-removal
	if revokedByID == targetID {
		return ErrCannotRemoveSelf
	}

	// Check if target exists
	exists, err := s.repo.SuperAdmin().Exists(targetID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrSuperAdminNotFound
	}

	return s.repo.SuperAdmin().Delete(targetID)
}

// GetAll returns all super admins
func (s *SuperAdminService) GetAll() ([]*domain.SuperAdmin, error) {
	return s.repo.SuperAdmin().GetAll()
}

// SeedFromEnv seeds the initial super admin from environment variable
// This should only add if no super admins exist OR if the ID doesn't exist yet
func (s *SuperAdminService) SeedFromEnv(telegramID int64) error {
	if telegramID == 0 {
		return nil // No seed ID provided
	}
	return s.repo.SuperAdmin().Create(telegramID) // INSERT OR IGNORE
}

// GetAllChallenges returns all challenges (for super admin observer mode)
func (s *SuperAdminService) GetAllChallenges() ([]*domain.Challenge, error) {
	return s.repo.Challenge().GetAll()
}
