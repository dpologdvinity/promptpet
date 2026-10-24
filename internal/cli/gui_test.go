package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dpologdvinity/promptpet/internal/cli"
	"github.com/dpologdvinity/promptpet/internal/companion"
	"github.com/dpologdvinity/promptpet/internal/event"
	"github.com/dpologdvinity/promptpet/internal/gui"
	"github.com/dpologdvinity/promptpet/internal/store"
)

func TestPetSwitchCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROMPTPET_HOME", tmpDir)
	t.Setenv("PROMPTPET_COMPANION_STATE_FILE", filepath.Join(tmpDir, "companion.json"))

	var out, errOut bytes.Buffer
	code := cli.Run([]string{"pet", "kochi"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0 exit code, got %d: %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "Kochi") {
		t.Fatalf("expected output to mention Kochi, got %q", out.String())
	}

	state, err := companion.LoadState(filepath.Join(tmpDir, "companion.json"))
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	if state.ActivePet != "kochi" {
		t.Fatalf("expected active pet kochi, got %q", state.ActivePet)
	}

	// Invalid pet should fail
	out.Reset()
	errOut.Reset()
	code = cli.Run([]string{"pet", "nonexistent"}, &out, &errOut)
	if code == 0 {
		t.Fatalf("expected non-zero exit code for invalid pet, got 0")
	}
}

func TestPetListCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROMPTPET_HOME", tmpDir)
	statePath := filepath.Join(tmpDir, "companion.json")
	t.Setenv("PROMPTPET_COMPANION_STATE_FILE", statePath)

	// Save initial state with boba active
	s := companion.NewDefaultState()
	s.ActivePet = "boba"
	if err := companion.SaveState(statePath, s); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	var out, errOut bytes.Buffer
	code := cli.Run([]string{"pet"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0 exit code, got %d: %s", code, errOut.String())
	}

	output := out.String()
	for _, pet := range companion.GetAllPets() {
		if !strings.Contains(output, string(pet.ID)) && !strings.Contains(output, pet.Name) {
			t.Errorf("expected output to contain pet %s (%s)", pet.ID, pet.Name)
		}
	}

	// Should highlight active pet (boba)
	if !strings.Contains(output, "*") || !strings.Contains(output, "boba") {
		t.Errorf("expected active pet boba to be highlighted, got:\n%s", output)
	}
}

func TestGamesListCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROMPTPET_HOME", tmpDir)
	statePath := filepath.Join(tmpDir, "companion.json")
	t.Setenv("PROMPTPET_COMPANION_STATE_FILE", statePath)

	s := companion.NewDefaultState()
	s.HighScores["Snack Catch"] = 42
	s.HighScores["Whack-a-Bug"] = 99
	if err := companion.SaveState(statePath, s); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	var out, errOut bytes.Buffer
	code := cli.Run([]string{"games"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0 exit code, got %d: %s", code, errOut.String())
	}

	output := out.String()
	expectedGames := []string{
		"Snack Catch",
		"Puddle Leap",
		"Look Away",
		"Rhythm Dance",
		"Memory Simon",
		"Whack-a-Bug",
		"Balloon Pop",
		"Word Snack",
	}
	for _, g := range expectedGames {
		if !strings.Contains(output, g) {
			t.Errorf("expected games list to contain %q", g)
		}
	}
	if !strings.Contains(output, "42") || !strings.Contains(output, "99") {
		t.Errorf("expected games list to contain high scores 42 and 99, got:\n%s", output)
	}
}

func TestGUICommand_Headless(t *testing.T) {
	t.Setenv("PROMPTPET_HEADLESS", "1")

	var out, errOut bytes.Buffer
	code := cli.Run([]string{"gui"}, &out, &errOut)
	if code != 1 {
		t.Fatalf("expected exit code 1 in headless mode, got %d", code)
	}
	expectedMsg := "promptpet gui requires a graphical display (X11/Wayland/WSLg)"
	if !strings.Contains(errOut.String(), expectedMsg) {
		t.Fatalf("expected stderr to contain %q, got %q", expectedMsg, errOut.String())
	}
}

func TestGUICommand_MockRun(t *testing.T) {
	tmpDir := t.TempDir()
	var out, errOut bytes.Buffer
	d := cli.RealDeps("dev")
	d.Stdout = &out
	d.Stderr = &errOut
	d.Getenv = func(k string) string {
		if k == "PROMPTPET_HOME" {
			return tmpDir
		}
		return os.Getenv(k)
	}

	called := false
	d.RunGUI = func(cfg gui.Config) error {
		called = true
		return nil
	}

	code := cli.Run([]string{"gui"}, d)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, errOut.String())
	}
	if !called {
		t.Fatalf("expected d.RunGUI to be called")
	}
}

func TestPromptLoggingEvent(t *testing.T) {
	tmpDir := t.TempDir()
	evFile := filepath.Join(tmpDir, "events.jsonl")
	t.Setenv("PROMPTPET_EVENT_FILE", evFile)
	t.Setenv("PROMPTPET_HOME", tmpDir)

	var out, errOut bytes.Buffer
	code := cli.Run([]string{"prompt", "--exit", "0"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, errOut.String())
	}

	events, err := event.ReadEventsSince(evFile, 0)
	if err != nil {
		t.Fatalf("failed to read events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 logged event, got %d", len(events))
	}
	if events[0].Type != "command_finish" || events[0].ExitCode != 0 {
		t.Fatalf("unexpected event: %+v", events[0])
	}
}

func TestStateFileIsolation(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROMPTPET_HOME", tmpDir)

	// 1. Run prompt -> writes shell pet state to state.json
	var out, errOut bytes.Buffer
	code := cli.Run([]string{"prompt", "--exit", "0"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("prompt run failed: %d", code)
	}

	shellStatePath := filepath.Join(tmpDir, "state.json")
	now := time.Now()
	shellState, _, err := store.LoadState(tmpDir, now)
	if err != nil {
		t.Fatalf("failed to load shell pet state: %v", err)
	}
	if shellState.Name == "" {
		t.Fatalf("expected non-empty shell pet name")
	}

	// 2. Run pet command -> writes companion state to companion.json
	out.Reset()
	errOut.Reset()
	code = cli.Run([]string{"pet", "kochi"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("pet command failed: %d: %s", code, errOut.String())
	}

	companionStatePath := filepath.Join(tmpDir, "companion.json")
	if _, err := os.Stat(companionStatePath); err != nil {
		t.Fatalf("companion.json was not created: %v", err)
	}

	compState, err := companion.LoadState(companionStatePath)
	if err != nil {
		t.Fatalf("failed to load companion state: %v", err)
	}
	if compState.ActivePet != "kochi" {
		t.Errorf("expected companion active pet kochi, got %q", compState.ActivePet)
	}

	// 3. Ensure shell pet state in state.json is completely unchanged and unharmed
	shellStateReloaded, _, err := store.LoadState(tmpDir, now)
	if err != nil {
		t.Fatalf("failed to reload shell pet state: %v", err)
	}
	if shellStateReloaded.Name != shellState.Name {
		t.Errorf("shell pet state was corrupted or overwritten: got %+v", shellStateReloaded)
	}

	// 4. Check raw file contents are different files
	if shellStatePath == companionStatePath {
		t.Fatalf("shell state path and companion state path must not be identical")
	}
}

func TestPromptLoggingEventGitMilestone(t *testing.T) {
	tmpDir := t.TempDir()
	evFile := filepath.Join(tmpDir, "events.jsonl")
	t.Setenv("PROMPTPET_EVENT_FILE", evFile)
	t.Setenv("PROMPTPET_HOME", tmpDir)

	// Create fake git repo
	repoDir := filepath.Join(tmpDir, "myrepo")
	gitDir := filepath.Join(repoDir, ".git")
	logsDir := filepath.Join(gitDir, "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// HEAD pointing to main branch
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// logs/HEAD with 5 commits
	var logContent strings.Builder
	for i := 1; i <= 5; i++ {
		logContent.WriteString("0000000 1111111 User <u@example.com> 1234567890 +0000\tcommit: commit " + string(rune('0'+i)) + "\n")
	}
	if err := os.WriteFile(filepath.Join(logsDir, "HEAD"), []byte(logContent.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	d := cli.RealDeps("dev")
	d.Stdout = &out
	d.Stderr = &errOut
	d.Getenv = func(k string) string {
		if k == "PROMPTPET_EVENT_FILE" {
			return evFile
		}
		if k == "PROMPTPET_HOME" {
			return tmpDir
		}
		return os.Getenv(k)
	}
	d.Getwd = func() (string, error) {
		return repoDir, nil
	}

	code := cli.Run([]string{"prompt", "--exit", "0"}, d)
	if code != 0 {
		t.Fatalf("prompt run failed: %d", code)
	}

	events, err := event.ReadEventsSince(evFile, 0)
	if err != nil {
		t.Fatalf("failed to read events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 logged event, got %d", len(events))
	}
	ev := events[0]
	if ev.GitAction != "commit" {
		t.Errorf("expected GitAction 'commit', got %q", ev.GitAction)
	}
	if ev.GitBranch != "main" {
		t.Errorf("expected GitBranch 'main', got %q", ev.GitBranch)
	}
	if ev.CommitCount != 5 {
		t.Errorf("expected CommitCount 5, got %d", ev.CommitCount)
	}
}
