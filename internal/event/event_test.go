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

func TestReadEventsFromOffset_Incremental(t *testing.T) {
	tmpDir := t.TempDir()
	evFile := filepath.Join(tmpDir, "events.jsonl")

	// 1. Initial 2 events
	ev1 := event.Event{Timestamp: 100, Type: "command_finish", ExitCode: 0}
	ev2 := event.Event{Timestamp: 200, Type: "command_finish", ExitCode: 1}
	if err := event.AppendEventToFile(evFile, ev1); err != nil {
		t.Fatal(err)
	}
	if err := event.AppendEventToFile(evFile, ev2); err != nil {
		t.Fatal(err)
	}

	events, off1, err := event.ReadEventsFromOffset(evFile, 0)
	if err != nil {
		t.Fatalf("first read failed: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if off1 <= 0 {
		t.Fatalf("expected positive offset, got %d", off1)
	}

	// 2. Read from off1 without any new events -> 0 events, same offset
	eventsEmpty, off2, err := event.ReadEventsFromOffset(evFile, off1)
	if err != nil {
		t.Fatalf("empty read failed: %v", err)
	}
	if len(eventsEmpty) != 0 {
		t.Fatalf("expected 0 events, got %d", len(eventsEmpty))
	}
	if off2 != off1 {
		t.Fatalf("expected same offset %d, got %d", off1, off2)
	}

	// 3. Append 3rd event and read from off1 -> returns ONLY 3rd event
	ev3 := event.Event{Timestamp: 300, Type: "command_finish", ExitCode: 0, GitAction: "commit"}
	if err := event.AppendEventToFile(evFile, ev3); err != nil {
		t.Fatal(err)
	}

	eventsThird, off3, err := event.ReadEventsFromOffset(evFile, off1)
	if err != nil {
		t.Fatalf("third read failed: %v", err)
	}
	if len(eventsThird) != 1 {
		t.Fatalf("expected 1 new event, got %d", len(eventsThird))
	}
	if eventsThird[0].Timestamp != 300 || eventsThird[0].GitAction != "commit" {
		t.Fatalf("unexpected event: %+v", eventsThird[0])
	}
	if off3 <= off1 {
		t.Fatalf("expected offset to advance past %d, got %d", off1, off3)
	}

	// 4. File truncation: reset to 0
	if err := os.WriteFile(evFile, []byte(`{"timestamp":400,"type":"command_finish"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	eventsReset, off4, err := event.ReadEventsFromOffset(evFile, off3)
	if err != nil {
		t.Fatalf("truncated read failed: %v", err)
	}
	if len(eventsReset) != 1 || eventsReset[0].Timestamp != 400 {
		t.Fatalf("expected 1 event from beginning after truncate, got %+v", eventsReset)
	}
	if off4 >= off3 {
		t.Fatalf("expected offset to reset lower than old %d, got %d", off3, off4)
	}
}

func TestReadEventsFromOffset_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	evFile := filepath.Join(tmpDir, "missing.jsonl")

	events, off, err := event.ReadEventsFromOffset(evFile, 0)
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if len(events) != 0 || off != 0 {
		t.Fatalf("expected empty events and 0 offset, got %d events, offset %d", len(events), off)
	}
}

func TestReadEventsFromOffset_CorruptLines(t *testing.T) {
	tmpDir := t.TempDir()
	evFile := filepath.Join(tmpDir, "corrupt.jsonl")

	content := `{"timestamp":100,"type":"command_finish","exit_code":0}` + "\n" +
		`{this is broken json` + "\n" +
		`{"timestamp":200,"type":"command_finish","exit_code":0}` + "\n"
	if err := os.WriteFile(evFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	events, off, err := event.ReadEventsFromOffset(evFile, 0)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 valid events skipping corrupt, got %d", len(events))
	}
	if off != int64(len(content)) {
		t.Fatalf("expected offset %d, got %d", len(content), off)
	}
}
