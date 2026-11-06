package cli

import (
	"github.com/dpologdvinity/promptpet/internal/event"
	"github.com/dpologdvinity/promptpet/internal/store"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestPromptFirstRun(t *testing.T) {
	d, out, errb := testDeps(t)
	if Run([]string{"prompt", "--exit", "0"}, d) != 0 {
		t.Fatal()
	}
	if out.String() != "(=^‿^=) ♥♥♥\n" || errb.Len() != 0 {
		t.Fatalf("%q %q", out, errb)
	} // 80/80/100 -> 86.7 -> 3 hearts, Happy at noon
}
func TestPromptMealAndExit(t *testing.T) {
	d, out, _ := testDeps(t)
	dir := d.Getenv("PROMPTPET_HOME")
	Run([]string{"prompt"}, d)
	out.Reset()
	store.AppendMeals(dir, []store.MealLine{{Unix: 1, SHA: "aaaaaaa"}})
	Run([]string{"prompt", "--exit", "1"}, d)
	if !strings.HasPrefix(out.String(), "(=^･ω･^=) nom ") {
		t.Fatalf("%q", out)
	}
	out.Reset()
	Run([]string{"prompt", "--exit", "127"}, d)
	if !strings.HasPrefix(out.String(), "(=･ω･=)? ") {
		t.Fatalf("%q", out)
	} // meal consumed once
	s, _, _ := store.LoadState(dir, t0)
	if s.Meals != 1 || s.Health != 88 {
		t.Fatalf("%+v", s)
	} // 100 -6 -6
}
func TestPromptSpawnsOnRepoChange(t *testing.T) {
	d, _, _ := testDeps(t)
	wt := t.TempDir()
	os.MkdirAll(filepath.Join(wt, ".git", "logs"), 0o755)
	os.WriteFile(filepath.Join(wt, ".git", "logs", "HEAD"), []byte("a\n"), 0o644)
	d.Getwd = func() (string, error) { return wt, nil }
	Run([]string{"prompt"}, d)
	Run([]string{"prompt"}, d)
	if got := spawned(d); len(got) != 1 || got[0] != wt {
		t.Fatal(got)
	} // unchanged sig: no second spawn
	os.WriteFile(filepath.Join(wt, ".git", "logs", "HEAD"), []byte("a\nb\n"), 0o644)
	Run([]string{"prompt"}, d)
	if got := spawned(d); len(got) != 2 {
		t.Fatal(got)
	}
}
func TestPromptCwdGone(t *testing.T) {
	d, out, _ := testDeps(t)
	d.Getwd = func() (string, error) { return "", os.ErrNotExist }
	if Run([]string{"prompt"}, d) != 0 || out.Len() == 0 {
		t.Fatal("no segment")
	}
}
func TestPromptLockBusy(t *testing.T) {
	d, out, _ := testDeps(t)
	unlock, _ := store.Lock(d.Getenv("PROMPTPET_HOME"), "state.lock", time.Second)
	defer unlock()
	start := time.Now()
	Run([]string{"prompt"}, d)
	if out.Len() == 0 || time.Since(start) > 50*time.Millisecond {
		t.Fatal("must print within budget")
	}
	if _, err := os.Stat(filepath.Join(d.Getenv("PROMPTPET_HOME"), "state.json")); err == nil {
		t.Fatal("saved without lock")
	}
}
func TestPromptDisabledAndCorrupt(t *testing.T) {
	d, out, errb := testDeps(t)
	os.WriteFile(filepath.Join(d.Getenv("PROMPTPET_HOME"), "state.json"), []byte("{garbage"), 0o600)
	Run([]string{"prompt", "--exit", "x"}, d)
	if out.Len() == 0 || errb.Len() != 0 {
		t.Fatalf("%q %q", out, errb)
	}
	out.Reset()
	d.Getenv = withEnv(d.Getenv, "PROMPTPET_DISABLE", "1")
	if Run([]string{"prompt"}, d) != 0 || out.Len() != 0 {
		t.Fatal(out)
	}
}
func TestPromptBudget(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	d, _, _ := testDeps(t)
	d.Stdout = io.Discard
	var times []time.Duration
	for i := 0; i < 20; i++ {
		s := time.Now()
		Run([]string{"prompt", "--exit", "0"}, d)
		times = append(times, time.Since(s))
	}
	slices.Sort(times)
	if times[10] >= 10*time.Millisecond {
		t.Fatalf("median %v >= 10ms", times[10])
	}
}
func BenchmarkPrompt(b *testing.B) {
	home := b.TempDir()
	d := Deps{Now: time.Now, Rand: rand.New(rand.NewPCG(1, 2)), Getenv: func(k string) string {
		if k == "PROMPTPET_HOME" {
			return home
		}
		return ""
	}, Getwd: func() (string, error) { return home, nil }, Spawn: func(string) error { return nil }, Stdout: io.Discard, Stderr: io.Discard}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Run([]string{"prompt"}, d)
	}
}

func TestPromptLockBusyDoesNotRecoverCorruptState(t *testing.T) {
	d, out, _ := testDeps(t)
	dir := d.Getenv("PROMPTPET_HOME")
	path := filepath.Join(dir, "state.json")
	if err := os.WriteFile(path, []byte("{broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	unlock, err := store.Lock(dir, "state.lock", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	if code := Run([]string{"prompt"}, d); code != 0 || out.Len() == 0 {
		t.Fatalf("code=%d output=%q", code, out.String())
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "{broken" {
		t.Fatalf("state changed: %q %v", got, err)
	}
	backups, err := filepath.Glob(path + ".corrupt-*")
	if err != nil || len(backups) != 0 {
		t.Fatalf("backups=%v err=%v", backups, err)
	}
}

// promptEvents runs `promptpet prompt` with args in d and returns every event
// appended to the test's PROMPTPET_EVENT_FILE so far.
func promptEvents(t *testing.T, d *Deps, args ...string) []event.Event {
	t.Helper()
	evFile := filepath.Join(d.Getenv("PROMPTPET_HOME"), "events-test.jsonl")
	if d.Getenv("PROMPTPET_EVENT_FILE") == "" {
		d.Getenv = withEnv(d.Getenv, "PROMPTPET_EVENT_FILE", evFile)
	}
	if code := Run(append([]string{"prompt"}, args...), *d); code != 0 {
		t.Fatalf("prompt exit %d", code)
	}
	evs, err := event.ReadEventsSince(evFile, 0)
	if err != nil {
		t.Fatal(err)
	}
	return evs
}

func TestPromptDurationFlag(t *testing.T) {
	d, _, _ := testDeps(t)
	promptEvents(t, &d, "--exit", "0", "--duration-ms", "1500")
	promptEvents(t, &d, "--exit=2", "--duration-ms=42")
	promptEvents(t, &d, "--exit", "0")
	evs := promptEvents(t, &d, "--exit", "0", "--duration-ms=-5")
	if len(evs) != 4 {
		t.Fatalf("events: %+v", evs)
	}
	if evs[0].DurationMs != 1500 || evs[1].DurationMs != 42 || evs[1].ExitCode != 2 || evs[2].DurationMs != 0 || evs[3].DurationMs != 0 {
		t.Fatalf("events: %+v", evs)
	}
	raw, _ := os.ReadFile(filepath.Join(d.Getenv("PROMPTPET_HOME"), "events-test.jsonl"))
	if lines := strings.Split(string(raw), "\n"); strings.Contains(lines[2], "duration_ms") {
		t.Fatalf("unknown duration must be omitted: %s", lines[2])
	}
}

func TestPromptGitBranchAndMergeFireOnce(t *testing.T) {
	d, _, _ := testDeps(t)
	wt := t.TempDir()
	gd := filepath.Join(wt, ".git")
	os.MkdirAll(filepath.Join(gd, "logs", "refs", "heads"), 0o755)
	os.WriteFile(filepath.Join(gd, "HEAD"), []byte("ref: refs/heads/feat\n"), 0o644)
	line := func(msg string) string {
		return "a b U <u@x> " + strconv.FormatInt(t0.Unix(), 10) + " +0000\t" + msg + "\n"
	}
	os.WriteFile(filepath.Join(gd, "logs", "refs", "heads", "feat"), []byte(line("branch: Created from main")), 0o644)
	headLog := line("commit: one") + line("checkout: moving from main to feat")
	os.WriteFile(filepath.Join(gd, "logs", "HEAD"), []byte(headLog), 0o644)
	d.Getwd = func() (string, error) { return wt, nil }

	promptEvents(t, &d, "--exit", "0")
	promptEvents(t, &d, "--exit", "0") // reflog unchanged: must not refire
	headLog += line("merge other: Merge made by the 'ort' strategy.")
	os.WriteFile(filepath.Join(gd, "logs", "HEAD"), []byte(headLog), 0o644)
	evs := promptEvents(t, &d, "--exit", "0")
	if len(evs) != 3 {
		t.Fatalf("events: %+v", evs)
	}
	if evs[0].GitAction != "branch" || evs[0].GitBranch != "feat" || evs[0].CommitCount != 1 {
		t.Fatalf("first: %+v", evs[0])
	}
	if evs[1].GitAction != "" {
		t.Fatalf("refired: %+v", evs[1])
	}
	if evs[2].GitAction != "merge" || evs[2].GitBranch != "feat" {
		t.Fatalf("merge: %+v", evs[2])
	}
}
