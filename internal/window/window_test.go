package window_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpologdvinity/promptpet/internal/cli"
	"github.com/dpologdvinity/promptpet/internal/window"
)

func writeExe(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), mode); err != nil {
		t.Fatal(err)
	}
}

func noPath(string) (string, error) { return "", errors.New("not found") }

func TestFindBinary(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "promptpet")
	writeExe(t, exe, 0o755)
	gui := filepath.Join(dir, window.BinaryName)

	// Nothing next to promptpet and nothing on PATH: an install hint.
	if _, err := window.FindBinary(exe, noPath); !errors.Is(err, window.ErrNotInstalled) {
		t.Fatalf("expected ErrNotInstalled, got %v", err)
	}

	// A sibling that is not executable doesn't count; PATH is used instead.
	writeExe(t, gui, 0o644)
	onPath := func(string) (string, error) { return "/somewhere/promptpet-gui", nil }
	if got, err := window.FindBinary(exe, onPath); err != nil || got != "/somewhere/promptpet-gui" {
		t.Fatalf("non-executable sibling: got %q, %v", got, err)
	}

	// An executable sibling wins over PATH.
	if err := os.Chmod(gui, 0o755); err != nil {
		t.Fatal(err)
	}
	if got, err := window.FindBinary(exe, onPath); err != nil || got != gui {
		t.Fatalf("sibling: got %q, %v", got, err)
	}

	// Homebrew links bin/promptpet to the real install dir: look next to the target.
	linkDir := t.TempDir()
	link := filepath.Join(linkDir, "promptpet")
	if err := os.Symlink(exe, link); err != nil {
		t.Fatal(err)
	}
	if got, err := window.FindBinary(link, noPath); err != nil || got != gui {
		t.Fatalf("symlinked promptpet: got %q, %v", got, err)
	}
}

func env(vals map[string]string) func(string) string {
	return func(k string) string { return vals[k] }
}

func TestMainArgsAndHeadless(t *testing.T) {
	home := t.TempDir()
	never := func(window.Config) error { t.Error("window must not open"); return nil }
	var out, errOut bytes.Buffer

	if code := window.Main([]string{"--version"}, env(nil), &out, &errOut, "1.2.3", never); code != 0 || out.String() != "promptpet-gui 1.2.3\n" {
		t.Fatalf("--version: code %d, out %q", code, out.String())
	}
	if code := window.Main([]string{"bogus"}, env(nil), &out, &errOut, "dev", never); code != 2 {
		t.Fatalf("unknown arg: code %d", code)
	}
	errOut.Reset()
	headless := env(map[string]string{"PROMPTPET_HOME": home, "PROMPTPET_HEADLESS": "1"})
	if code := window.Main(nil, headless, &out, &errOut, "dev", never); code != 1 || !strings.Contains(errOut.String(), window.HeadlessMessage) {
		t.Fatalf("headless: code %d, stderr %q", code, errOut.String())
	}
}

func TestMainPassesPathsAndReportsErrors(t *testing.T) {
	home := t.TempDir()
	getenv := env(map[string]string{"PROMPTPET_HOME": home, "DISPLAY": ":0"})
	var out, errOut bytes.Buffer
	var got window.Config
	code := window.Main(nil, getenv, &out, &errOut, "dev", func(c window.Config) error { got = c; return nil })
	if code != 0 || got.StatePath != filepath.Join(home, "companion.json") || got.EventPath != filepath.Join(home, "events.jsonl") {
		t.Fatalf("code %d, config %+v", code, got)
	}

	code = window.Main(nil, getenv, &out, &errOut, "dev", func(window.Config) error { return window.ErrHeadless })
	if code != 1 || !strings.Contains(errOut.String(), window.HeadlessMessage) {
		t.Fatalf("ErrHeadless: code %d, stderr %q", code, errOut.String())
	}
	errOut.Reset()
	code = window.Main(nil, getenv, &out, &errOut, "dev", func(window.Config) error { return errors.New("no GPU") })
	if code != 1 || errOut.String() != "promptpet-gui: no GPU\n" {
		t.Fatalf("run error: code %d, stderr %q", code, errOut.String())
	}
}

// TestMainHoldsCompanionLock: while the window runs it owns companion.json, so
// `promptpet pet <name>` is refused (the window would overwrite it) and a second
// window can't start; once it closes, both work again.
func TestMainHoldsCompanionLock(t *testing.T) {
	home := t.TempDir()
	t.Setenv("PROMPTPET_HOME", home)
	t.Setenv("PROMPTPET_COMPANION_STATE_FILE", filepath.Join(home, "companion.json"))
	t.Setenv("PROMPTPET_HEADLESS", "")
	t.Setenv("DISPLAY", ":0")

	var out, errOut bytes.Buffer
	var petCode, secondCode int
	var petErr, secondErr string
	code := window.Main(nil, os.Getenv, &out, &errOut, "dev", func(window.Config) error {
		var pOut, pErr bytes.Buffer
		petCode = cli.Run([]string{"pet", "kochi"}, &pOut, &pErr)
		petErr = pErr.String()

		var sOut, sErr bytes.Buffer
		secondCode = window.Main(nil, os.Getenv, &sOut, &sErr, "dev", func(window.Config) error {
			t.Error("second window must not start")
			return nil
		})
		secondErr = sErr.String()
		return nil
	})
	if code != 0 {
		t.Fatalf("window exited %d: %s", code, errOut.String())
	}
	if petCode == 0 || !strings.Contains(petErr, "companion window") {
		t.Errorf("pet switch while window open: code %d, stderr %q", petCode, petErr)
	}
	if secondCode == 0 || !strings.Contains(secondErr, "already running") {
		t.Errorf("second window: code %d, stderr %q", secondCode, secondErr)
	}

	var pOut, pErr bytes.Buffer
	if code := cli.Run([]string{"pet", "kochi"}, &pOut, &pErr); code != 0 {
		t.Fatalf("pet switch after window closed: %d %s", code, pErr.String())
	}
}
