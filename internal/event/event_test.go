package event_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dpologdvinity/promptpet/internal/event"
)

func TestEventSerialization(t *testing.T) {
	tmpDir := t.TempDir()
	evFile := filepath.Join(tmpDir, "events.jsonl")

	ev1 := event.Event{
		Timestamp: time.Now().Unix(),
		Type:      "command_finish",
		ExitCode:  0,
	}
	if err := event.AppendEventToFile(evFile, ev1); err != nil {
		t.Fatalf("failed to append event: %v", err)
	}

	events, err := event.ReadEventsSince(evFile, 0)
	if err != nil {
		t.Fatalf("failed to read events: %v", err)
	}
	if len(events) != 1 || events[0].ExitCode != 0 {
		t.Fatalf("expected 1 event with exit code 0, got %+v", events)
	}
}

func TestReadEventsSince_Filtering(t *testing.T) {
	tmpDir := t.TempDir()
	evFile := filepath.Join(tmpDir, "events.jsonl")

	ev1 := event.Event{Timestamp: 100, Type: "command_finish", ExitCode: 0}
	ev2 := event.Event{Timestamp: 200, Type: "command_finish", ExitCode: 1}
	ev3 := event.Event{Timestamp: 300, Type: "git_commit", GitAction: "commit", CommitCount: 5}

	for _, ev := range []event.Event{ev1, ev2, ev3} {
		if err := event.AppendEventToFile(evFile, ev); err != nil {
			t.Fatalf("failed to append: %v", err)
		}
	}

	events, err := event.ReadEventsSince(evFile, 200)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events since 200, got %d", len(events))
	}
	if events[0].Timestamp != 200 || events[0].ExitCode != 1 {
		t.Errorf("unexpected event[0]: %+v", events[0])
	}
	if events[1].Timestamp != 300 || events[1].GitAction != "commit" {
		t.Errorf("unexpected event[1]: %+v", events[1])
	}
}

func TestReadEventsSince_CorruptLines(t *testing.T) {
	tmpDir := t.TempDir()
	evFile := filepath.Join(tmpDir, "events.jsonl")

	content := `{"timestamp":100,"type":"command_finish","exit_code":0}
not a valid json line
{"timestamp":200,"type":"command_finish","exit_code":2}
{"truncated_json":
{"timestamp":300,"type":"command_finish","exit_code":0}
`
	if err := os.WriteFile(evFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	events, err := event.ReadEventsSince(evFile, 0)
	if err != nil {
		t.Fatalf("failed to read events: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 valid events skipping corrupt lines, got %d: %+v", len(events), events)
	}
	if events[0].Timestamp != 100 || events[1].Timestamp != 200 || events[2].Timestamp != 300 {
		t.Errorf("unexpected events: %+v", events)
	}
}

func TestReadEventsSince_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	evFile := filepath.Join(tmpDir, "does-not-exist.jsonl")

	events, err := event.ReadEventsSince(evFile, 0)
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected empty events slice, got %+v", events)
	}
}

func TestLogEvent_And_GetEventFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	evFile := filepath.Join(tmpDir, "sub", "events.jsonl")
	t.Setenv("PROMPTPET_EVENT_FILE", evFile)

	if got := event.GetEventFilePath(); got != evFile {
		t.Fatalf("expected %q, got %q", evFile, got)
	}

	ev := event.Event{
		Timestamp: 123456,
		Type:      "command_finish",
		ExitCode:  0,
	}
	if err := event.LogEvent(ev); err != nil {
		t.Fatalf("LogEvent failed: %v", err)
	}

	events, err := event.ReadEventsSince(evFile, 0)
	if err != nil {
		t.Fatalf("ReadEventsSince failed: %v", err)
	}
	if len(events) != 1 || events[0].Timestamp != 123456 {
		t.Fatalf("unexpected events: %+v", events)
	}
}
