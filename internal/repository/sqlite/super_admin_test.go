package sqlite

import (
	"testing"
)

func TestSuperAdminRepo_Create(t *testing.T) {
	repo := setupTestDB(t)

	err := repo.SuperAdmin().Create(123456789)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify it exists
	exists, err := repo.SuperAdmin().Exists(123456789)
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Super admin should exist after creation")
	}
}

func TestSuperAdminRepo_Create_Duplicate(t *testing.T) {
	repo := setupTestDB(t)

	// Create first time
	err := repo.SuperAdmin().Create(123456789)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Create again - should be ignored (INSERT OR IGNORE)
	err = repo.SuperAdmin().Create(123456789)
	if err != nil {
		t.Fatalf("Create() duplicate should not error, got = %v", err)
	}

	// Should still only have one entry
	admins, _ := repo.SuperAdmin().GetAll()
	if len(admins) != 1 {
		t.Errorf("GetAll() count = %d, want 1", len(admins))
	}
}

func TestSuperAdminRepo_IsSuperAdmin(t *testing.T) {
	repo := setupTestDB(t)

	// Should be false for non-existent user
	isSuperAdmin, err := repo.SuperAdmin().IsSuperAdmin(123456789)
	if err != nil {
		t.Fatalf("IsSuperAdmin() error = %v", err)
	}
	if isSuperAdmin {
		t.Error("User should not be super admin before creation")
	}

	// Create and verify
	repo.SuperAdmin().Create(123456789)
	isSuperAdmin, err = repo.SuperAdmin().IsSuperAdmin(123456789)
	if err != nil {
		t.Fatalf("IsSuperAdmin() error = %v", err)
	}
	if !isSuperAdmin {
		t.Error("User should be super admin after creation")
	}
}

func TestSuperAdminRepo_Delete(t *testing.T) {
	repo := setupTestDB(t)

	repo.SuperAdmin().Create(123456789)

	err := repo.SuperAdmin().Delete(123456789)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	exists, _ := repo.SuperAdmin().Exists(123456789)
	if exists {
		t.Error("Super admin should not exist after deletion")
	}
}

func TestSuperAdminRepo_Delete_NonExistent(t *testing.T) {
	repo := setupTestDB(t)

	// Deleting non-existent should not error
	err := repo.SuperAdmin().Delete(999999)
	if err != nil {
		t.Fatalf("Delete() non-existent should not error, got = %v", err)
	}
}

func TestSuperAdminRepo_GetAll(t *testing.T) {
	repo := setupTestDB(t)

	repo.SuperAdmin().Create(111)
	repo.SuperAdmin().Create(222)
	repo.SuperAdmin().Create(333)

	admins, err := repo.SuperAdmin().GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(admins) != 3 {
		t.Errorf("GetAll() count = %d, want 3", len(admins))
	}

	// Verify IDs are present
	ids := make(map[int64]bool)
	for _, admin := range admins {
		ids[admin.TelegramID] = true
	}
	for _, expected := range []int64{111, 222, 333} {
		if !ids[expected] {
			t.Errorf("GetAll() missing ID %d", expected)
		}
	}
}

func TestSuperAdminRepo_GetAll_Empty(t *testing.T) {
	repo := setupTestDB(t)

	admins, err := repo.SuperAdmin().GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(admins) != 0 {
		t.Errorf("GetAll() count = %d, want 0", len(admins))
	}
}

func TestSuperAdminRepo_Exists(t *testing.T) {
	repo := setupTestDB(t)

	// Non-existent
	exists, err := repo.SuperAdmin().Exists(123456789)
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Exists() should return false for non-existent")
	}

	// After creation
	repo.SuperAdmin().Create(123456789)
	exists, err = repo.SuperAdmin().Exists(123456789)
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() should return true after creation")
	}
}
