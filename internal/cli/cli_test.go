package cli

import (
	"bytes"
	"github.com/dpologdvinity/promptpet/internal/hooks"
	"github.com/dpologdvinity/promptpet/internal/pet"
	"github.com/dpologdvinity/promptpet/internal/store"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var spawnCalls = map[string][]string{}

func spawned(d Deps) []string { return spawnCalls[d.Getenv("PROMPTPET_HOME")] }
func withEnv(getenv func(string) string, k, v string) func(string) string {
	return func(x string) string {
		if x == k {
			return v
		}
		return getenv(x)
	}
}

var t0 = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

func testDeps(t *testing.T) (Deps, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	out, errb := new(bytes.Buffer), new(bytes.Buffer)
	home, cwd := t.TempDir(), t.TempDir()
	d := Deps{Version: "dev", Now: func() time.Time { return t0 }, Rand: rand.New(rand.NewPCG(1, 2)), Getenv: func(k string) string {
		if k == "PROMPTPET_HOME" {
			return home
		}
		return ""
	}, Getwd: func() (string, error) { return cwd, nil }, Spawn: func(wt string) error { spawnCalls[home] = append(spawnCalls[home], wt); return nil }, Stdout: out, Stderr: errb}
	return d, out, errb
}
func TestVersion(t *testing.T) {
	d, out, _ := testDeps(t)
	d.Version = "1.2.3"
	if code := Run([]string{"--version"}, d); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if out.String() != "promptpet 1.2.3\n" {
		t.Fatalf("got %q", out.String())
	}
}
func TestHelpAndUnknown(t *testing.T) {
	d, out, errb := testDeps(t)
	if Run([]string{"--help"}, d) != 0 || !strings.HasPrefix(out.String(), "promptpet - a tiny pet that lives in your shell prompt\n") {
		t.Fatal(out.String())
	}
	if Run([]string{"bogus"}, d) != 2 || errb.String() != "promptpet: unknown command \"bogus\" (see promptpet --help)\n" {
		t.Fatal(errb.String())
	}
}

func TestCardCommand(t *testing.T) {
	d, out, _ := testDeps(t)
	if Run(nil, d) != 0 || !strings.Contains(out.String(), "Pip the cat - feeling great\n") {
		t.Fatal(out)
	}
}
func TestInteractions(t *testing.T) {
	d, out, errb := testDeps(t)
	if Run([]string{"name", "  Mochi "}, d) != 0 || out.String() != "Your pet is now called Mochi.\n" {
		t.Fatal(out)
	}
	out.Reset()
	Run([]string{"species", "blob"}, d)
	if out.String() != "Mochi is now a blob.\n" {
		t.Fatal(out)
	}
	out.Reset()
	Run([]string{"pet"}, d)
	if out.String() != "Mochi wobbles contentedly.\n" {
		t.Fatal(out)
	}
	if Run([]string{"species", "dog"}, d) != 2 || errb.String() != "promptpet: unknown species \"dog\" (choose cat, dragon or blob)\n" {
		t.Fatal(errb)
	}
	errb.Reset()
	if Run([]string{"name", "a$b"}, d) != 2 || errb.String() != "promptpet: names can use letters, digits, spaces and - _ ' . (1-20 characters)\n" {
		t.Fatal(errb)
	}
	errb.Reset()
	if Run([]string{"name"}, d) != 2 || errb.String() != "promptpet: usage: promptpet name <name>\n" {
		t.Fatal(errb)
	}
}
func TestCardShowsMessageOnce(t *testing.T) {
	d, out, _ := testDeps(t)
	dir := d.Getenv("PROMPTPET_HOME")
	s := pet.New(t0)
	s.Away = true
	s.Food = 0
	unlock, _ := store.Lock(dir, "state.lock", time.Second)
	store.SaveState(dir, s)
	unlock()
	Run([]string{"pet"}, d)
	out.Reset()
	Run(nil, d)
	if !strings.Contains(out.String(), "came back from an adventure") {
		t.Fatal(out)
	}
	out.Reset()
	Run(nil, d)
	if strings.Contains(out.String(), "came back from an adventure") {
		t.Fatal("message shown twice")
	}
}
func TestInit(t *testing.T) {
	d, out, errb := testDeps(t)
	want, _ := hooks.Script("fish")
	if Run([]string{"init", "fish"}, d) != 0 || out.String() != want {
		t.Fatal(out)
	}
	if Run([]string{"init", "tcsh"}, d) != 2 || errb.String() != "promptpet: init needs a shell: zsh, bash or fish\n" {
		t.Fatal(errb)
	}
}
func TestCompletion(t *testing.T) {
	d, out, errb := testDeps(t)
	want, _ := hooks.Completion("bash")
	if Run([]string{"completion", "bash"}, d) != 0 || out.String() != want {
		t.Fatal(out)
	}
	if Run([]string{"completion", "tcsh"}, d) != 2 || errb.String() != "promptpet: completion needs a shell: zsh, bash or fish\n" {
		t.Fatal(errb)
	}
	errb.Reset()
	if Run([]string{"completion"}, d) != 2 || errb.String() != "promptpet: completion needs a shell: zsh, bash or fish\n" {
		t.Fatal(errb)
	}
}
func TestCorruptStateWarns(t *testing.T) {
	d, _, errb := testDeps(t)
	os.WriteFile(filepath.Join(d.Getenv("PROMPTPET_HOME"), "state.json"), []byte("{"), 0o600)
	if Run(nil, d) != 0 || !strings.HasPrefix(errb.String(), "promptpet: your pet's state file was unreadable; moved it to ") {
		t.Fatal(errb)
	}
}
func TestFeedCommandSilent(t *testing.T) {
	d, out, errb := testDeps(t)
	if Run([]string{"_feed", t.TempDir()}, d) != 0 || out.Len()+errb.Len() != 0 {
		t.Fatal(out, errb)
	} // not a repo: silent
}
