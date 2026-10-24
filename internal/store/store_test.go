package store

import (
	"errors"
	"fmt"
	"github.com/dpologdvinity/promptpet/internal/pet"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var t0 = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func TestDir(t *testing.T) {
	if d, _ := Dir(env(map[string]string{"PROMPTPET_HOME": "/x/pp", "XDG_STATE_HOME": "/s", "HOME": "/h"})); d != "/x/pp" {
		t.Fatal(d)
	}
	if d, _ := Dir(env(map[string]string{"XDG_STATE_HOME": "/s", "HOME": "/h"})); d != "/s/promptpet" {
		t.Fatal(d)
	}
	if d, _ := Dir(env(map[string]string{"XDG_STATE_HOME": "rel", "HOME": "/h"})); d != "/h/.local/state/promptpet" {
		t.Fatal(d)
	}
	if _, err := Dir(env(map[string]string{})); err == nil {
		t.Fatal("want error without HOME")
	}
}
func TestSaveLoadRoundTrip(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "pp")
	unlock, err := Lock(dir, "state.lock", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	s := pet.New(t0)
	s.Name = "Mochi"
	s.Food = 33.5
	if err := SaveState(dir, s); err != nil {
		t.Fatal(err)
	}
	got, moved, err := LoadState(dir, t0.Add(time.Hour))
	if err != nil || moved != "" || got.Name != "Mochi" || got.Food != 33.5 {
		t.Fatalf("%+v %q %v", got, moved, err)
	}
	if fi, _ := os.Stat(dir); fi.Mode().Perm() != 0o700 {
		t.Fatal(fi.Mode())
	}
	if fi, _ := os.Stat(filepath.Join(dir, "state.json")); fi.Mode().Perm() != 0o600 {
		t.Fatal(fi.Mode())
	}
}
func TestLoadStateMissing(t *testing.T) {
	s, moved, err := LoadState(t.TempDir(), t0)
	if err != nil || moved != "" || s.Name != "Pip" {
		t.Fatal(s, moved, err)
	}
}
func TestLoadStateCorrupt(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "state.json"), []byte(`{"version":1,"name":"Mo`), 0o600)
	s, moved, err := LoadState(dir, t0)
	if err != nil || s.Name != "Pip" {
		t.Fatal(s, err)
	}
	if moved != filepath.Join(dir, fmt.Sprintf("state.json.corrupt-%d", t0.Unix())) {
		t.Fatal(moved)
	}
	if _, err := os.Stat(moved); err != nil {
		t.Fatal(err)
	}
}
func TestLoadStateNewer(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "state.json"), []byte(`{"version":2}`), 0o600)
	if _, _, err := LoadState(dir, t0); !errors.Is(err, ErrNewer) {
		t.Fatal(err)
	}
}
func TestLockBusy(t *testing.T) {
	dir := t.TempDir()
	unlock, _ := Lock(dir, "state.lock", time.Second)
	defer unlock()
	start := time.Now()
	if _, err := Lock(dir, "state.lock", 5*time.Millisecond); !errors.Is(err, ErrBusy) {
		t.Fatal(err)
	}
	if time.Since(start) > 50*time.Millisecond {
		t.Fatal("lock wait exceeded budget")
	}
}

func TestLoadStateCorruptPreservesBothBackupsInSameSecond(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	if err := os.WriteFile(path, []byte("{first"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, first, err := LoadState(dir, t0)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{second"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, second, err := LoadState(dir, t0)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatalf("backup collision: %s", first)
	}
	for path, want := range map[string]string{first: "{first", second: "{second"} {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != want {
			t.Fatalf("%s: %q %v", path, got, err)
		}
	}
}
