package companion_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dpologdvinity/promptpet/internal/companion"
)

func TestStatePersistence(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "state.json")

	// 1. Loading non-existent file returns default state
	s, err := companion.LoadState(path)
	if err != nil {
		t.Fatalf("unexpected error loading non-existent state: %v", err)
	}
	if s.ActivePet != "usagi" {
		t.Errorf("expected default active pet 'usagi', got %q", s.ActivePet)
	}
	if s.HighScores == nil {
		t.Errorf("expected non-nil high scores map")
	}

	// 2. Modifying and saving state
	s.ActivePet = "kochi"
	s.FriendshipHearts = 5
	s.HighScores["snack_catch"] = 100
	s.LastWaterTime = time.Now().Unix()
	s.TotalCommandsWorked = 42

	if err := companion.SaveState(path, s); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// 3. Loading saved state
	loaded, err := companion.LoadState(path)
	if err != nil {
		t.Fatalf("failed to reload state: %v", err)
	}
	if loaded.ActivePet != "kochi" {
		t.Errorf("expected active pet 'kochi', got %q", loaded.ActivePet)
	}
	if loaded.FriendshipHearts != 5 {
		t.Errorf("expected friendship hearts 5, got %d", loaded.FriendshipHearts)
	}
	if loaded.HighScores["snack_catch"] != 100 {
		t.Errorf("expected high score 100, got %d", loaded.HighScores["snack_catch"])
	}
	if loaded.TotalCommandsWorked != 42 {
		t.Errorf("expected total commands worked 42, got %d", loaded.TotalCommandsWorked)
	}

	// 4. Broken JSON recovers to default state without crash
	corruptPath := filepath.Join(tmp, "corrupt.json")
	if err := os.WriteFile(corruptPath, []byte("{invalid json"), 0o600); err != nil {
		t.Fatalf("failed to write corrupt file: %v", err)
	}
	sCorrupt, err := companion.LoadState(corruptPath)
	if err != nil {
		t.Fatalf("load corrupt state should recover gracefully, got error: %v", err)
	}
	if sCorrupt.ActivePet != "usagi" {
		t.Errorf("expected default pet 'usagi' after corrupt recovery, got %q", sCorrupt.ActivePet)
	}

	// 5. Empty active_pet in JSON falls back to "usagi"
	emptyPetPath := filepath.Join(tmp, "empty_pet.json")
	if err := os.WriteFile(emptyPetPath, []byte(`{"active_pet":""}`), 0o600); err != nil {
		t.Fatalf("failed to write empty pet file: %v", err)
	}
	sEmpty, err := companion.LoadState(emptyPetPath)
	if err != nil {
		t.Fatalf("failed to load state with empty pet: %v", err)
	}
	if sEmpty.ActivePet != "usagi" {
		t.Errorf("expected fallback to 'usagi', got %q", sEmpty.ActivePet)
	}
}
