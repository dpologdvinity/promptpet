package cli

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/dpologdvinity/promptpet/internal/companion"
	"github.com/dpologdvinity/promptpet/internal/pet"
	"github.com/dpologdvinity/promptpet/internal/store"
	"github.com/dpologdvinity/promptpet/internal/window"
)

func statePath(d Deps) string { return window.StatePath(d.Getenv) }

func eventPath(d Deps) string { return window.EventPath(d.Getenv) }

// lockCompanion takes companion.lock next to the companion state file.
func lockCompanion(d Deps) (func(), error) { return window.Lock(d.Getenv) }

// runGUI starts the companion window. The window is a separate binary,
// promptpet-gui, so this one never links Ebitengine (see internal/window).
func runGUI(args []string, d Deps) int {
	detach := false
	if len(args) > 0 {
		if len(args) != 1 || (args[0] != "--detach" && args[0] != "-d") {
			return usageError(d, "usage: promptpet gui [--detach]")
		}
		detach = true
	}
	// Check here: a detached window has no terminal to report a missing display.
	if window.Headless(d.Getenv) {
		fmt.Fprintln(d.Stderr, window.HeadlessMessage)
		return 1
	}
	unlock, err := lockCompanion(d)
	if err != nil {
		fmt.Fprintln(d.Stderr, window.RunningMessage)
		return 1
	}
	unlock() // only a pre-check; promptpet-gui holds the lock for the window's whole life

	if detach {
		if err := d.Detach(); err != nil {
			return runtimeError(d, err)
		}
		fmt.Fprintln(d.Stdout, "promptpet companion started in the background (press Q in its window to quit).")
		return 0
	}
	if err := d.RunGUI(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
				fmt.Fprintf(d.Stderr, "promptpet: the companion window was killed (%v)\n", ws.Signal())
				return 128 + int(ws.Signal())
			}
			return exitErr.ExitCode() // promptpet-gui already explained on stderr
		}
		return runtimeError(d, err)
	}
	return 0
}

func runPet(args []string, d Deps) int {
	if len(args) == 0 {
		return withState(d, func(s *pet.State, _ bool) string {
			return strings.Join(pet.PetIt(s, d.Rand), "\n") + "\n"
		})
	}
	if len(args) == 1 {
		if args[0] == "list" {
			return runPetList(d)
		}
		return runPetSwitch(args[0], d)
	}
	return usageError(d, "usage: promptpet pet [name|list]")
}

func runPetList(d Deps) int {
	sp := statePath(d)
	dir := filepath.Dir(sp)
	unlock, lockErr := store.Lock(dir, "companion.lock", 200*time.Millisecond)
	if lockErr == nil {
		defer unlock()
	}
	s, err := companion.LoadState(sp)
	if err != nil {
		s = companion.NewDefaultState()
	}

	allPets := companion.GetAllPets()
	fmt.Fprintln(d.Stdout, "promptpet roster:")
	for _, p := range allPets {
		marker := "  "
		tag := ""
		if strings.EqualFold(string(p.ID), s.ActivePet) {
			marker = "* "
			tag = " (active)"
		}
		fmt.Fprintf(d.Stdout, "%s%-7s - %s: %s (favorite snack: %s)%s\n", marker, p.ID, p.Name, p.Species, p.FavoriteSnack, tag)
	}
	return 0
}

func runPetSwitch(name string, d Deps) int {
	var matched *companion.PetDef
	for _, p := range companion.GetAllPets() {
		if strings.EqualFold(string(p.ID), name) || strings.EqualFold(p.Name, name) {
			pCopy := p
			matched = &pCopy
			break
		}
	}
	if matched == nil {
		var valid []string
		for _, p := range companion.GetAllPets() {
			valid = append(valid, string(p.ID))
		}
		return usageError(d, fmt.Sprintf("unknown pet %q (choose %s)", name, strings.Join(valid, ", ")))
	}

	sp := statePath(d)
	// Writes must hold the lock; reads (list, games) may proceed without it since saves are atomic.
	unlock, lockErr := lockCompanion(d)
	if lockErr != nil {
		return runtimeError(d, errors.New("the companion window is open; switch pets there with Tab, or close it first"))
	}
	defer unlock()
	s, err := companion.LoadState(sp)
	if err != nil {
		s = companion.NewDefaultState()
	}
	s.ActivePet = string(matched.ID)
	if err := companion.SaveState(sp, s); err != nil {
		return runtimeError(d, err)
	}

	fmt.Fprintf(d.Stdout, "Switched active pet to %s (%s).\n", matched.Name, matched.Species)
	return 0
}

type gameInfo struct {
	ID       int
	Name     string
	Controls string
}

var allGameInfos = []gameInfo{
	{ID: 1, Name: "Snack Catch", Controls: "Left/Right arrows (or A/D)"},
	{ID: 2, Name: "Puddle Leap", Controls: "Spacebar (or Up arrow/W)"},
	{ID: 3, Name: "Look Away", Controls: "Left/Right arrows (or A/D)"},
	{ID: 4, Name: "Rhythm Dance", Controls: "Arrow keys (or WASD)"},
	{ID: 5, Name: "Memory Simon", Controls: "1, 2, 3"},
	{ID: 6, Name: "Whack-a-Bug", Controls: "1, 2, 3"},
	{ID: 7, Name: "Balloon Pop", Controls: "Up/Down arrows to aim, Spacebar to shoot"},
	{ID: 8, Name: "Word Snack", Controls: "Type floating word letters (A-Z)"},
}

func runGames(args []string, d Deps) int {
	sp := statePath(d)
	dir := filepath.Dir(sp)
	unlock, lockErr := store.Lock(dir, "companion.lock", 200*time.Millisecond)
	if lockErr == nil {
		defer unlock()
	}
	s, err := companion.LoadState(sp)
	if err != nil {
		s = companion.NewDefaultState()
	}

	fmt.Fprintln(d.Stdout, "promptpet mini-games (play in GUI with keys 1-8):")
	for _, gi := range allGameInfos {
		score := 0
		if s.HighScores != nil {
			score = s.HighScores[gi.Name]
		}
		fmt.Fprintf(d.Stdout, "  %d. %-13s - Controls: %-36s - High Score: %d\n", gi.ID, gi.Name, gi.Controls, score)
	}
	return 0
}

// companionStatus renders the companion section of `promptpet status`:
// active pet, hearts, casing and every mini-game's high score.
func companionStatus(cs *companion.State) string {
	casing := cs.Casing
	if casing == "" {
		casing = companion.Casings[0]
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\nCompanion: %s (Hearts: %d, Casing: %s)\n", cs.ActivePet, cs.FriendshipHearts, casing)
	b.WriteString("High scores:\n")
	for _, gi := range allGameInfos {
		fmt.Fprintf(&b, "  %-13s %d\n", gi.Name, cs.HighScores[gi.Name])
	}
	return b.String()
}
