package cli

import (
	"github.com/dpologdvinity/promptpet/internal/store"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
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
