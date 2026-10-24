package companion

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dpologdvinity/promptpet/internal/store"
)

type State struct {
	ActivePet           string         `json:"active_pet"`
	FriendshipHearts    int            `json:"friendship_hearts"`
	HighScores          map[string]int `json:"high_scores"`
	LastWaterTime       int64          `json:"last_water_time"`
	LastStretchTime     int64          `json:"last_stretch_time"`
	LastLookAwayTime    int64          `json:"last_lookaway_time"`
	TotalCommandsWorked int            `json:"total_commands_worked"`
}

const DefaultPet = "usagi"

func NewDefaultState() *State {
	return &State{
		ActivePet:        DefaultPet,
		FriendshipHearts: 0,
		HighScores:       make(map[string]int),
	}
}

func DefaultStatePath() string {
	if f := os.Getenv("PROMPTPET_STATE_FILE"); f != "" {
		return f
	}
	dir, err := store.Dir(os.Getenv)
	if err != nil {
		return filepath.Join(os.TempDir(), "promptpet", "state.json")
	}
	return filepath.Join(dir, "state.json")
}

func LoadState(path string) (*State, error) {
	if path == "" {
		path = DefaultStatePath()
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return NewDefaultState(), nil
	}
	if err != nil {
		return nil, err
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		// Ignore corrupted state and recover with default
		return NewDefaultState(), nil
	}

	if s.ActivePet == "" {
		s.ActivePet = DefaultPet
	}
	if s.HighScores == nil {
		s.HighScores = make(map[string]int)
	}

	return &s, nil
}

func SaveState(path string, s *State) error {
	if path == "" {
		path = DefaultStatePath()
	}
	if s == nil {
		s = NewDefaultState()
	}
	if s.ActivePet == "" {
		s.ActivePet = DefaultPet
	}
	if s.HighScores == nil {
		s.HighScores = make(map[string]int)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	tmp := fmt.Sprintf("%s.tmp-%d", path, os.Getpid())
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
