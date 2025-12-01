package service

import (
	"testing"
)

func TestSuperAdminService_IsSuperAdmin(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewSuperAdminService(repo)

	// Not a super admin initially
	isSuperAdmin, err := svc.IsSuperAdmin(123456789)
	if err != nil {
		t.Fatalf("IsSuperAdmin() error = %v", err)
	}
	if isSuperAdmin {
		t.Error("User should not be super admin initially")
	}

	// Seed and check again
	svc.SeedFromEnv(123456789)
	isSuperAdmin, err = svc.IsSuperAdmin(123456789)
	if err != nil {
		t.Fatalf("IsSuperAdmin() error = %v", err)
	}
	if !isSuperAdmin {
		t.Error("User should be super admin after seeding")
	}
}

func TestSuperAdminService_SeedFromEnv(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewSuperAdminService(repo)

	// Seed with valid ID
	err := svc.SeedFromEnv(123456789)
	if err != nil {
		t.Fatalf("SeedFromEnv() error = %v", err)
	}

	isSuperAdmin, _ := svc.IsSuperAdmin(123456789)
	if !isSuperAdmin {
		t.Error("User should be super admin after seeding")
	}
}

func TestSuperAdminService_SeedFromEnv_ZeroID(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewSuperAdminService(repo)

	// Seed with 0 should be no-op
	err := svc.SeedFromEnv(0)
	if err != nil {
		t.Fatalf("SeedFromEnv(0) error = %v", err)
	}

	admins, _ := svc.GetAll()
	if len(admins) != 0 {
		t.Errorf("GetAll() count = %d, want 0 after seeding with 0", len(admins))
	}
}

func TestSuperAdminService_Grant(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewSuperAdminService(repo)

	// First, seed an initial super admin
	svc.SeedFromEnv(111)

	// Grant to another user
	err := svc.Grant(111, 222)
	if err != nil {
		t.Fatalf("Grant() error = %v", err)
	}

	// Verify new user is super admin
	isSuperAdmin, _ := svc.IsSuperAdmin(222)
	if !isSuperAdmin {
		t.Error("User 222 should be super admin after grant")
	}
}

func TestSuperAdminService_Grant_NotSuperAdmin(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewSuperAdminService(repo)

	// Try to grant without being super admin
	err := svc.Grant(111, 222)
	if err != ErrNotSuperAdmin {
		t.Errorf("Grant() by non-super-admin: error = %v, want ErrNotSuperAdmin", err)
	}
}

func TestSuperAdminService_Grant_AlreadySuperAdmin(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewSuperAdminService(repo)

	svc.SeedFromEnv(111)
	svc.SeedFromEnv(222)

	// Try to grant to existing super admin
	err := svc.Grant(111, 222)
	if err != ErrAlreadySuperAdmin {
		t.Errorf("Grant() to existing super admin: error = %v, want ErrAlreadySuperAdmin", err)
	}
}

func TestSuperAdminService_Revoke(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewSuperAdminService(repo)

	svc.SeedFromEnv(111)
	svc.Grant(111, 222)

	// Revoke from 222
	err := svc.Revoke(111, 222)
	if err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}

	isSuperAdmin, _ := svc.IsSuperAdmin(222)
	if isSuperAdmin {
		t.Error("User 222 should not be super admin after revoke")
	}
}

func TestSuperAdminService_Revoke_NotSuperAdmin(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewSuperAdminService(repo)

	// Try to revoke without being super admin
	err := svc.Revoke(111, 222)
	if err != ErrNotSuperAdmin {
		t.Errorf("Revoke() by non-super-admin: error = %v, want ErrNotSuperAdmin", err)
	}
}

func TestSuperAdminService_Revoke_CannotRemoveSelf(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewSuperAdminService(repo)

	svc.SeedFromEnv(111)

	err := svc.Revoke(111, 111)
	if err != ErrCannotRemoveSelf {
		t.Errorf("Revoke() self: error = %v, want ErrCannotRemoveSelf", err)
	}

	// Verify still super admin
	isSuperAdmin, _ := svc.IsSuperAdmin(111)
	if !isSuperAdmin {
		t.Error("User 111 should still be super admin after failed self-revoke")
	}
}

func TestSuperAdminService_Revoke_NotFound(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewSuperAdminService(repo)

	svc.SeedFromEnv(111)

	err := svc.Revoke(111, 999)
	if err != ErrSuperAdminNotFound {
		t.Errorf("Revoke() non-existent: error = %v, want ErrSuperAdminNotFound", err)
	}
}

func TestSuperAdminService_GetAll(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewSuperAdminService(repo)

	svc.SeedFromEnv(111)
	svc.Grant(111, 222)
	svc.Grant(111, 333)

	admins, err := svc.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(admins) != 3 {
		t.Errorf("GetAll() count = %d, want 3", len(admins))
	}
}

func TestSuperAdminService_GetAllChallenges(t *testing.T) {
	repo := setupTestRepo(t)
	superAdminSvc := NewSuperAdminService(repo)
	challengeSvc := NewChallengeService(repo)

	// Create challenges by different users
	challengeSvc.Create("Challenge 1", "", 111, 0, false)
	challengeSvc.Create("Challenge 2", "", 222, 0, false)
	challengeSvc.Create("Challenge 3", "", 333, 0, false)

	// Super admin should see all challenges
	challenges, err := superAdminSvc.GetAllChallenges()
	if err != nil {
		t.Fatalf("GetAllChallenges() error = %v", err)
	}
	if len(challenges) != 3 {
		t.Errorf("GetAllChallenges() count = %d, want 3", len(challenges))
	}
}
