package event

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/dpologdvinity/promptpet/internal/store"
)

type Event struct {
	Timestamp   int64  `json:"timestamp"`
	Type        string `json:"type"` // "command_finish", "git_commit", "git_branch"
	ExitCode    int    `json:"exit_code"`
	DurationMs  int64  `json:"duration_ms"`
	GitAction   string `json:"git_action,omitempty"`
	GitBranch   string `json:"git_branch,omitempty"`
	CommitCount int    `json:"commit_count,omitempty"`
}

func GetEventFilePath() string {
	if f := os.Getenv("PROMPTPET_EVENT_FILE"); f != "" {
		return f
	}
	dir, err := store.Dir(os.Getenv)
	if err != nil {
		return filepath.Join(os.TempDir(), "promptpet", "events.jsonl")
	}
	return filepath.Join(dir, "events.jsonl")
}

func AppendEventToFile(filePath string, ev Event) error {
	if ev.Timestamp == 0 {
		ev.Timestamp = time.Now().Unix()
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}

func LogEvent(ev Event) error {
	return AppendEventToFile(GetEventFilePath(), ev)
}

func ReadEventsSince(filePath string, sinceTimestamp int64) ([]Event, error) {
	f, err := os.Open(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return []Event{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var ev Event
		if err := json.Unmarshal(line, &ev); err != nil {
			// Skip corrupted lines
			continue
		}
		if ev.Timestamp >= sinceTimestamp {
			events = append(events, ev)
		}
	}
	if events == nil {
		events = []Event{}
	}
	return events, nil
}
