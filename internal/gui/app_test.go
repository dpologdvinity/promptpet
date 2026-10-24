package gui_test

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/dpologdvinity/promptpet/internal/companion"
	"github.com/dpologdvinity/promptpet/internal/event"
	"github.com/dpologdvinity/promptpet/internal/gui"
)

func TestAppInitialization(t *testing.T) {
	tmp := t.TempDir()
	cfg := gui.Config{
		StatePath: filepath.Join(tmp, "state.json"),
		EventPath: filepath.Join(tmp, "events.jsonl"),
	}
	app, err := gui.NewApp(cfg)
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}
	w, h := app.Layout(0, 0)
	if w != 240 || h != 340 {
		t.Errorf("expected 240x340 layout, got %dx%d", w, h)
	}
	if app.ActivePet().ID != companion.PetUsagi {
		t.Errorf("expected initial pet to be %s, got %s", companion.PetUsagi, app.ActivePet().ID)
	}
}

func TestGracefulHeadlessCheck(t *testing.T) {
	tmp := t.TempDir()
	cfg := gui.Config{
		StatePath: filepath.Join(tmp, "state.json"),
		EventPath: filepath.Join(tmp, "events.jsonl"),
	}
	t.Setenv("PROMPTPET_HEADLESS", "1")
	err := gui.Run(cfg)
	if err == nil {
		t.Fatal("expected headless run to return an error, got nil")
	}
	if !errors.Is(err, gui.ErrHeadless) {
		t.Errorf("expected ErrHeadless, got %v", err)
	}
}

func TestPetCyclingAndStateSaving(t *testing.T) {
	tmp := t.TempDir()
	stateFile := filepath.Join(tmp, "state.json")
	cfg := gui.Config{
		StatePath: stateFile,
		EventPath: filepath.Join(tmp, "events.jsonl"),
	}
	app, err := gui.NewApp(cfg)
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	initialPet := app.ActivePet().ID
	if initialPet != companion.PetUsagi {
		t.Fatalf("expected initial pet usagi, got %s", initialPet)
	}

	// Cycle to next pet
	app.CyclePet()
	nextPet := app.ActivePet().ID
	if nextPet == initialPet {
		t.Fatalf("expected pet to change after cycling, still %s", nextPet)
	}

	// Verify state file was updated
	savedState, err := companion.LoadState(stateFile)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	if savedState.ActivePet != string(nextPet) {
		t.Errorf("expected saved state to have %s, got %s", nextPet, savedState.ActivePet)
	}

	// Cycle through remaining pets and ensure full rotation of all 8 pets
	visited := map[companion.PetID]bool{
		initialPet: true,
		nextPet:    true,
	}
	for i := 0; i < 6; i++ {
		app.CyclePet()
		visited[app.ActivePet().ID] = true
	}
	if len(visited) != 8 {
		t.Errorf("expected all 8 pets visited in rotation, visited %d", len(visited))
	}

	// 8th cycle wraps back to initial pet
	app.CyclePet()
	if app.ActivePet().ID != initialPet {
		t.Errorf("expected wrap around to %s, got %s", initialPet, app.ActivePet().ID)
	}
}

func TestAppDrawAllPets(t *testing.T) {
	tmp := t.TempDir()
	cfg := gui.Config{
		StatePath: filepath.Join(tmp, "state.json"),
		EventPath: filepath.Join(tmp, "events.jsonl"),
	}
	app, err := gui.NewApp(cfg)
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	screen := ebiten.NewImage(240, 340)

	// Draw for each pet without panics
	for i := 0; i < 8; i++ {
		app.Draw(screen)
		app.CyclePet()
	}
}

func TestAppUpdateRhythmAndWellness(t *testing.T) {
	tmp := t.TempDir()
	eventFile := filepath.Join(tmp, "events.jsonl")
	stateFile := filepath.Join(tmp, "state.json")
	cfg := gui.Config{
		StatePath: stateFile,
		EventPath: eventFile,
	}
	app, err := gui.NewApp(cfg)
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	// Log a success event
	err = event.AppendEventToFile(eventFile, event.Event{
		Timestamp: time.Now().Unix(),
		Type:      "command_finish",
		ExitCode:  0,
	})
	if err != nil {
		t.Fatalf("failed to append event: %v", err)
	}

	// Update app
	if err := app.Update(); err != nil {
		t.Fatalf("app update failed: %v", err)
	}

	if app.Speech() == "" {
		t.Errorf("expected non-empty speech after update")
	}
}

func TestAppUpdateMultipleTicksNoEventReprocessingOrWellnessClobber(t *testing.T) {
	tmp := t.TempDir()
	eventFile := filepath.Join(tmp, "events.jsonl")
	stateFile := filepath.Join(tmp, "state.json")
	cfg := gui.Config{
		StatePath: stateFile,
		EventPath: eventFile,
	}

	// Setup state with overdue hydration reminder
	s := companion.NewDefaultState()
	s.LastWaterTime = time.Now().Add(-60 * time.Minute).Unix()
	s.LastStretchTime = time.Now().Unix()
	s.LastLookAwayTime = time.Now().Unix()
	if err := companion.SaveState(stateFile, s); err != nil {
		t.Fatalf("failed to save initial state: %v", err)
	}

	app, err := gui.NewApp(cfg)
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	// Append a single initial event
	evTs := time.Now().Unix()
	err = event.AppendEventToFile(eventFile, event.Event{
		Timestamp: evTs,
		Type:      "command_finish",
		ExitCode:  0,
	})
	if err != nil {
		t.Fatalf("failed to append event: %v", err)
	}

	// Tick 1: processes the event and activates the overdue wellness reminder
	if err := app.Update(); err != nil {
		t.Fatalf("tick 1 update failed: %v", err)
	}

	expectedWellnessSpeech := app.ActivePet().SpeechFor(string(companion.WellnessWater))
	if app.Speech() != expectedWellnessSpeech {
		t.Fatalf("expected wellness speech %q on tick 1, got %q", expectedWellnessSpeech, app.Speech())
	}

	// Ticks 2 through 10: ensure active wellness speech is NOT clobbered
	// and seen events are NOT re-evaluated on subsequent ticks
	for tick := 2; tick <= 10; tick++ {
		if err := app.Update(); err != nil {
			t.Fatalf("tick %d update failed: %v", tick, err)
		}
		if app.Speech() != expectedWellnessSpeech {
			t.Fatalf("wellness speech clobbered on tick %d: expected %q, got %q", tick, expectedWellnessSpeech, app.Speech())
		}
	}
}
