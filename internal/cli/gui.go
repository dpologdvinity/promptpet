package cli

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dpologdvinity/promptpet/internal/companion"
	"github.com/dpologdvinity/promptpet/internal/event"
	"github.com/dpologdvinity/promptpet/internal/gui"
	"github.com/dpologdvinity/promptpet/internal/store"
)

func statePath(d Deps) string {
	if d.Getenv != nil {
		if f := d.Getenv("PROMPTPET_STATE_FILE"); f != "" {
			return f
		}
		if dir, err := store.Dir(d.Getenv); err == nil {
			return filepath.Join(dir, "state.json")
		}
	}
	return companion.DefaultStatePath()
}

func eventPath(d Deps) string {
	if d.Getenv != nil {
		if f := d.Getenv("PROMPTPET_EVENT_FILE"); f != "" {
			return f
		}
		if dir, err := store.Dir(d.Getenv); err == nil {
			return filepath.Join(dir, "events.jsonl")
		}
	}
	return event.GetEventFilePath()
}

func runGUI(args []string, d Deps) int {
	cfg := gui.Config{
		StatePath: statePath(d),
		EventPath: eventPath(d),
	}

	runFn := gui.Run
	if d.RunGUI != nil {
		runFn = d.RunGUI
	}

	if err := runFn(cfg); err != nil {
		if errors.Is(err, gui.ErrHeadless) {
			fmt.Fprintln(d.Stderr, "promptpet gui requires a graphical display (X11/Wayland/WSLg)")
		} else {
			fmt.Fprintf(d.Stderr, "promptpet: %v\n", err)
		}
		return 1
	}
	return 0
}

func runPet(args []string, d Deps) int {
	if len(args) == 0 {
		return runPetList(d)
	}
	if len(args) == 1 {
		return runPetSwitch(args[0], d)
	}
	return usageError(d, "usage: promptpet pet [name]")
}

func runPetList(d Deps) int {
	sp := statePath(d)
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
