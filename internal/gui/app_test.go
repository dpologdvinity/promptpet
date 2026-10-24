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

func TestAppMiniGameLauncherAndEsc(t *testing.T) {
	tmp := t.TempDir()
	cfg := gui.Config{
		StatePath: filepath.Join(tmp, "state.json"),
		EventPath: filepath.Join(tmp, "events.jsonl"),
	}
	app, err := gui.NewApp(cfg)
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	if app.ActiveGame() != nil {
		t.Fatalf("expected activeGame to be nil initially")
	}

	// Launch game 1 (Catch)
	app.LaunchGameByID(1)
	if app.ActiveGame() == nil || app.ActiveGame().Name() != "Snack Catch" {
		t.Fatalf("expected Snack Catch game active, got %v", app.ActiveGame())
	}

	// Render frame while game active
	screen := ebiten.NewImage(240, 340)
	app.Draw(screen)

	// Press Esc / ExitGame
	app.ExitGame()
	if app.ActiveGame() != nil {
		t.Fatalf("expected activeGame to be nil after exit")
	}

	// Launch game 2 (Leap)
	app.LaunchGameByID(2)
	if app.ActiveGame() == nil || app.ActiveGame().Name() != "Puddle Leap" {
		t.Fatalf("expected Puddle Leap active, got %v", app.ActiveGame())
	}

	// Launch game 3 (Look Away)
	app.LaunchGameByID(3)
	if app.ActiveGame() == nil || app.ActiveGame().Name() != "Look Away" {
		t.Fatalf("expected Look Away active, got %v", app.ActiveGame())
	}

	// Launch game 4 (Rhythm)
	app.LaunchGameByID(4)
	if app.ActiveGame() == nil || app.ActiveGame().Name() != "Rhythm Dance" {
		t.Fatalf("expected Rhythm Dance active, got %v", app.ActiveGame())
	}
}

func TestAppMiniGameWinAndLossLifecycle(t *testing.T) {
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

	initialHearts := app.State().FriendshipHearts

	// Launch Look Away and simulate a win
	app.LaunchGameByID(3)
	lookAway, ok := app.ActiveGame().(interface{ SimulateRound(bool) })
	if !ok {
		t.Fatalf("active game does not implement SimulateRound")
	}
	for i := 0; i < 3; i++ {
		lookAway.SimulateRound(true)
	}
	for i := 0; i < 2; i++ {
		lookAway.SimulateRound(false)
	}

	// Update app to process game finish
	if err := app.Update(); err != nil {
		t.Fatalf("app update failed: %v", err)
	}

	if app.ActiveGame() != nil {
		t.Fatalf("expected game to be cleared after finish")
	}
	if app.Speech() != "You won!" {
		t.Errorf("expected win speech 'You won!', got %q", app.Speech())
	}
	if app.State().FriendshipHearts != initialHearts+1 {
		t.Errorf("expected hearts to increment to %d, got %d", initialHearts+1, app.State().FriendshipHearts)
	}
	if app.State().HighScores["Look Away"] <= 0 {
		t.Errorf("expected high score for Look Away, got %d", app.State().HighScores["Look Away"])
	}

	// Test loss condition: launch game 3 again and lose
	app.LaunchGameByID(3)
	lookAway2, ok := app.ActiveGame().(interface{ SimulateRound(bool) })
	if !ok {
		t.Fatalf("active game does not implement SimulateRound")
	}
	for i := 0; i < 5; i++ {
		lookAway2.SimulateRound(false)
	}

	if err := app.Update(); err != nil {
		t.Fatalf("app update failed: %v", err)
	}

	if app.ActiveGame() != nil {
		t.Fatalf("expected game to be cleared after finish")
	}
	if app.Speech() != "Good try!" {
		t.Errorf("expected loss speech 'Good try!', got %q", app.Speech())
	}
	// Hearts should not increment
	if app.State().FriendshipHearts != initialHearts+1 {
		t.Errorf("hearts should remain unchanged on loss, got %d", app.State().FriendshipHearts)
	}
}

