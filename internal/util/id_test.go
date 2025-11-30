package util

import (
	"testing"
)

func TestGenerateID(t *testing.T) {
	id, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID() error = %v", err)
	}

	if len(id) != 8 {
		t.Errorf("GenerateID() length = %d, want 8", len(id))
	}

	// Check all characters are valid
	for _, c := range id {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			t.Errorf("GenerateID() contains invalid character: %c", c)
		}
	}
}

func TestGenerateID_Uniqueness(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := GenerateID()
		if err != nil {
			t.Fatalf("GenerateID() error = %v", err)
		}
		if ids[id] {
			t.Errorf("GenerateID() produced duplicate ID: %s", id)
		}
		ids[id] = true
	}
}
