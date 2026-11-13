package cli_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dpologdvinity/promptpet/internal/cli"
	"github.com/dpologdvinity/promptpet/internal/companion"
	"github.com/dpologdvinity/promptpet/internal/event"
	"github.com/dpologdvinity/promptpet/internal/store"
	"github.com/dpologdvinity/promptpet/internal/window"
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
	code := cli.Run([]string{"pet", "list"}, &out, &errOut)
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

	// Calling `promptpet pet` without arguments pets the shell pet
	out.Reset()
	code = cli.Run([]string{"pet"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0 exit code for petting, got %d: %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "purrs") && !strings.Contains(out.String(), "rubs") {
		t.Errorf("expected pet reaction output, got: %s", out.String())
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

	t.Setenv("PROMPTPET_HEADLESS", "")
	t.Setenv("DISPLAY", ":0")
	called := false
	d.RunGUI = func() error {
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
		// CommitCount counts today's commits, so stamp them now.
		logContent.WriteString("0000000 1111111 User <u@example.com> " + strconv.FormatInt(time.Now().Unix(), 10) + " +0000\tcommit: commit " + string(rune('0'+i)) + "\n")
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

func TestStatusCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROMPTPET_HOME", tmpDir)

	var out, errOut bytes.Buffer
	code := cli.Run([]string{"status"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0 exit code, got %d: %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "Pip") {
		t.Errorf("expected output to contain pet name Pip, got: %s", out.String())
	}
}

func TestPromptExitEqualSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	evFile := filepath.Join(tmpDir, "events.jsonl")
	t.Setenv("PROMPTPET_HOME", tmpDir)
	t.Setenv("PROMPTPET_EVENT_FILE", evFile)

	var out, errOut bytes.Buffer
	code := cli.Run([]string{"prompt", "--exit=42"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0 exit code, got %d: %s", code, errOut.String())
	}

	events, err := event.ReadEventsSince(evFile, 0)
	if err != nil || len(events) != 1 {
		t.Fatalf("expected 1 logged event, got %d (err: %v)", len(events), err)
	}
	if events[0].ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", events[0].ExitCode)
	}
}

func TestStatusShowsHighScoresAndCasing(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROMPTPET_HOME", tmpDir)
	sp := filepath.Join(tmpDir, "companion.json")
	t.Setenv("PROMPTPET_COMPANION_STATE_FILE", sp)

	s := companion.NewDefaultState()
	s.Casing = "mint"
	s.HighScores["Balloon Pop"] = 1100
	if err := companion.SaveState(sp, s); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	if code := cli.Run([]string{"status"}, &out, &errOut); code != 0 {
		t.Fatalf("expected 0 exit code, got %d: %s", code, errOut.String())
	}
	for _, want := range []string{"Casing: mint", "Balloon Pop", "1100", "Snack Catch"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("status output missing %q:\n%s", want, out.String())
		}
	}
}

func TestPetSwitchRefusesWhenLocked(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROMPTPET_HOME", tmpDir)
	sp := filepath.Join(tmpDir, "companion.json")
	t.Setenv("PROMPTPET_COMPANION_STATE_FILE", sp)

	unlock, err := store.Lock(tmpDir, "companion.lock", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()

	var out, errOut bytes.Buffer
	if code := cli.Run([]string{"pet", "kochi"}, &out, &errOut); code == 0 {
		t.Fatal("expected non-zero exit when the companion lock is held")
	}
	if s, _ := companion.LoadState(sp); s.ActivePet == "kochi" {
		t.Fatal("state must not be written without the lock")
	}
}

func TestGUIDetach(t *testing.T) {
	tmpDir := t.TempDir()
	var out, errOut bytes.Buffer
	d := cli.RealDeps("dev")
	d.Stdout, d.Stderr = &out, &errOut
	d.Getenv = func(k string) string {
		if k == "PROMPTPET_HOME" {
			return tmpDir
		}
		return os.Getenv(k)
	}
	d.RunGUI = func() error { t.Fatal("--detach must not wait for the window"); return nil }
	detached := false
	d.Detach = func() error { detached = true; return nil }
	t.Setenv("PROMPTPET_HEADLESS", "")
	t.Setenv("DISPLAY", ":0")

	if code := cli.Run([]string{"gui", "--detach"}, d); code != 0 {
		t.Fatalf("expected 0, got %d: %s", code, errOut.String())
	}
	if !detached {
		t.Fatal("expected promptpet-gui to be started detached")
	}

	// Headless: refuse up front instead of spawning a child that dies silently.
	detached = false
	t.Setenv("PROMPTPET_HEADLESS", "1")
	if code := cli.Run([]string{"gui", "--detach"}, d); code != 1 || detached {
		t.Fatalf("expected headless refusal, got code %d detached %v", code, detached)
	}

	if code := cli.Run([]string{"gui", "--bogus"}, d); code == 0 {
		t.Fatal("expected usage error for unknown gui flag")
	}
}

// TestPromptRepoSwitchNoStaleMilestone: cd-ing into another repo changes the reflog
// signature, but its newest reflog entry happened before the last prompt, so no
// milestone may fire (and none on coming back either).
func TestPromptRepoSwitchNoStaleMilestone(t *testing.T) {
	tmpDir := t.TempDir()
	evFile := filepath.Join(tmpDir, "events.jsonl")
	t.Setenv("PROMPTPET_EVENT_FILE", evFile)
	t.Setenv("PROMPTPET_HOME", tmpDir)

	mkRepo := func(name, entry string, ts time.Time) string {
		repoDir := filepath.Join(tmpDir, name)
		gitDir := filepath.Join(repoDir, ".git")
		if err := os.MkdirAll(filepath.Join(gitDir, "logs"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		line := "0000000 1111111 User <u@example.com> " + strconv.FormatInt(ts.Unix(), 10) + " +0000\t" + entry + "\n"
		if err := os.WriteFile(filepath.Join(gitDir, "logs", "HEAD"), []byte(line), 0o644); err != nil {
			t.Fatal(err)
		}
		return repoDir
	}
	repoA := mkRepo("a", "commit: fresh work", time.Now().Add(-10*time.Second))
	repoB := mkRepo("b", "merge feat: Merge made by the 'ort' strategy.", time.Now().Add(-48*time.Hour))

	cwd := repoA
	d := cli.RealDeps("dev")
	d.Stdout, d.Stderr = io.Discard, io.Discard
	d.Getenv = func(k string) string {
		switch k {
		case "PROMPTPET_EVENT_FILE":
			return evFile
		case "PROMPTPET_HOME":
			return tmpDir
		}
		return os.Getenv(k)
	}
	d.Getwd = func() (string, error) { return cwd, nil }

	for _, dir := range []string{repoA, repoB, repoA} {
		cwd = dir
		if code := cli.Run([]string{"prompt", "--exit", "0"}, d); code != 0 {
			t.Fatalf("prompt in %s failed: %d", dir, code)
		}
	}

	events, err := event.ReadEventsSince(evFile, 0)
	if err != nil || len(events) != 3 {
		t.Fatalf("expected 3 events, got %d (err %v)", len(events), err)
	}
	if events[0].GitAction != "commit" {
		t.Errorf("first prompt after a fresh commit should report it, got %q", events[0].GitAction)
	}
	for i, ev := range events[1:] {
		if ev.GitAction != "" {
			t.Errorf("event %d after repo switch: stale GitAction %q", i+1, ev.GitAction)
		}
	}
}

// TestGUIRefusesWhileWindowOpen: while a window holds companion.lock, `promptpet gui`
// refuses before launching promptpet-gui. (The window side is tested in internal/window.)
func TestGUIRefusesWhileWindowOpen(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROMPTPET_HOME", tmpDir)
	t.Setenv("PROMPTPET_COMPANION_STATE_FILE", filepath.Join(tmpDir, "companion.json"))
	t.Setenv("PROMPTPET_HEADLESS", "")
	t.Setenv("DISPLAY", ":0")
	unlock, err := window.Lock(os.Getenv)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()

	var out, errOut bytes.Buffer
	d := cli.RealDeps("dev")
	d.Stdout, d.Stderr, d.Getenv = &out, &errOut, os.Getenv
	d.RunGUI = func() error { t.Error("must not launch a second window"); return nil }
	d.Detach = func() error { t.Error("must not launch a second window"); return nil }
	for _, args := range [][]string{{"gui"}, {"gui", "--detach"}} {
		if code := cli.Run(args, d); code != 1 || !strings.Contains(errOut.String(), "already running") {
			t.Errorf("%v: code %d, stderr %q", args, code, errOut.String())
		}
	}
}

// TestGUIReportsMissingBinaryAndChildExit: a missing promptpet-gui gets an install
// hint, and promptpet-gui's own exit code is passed through without a second message.
func TestGUIReportsMissingBinaryAndChildExit(t *testing.T) {
	t.Setenv("PROMPTPET_HOME", t.TempDir())
	t.Setenv("PROMPTPET_HEADLESS", "")
	t.Setenv("DISPLAY", ":0")
	var out, errOut bytes.Buffer
	d := cli.RealDeps("dev")
	d.Stdout, d.Stderr, d.Getenv = &out, &errOut, os.Getenv

	d.RunGUI = func() error { return window.ErrNotInstalled }
	if code := cli.Run([]string{"gui"}, d); code != 1 || !strings.Contains(errOut.String(), "go install github.com/dpologdvinity/promptpet/cmd/promptpet-gui@latest") {
		t.Fatalf("missing binary: code %d, stderr %q", code, errOut.String())
	}

	errOut.Reset()
	d.RunGUI = func() error { return exec.Command("sh", "-c", "exit 3").Run() }
	if code := cli.Run([]string{"gui"}, d); code != 3 || errOut.Len() != 0 {
		t.Fatalf("child exit: code %d, stderr %q", code, errOut.String())
	}
}
