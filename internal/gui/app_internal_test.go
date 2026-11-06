package gui

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/dpologdvinity/promptpet/internal/companion"
	"github.com/dpologdvinity/promptpet/internal/event"
)

// TestMoodRelaxesWhenIdle: a cheer from a shell event fades back to resting after
// moodIdleAfter with no new events, instead of animating confetti indefinitely.
func TestMoodRelaxesWhenIdle(t *testing.T) {
	tmp := t.TempDir()
	app, err := NewApp(Config{
		StatePath: filepath.Join(tmp, "companion.json"),
		EventPath: filepath.Join(tmp, "events.jsonl"),
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	app.clock = func() time.Time { return now }

	if err := event.AppendEventToFile(app.cfg.EventPath, event.Event{Timestamp: now.Unix(), Type: "command_finish"}); err != nil {
		t.Fatal(err)
	}
	if err := app.Update(); err != nil {
		t.Fatal(err)
	}
	if app.mood != companion.MoodCheer {
		t.Fatalf("expected cheer after a successful command, got %q", app.mood)
	}

	now = now.Add(moodIdleAfter - time.Second)
	_ = app.Update()
	if app.mood != companion.MoodCheer {
		t.Fatalf("mood must hold until moodIdleAfter, got %q", app.mood)
	}

	now = now.Add(2 * time.Second)
	_ = app.Update()
	if app.mood != companion.MoodResting {
		t.Fatalf("expected resting after %v idle, got %q", moodIdleAfter, app.mood)
	}
	if want := app.activePet.SpeechFor("resting"); app.speech != want {
		t.Errorf("expected resting speech %q, got %q", want, app.speech)
	}
}
