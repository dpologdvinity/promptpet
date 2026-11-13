// Package window holds everything about the companion window that does not need
// Ebitengine: where its files live, the headless check, the companion.lock it
// owns, how promptpet finds the separate promptpet-gui binary, and that binary's
// main loop around the renderer.
//
// It exists so that the promptpet binary, which every shell prompt runs, never
// links Ebitengine. Go runs the init code of every linked package on every start,
// and Ebitengine's takes tens of milliseconds; only cmd/promptpet-gui imports
// internal/gui.
package window

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/dpologdvinity/promptpet/internal/companion"
	"github.com/dpologdvinity/promptpet/internal/event"
	"github.com/dpologdvinity/promptpet/internal/store"
)

// BinaryName is the executable that renders the companion window.
const BinaryName = "promptpet-gui"

// HeadlessMessage is printed when there is no display to open a window on.
const HeadlessMessage = "promptpet gui requires a graphical display (X11/Wayland/WSLg)"

// RunningMessage is printed when another window already holds companion.lock.
const RunningMessage = "promptpet: the companion window is already running"

// ErrHeadless is returned when attempting to run the GUI in an environment without a display server.
var ErrHeadless = errors.New("gui: no display available (headless environment)")

// ErrNotInstalled means promptpet-gui is neither next to promptpet nor on PATH.
var ErrNotInstalled = errors.New("promptpet gui needs the promptpet-gui binary next to promptpet or on your PATH " +
	"(go install github.com/dpologdvinity/promptpet/cmd/promptpet-gui@latest)")

// Config holds initialization paths for companion state and event IPC.
type Config struct {
	StatePath          string
	EventPath          string
	InitialEventOffset *int64
}

// StatePath returns companion.json: PROMPTPET_COMPANION_STATE_FILE, else the state dir.
func StatePath(getenv func(string) string) string {
	if f := getenv("PROMPTPET_COMPANION_STATE_FILE"); f != "" {
		return f
	}
	if dir, err := store.Dir(getenv); err == nil {
		return filepath.Join(dir, "companion.json")
	}
	return companion.DefaultStatePath()
}

// EventPath returns events.jsonl: PROMPTPET_EVENT_FILE, else the state dir.
func EventPath(getenv func(string) string) string {
	if f := getenv("PROMPTPET_EVENT_FILE"); f != "" {
		return f
	}
	if dir, err := store.Dir(getenv); err == nil {
		return filepath.Join(dir, "events.jsonl")
	}
	return event.GetEventFilePath()
}

// Lock takes companion.lock next to the companion state file.
func Lock(getenv func(string) string) (func(), error) {
	return store.Lock(filepath.Dir(StatePath(getenv)), "companion.lock", 200*time.Millisecond)
}

// Headless reports whether there is no display: PROMPTPET_HEADLESS=1, or Linux
// with neither DISPLAY nor WAYLAND_DISPLAY set.
func Headless(getenv func(string) string) bool {
	if getenv("PROMPTPET_HEADLESS") == "1" {
		return true
	}
	if runtime.GOOS == "linux" {
		return getenv("DISPLAY") == "" && getenv("WAYLAND_DISPLAY") == ""
	}
	return false
}

// FindBinary locates promptpet-gui for the promptpet executable at exe: first in
// the same directory (release archives, packages, Homebrew and go install all put
// both there), then on PATH. exe's symlinks are resolved first, so a Homebrew
// bin/ symlink finds the sibling in the real install directory.
func FindBinary(exe string, lookPath func(string) (string, error)) (string, error) {
	if exe != "" {
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			exe = real
		}
		sibling := filepath.Join(filepath.Dir(exe), BinaryName)
		if info, err := os.Stat(sibling); err == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
			return sibling, nil
		}
	}
	if p, err := lookPath(BinaryName); err == nil {
		return p, nil
	}
	return "", ErrNotInstalled
}

// RealFind finds promptpet-gui for the running promptpet executable.
func RealFind() (string, error) {
	exe, _ := os.Executable()
	return FindBinary(exe, exec.LookPath)
}

// Main is the body of promptpet-gui. The window keeps companion.json in memory
// and saves it, so it holds companion.lock for its whole life: a second window
// refuses to start and `promptpet pet <name>` refuses instead of being silently
// overwritten.
func Main(args []string, getenv func(string) string, stdout, stderr io.Writer, version string, run func(Config) error) int {
	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v") {
		fmt.Fprintf(stdout, "%s %s\n", BinaryName, version)
		return 0
	}
	if len(args) != 0 {
		fmt.Fprintf(stderr, "usage: %s (or run: promptpet gui [--detach])\n", BinaryName)
		return 2
	}
	if Headless(getenv) {
		fmt.Fprintln(stderr, HeadlessMessage)
		return 1
	}
	unlock, err := Lock(getenv)
	if err != nil {
		fmt.Fprintln(stderr, RunningMessage)
		return 1
	}
	defer unlock()

	if err := run(Config{StatePath: StatePath(getenv), EventPath: EventPath(getenv)}); err != nil {
		if errors.Is(err, ErrHeadless) {
			fmt.Fprintln(stderr, HeadlessMessage)
		} else {
			fmt.Fprintf(stderr, "%s: %v\n", BinaryName, err)
		}
		return 1
	}
	return 0
}
