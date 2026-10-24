package cli

import (
	"context"
	"fmt"
	"github.com/dpologdvinity/promptpet/internal/feed"
	"github.com/dpologdvinity/promptpet/internal/hooks"
	"github.com/dpologdvinity/promptpet/internal/pet"
	"github.com/dpologdvinity/promptpet/internal/store"
	"io"
	"math/rand/v2"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type Deps struct {
	Version string
	Now     func() time.Time
	Rand    *rand.Rand
	Getenv  func(string) string
	Getwd   func() (string, error)
	Spawn   func(string) error
	Stdout  io.Writer
	Stderr  io.Writer
}

func RealDeps(version string) Deps {
	seed := uint64(time.Now().UnixNano())
	return Deps{version, time.Now, rand.New(rand.NewPCG(seed, seed+1)), os.Getenv, os.Getwd, realSpawn, os.Stdout, os.Stderr}
}
func realSpawn(worktree string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "_feed", worktree)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err = cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

const usage = `promptpet - a tiny pet that lives in your shell prompt

usage:
  promptpet                  show your pet
  promptpet init <shell>     print the shell hook (zsh, bash, fish)
  promptpet pet              pet your pet
  promptpet play             play with your pet
  promptpet name <name>      rename your pet
  promptpet species <kind>   change species (cat, dragon, blob)
  promptpet prompt           print the prompt segment (used by the hook)
  promptpet --version        print the version

Your pet eats when you make git commits.
Setup (zsh):  add  eval "$(promptpet init zsh)"  at the end of ~/.zshrc
Setup (bash): add  eval "$(promptpet init bash)"  at the end of ~/.bashrc
Setup (fish): add  promptpet init fish | source  to ~/.config/fish/config.fish
`

func usageError(d Deps, message string) int { fmt.Fprintln(d.Stderr, "promptpet: "+message); return 2 }
func runtimeError(d Deps, err error) int    { fmt.Fprintln(d.Stderr, "promptpet: "+err.Error()); return 1 }
func withState(d Deps, fn func(*pet.State, bool) string) int {
	now := d.Now()
	dir, err := store.Dir(d.Getenv)
	if err != nil {
		return runtimeError(d, err)
	}
	unlock, err := store.Lock(dir, "state.lock", 2*time.Second)
	if err != nil {
		return runtimeError(d, err)
	}
	defer unlock()
	s, moved, err := store.LoadState(dir, now)
	if err != nil {
		return runtimeError(d, err)
	}
	if moved != "" {
		fmt.Fprintf(d.Stderr, "promptpet: your pet's state file was unreadable; moved it to %s and started fresh\n", moved)
	}
	pet.Tick(s, now)
	n, off, err := store.ReadMeals(dir, s.EventOffset)
	if err != nil {
		return runtimeError(d, err)
	}
	for i := 0; i < n; i++ {
		pet.Meal(s, d.Rand)
	}
	s.EventOffset = off
	output := fn(s, n > 0)
	if err := store.SaveState(dir, s); err != nil {
		return runtimeError(d, err)
	}
	fmt.Fprint(d.Stdout, output)
	return 0
}
func Run(args []string, d Deps) int {
	if len(args) == 0 {
		return withState(d, func(s *pet.State, ate bool) string {
			mood := pet.MoodOf(s, pet.MoodInput{Now: d.Now(), PrevPrompt: s.LastPrompt, Exit: -1, Ate: ate})
			card := pet.Card(s, mood, d.Now())
			s.Message = ""
			return card
		})
	}
	switch args[0] {
	case "prompt":
		return runPrompt(args[1:], d)
	case "_feed":
		if len(args) == 2 {
			if dir, err := store.Dir(d.Getenv); err == nil {
				_ = feed.Run(context.Background(), dir, args[1], d.Now())
			}
		}
		return 0
	case "--version", "-V":
		if len(args) == 1 {
			fmt.Fprintf(d.Stdout, "promptpet %s\n", d.Version)
			return 0
		}
	case "--help", "-h", "help":
		if len(args) == 1 {
			fmt.Fprint(d.Stdout, usage)
			return 0
		}
	case "init":
		if len(args) != 2 {
			return usageError(d, "init needs a shell: zsh, bash or fish")
		}
		script, ok := hooks.Script(args[1])
		if !ok {
			return usageError(d, "init needs a shell: zsh, bash or fish")
		}
		fmt.Fprint(d.Stdout, script)
		return 0
	case "name":
		if len(args) != 2 {
			return usageError(d, "usage: promptpet name <name>")
		}
		name, ok := pet.ValidName(args[1])
		if !ok {
			return usageError(d, "names can use letters, digits, spaces and - _ ' . (1-20 characters)")
		}
		return withState(d, func(s *pet.State, _ bool) string {
			s.Name = name
			return fmt.Sprintf("Your pet is now called %s.\n", name)
		})
	case "species":
		if len(args) != 2 || !pet.ValidSpecies(args[1]) {
			kind := ""
			if len(args) > 1 {
				kind = args[1]
			}
			return usageError(d, fmt.Sprintf("unknown species %q (choose cat, dragon or blob)", kind))
		}
		return withState(d, func(s *pet.State, _ bool) string {
			s.Species = args[1]
			return fmt.Sprintf("%s is now a %s.\n", s.Name, s.Species)
		})
	case "pet":
		if len(args) != 1 {
			return usageError(d, "usage: promptpet pet")
		}
		return withState(d, func(s *pet.State, _ bool) string { return strings.Join(pet.PetIt(s, d.Rand), "\n") + "\n" })
	case "play":
		if len(args) != 1 {
			return usageError(d, "usage: promptpet play")
		}
		return withState(d, func(s *pet.State, _ bool) string { return strings.Join(pet.Play(s, d.Rand), "\n") + "\n" })
	}
	if len(args) == 1 {
		return usageError(d, fmt.Sprintf("unknown command %q (see promptpet --help)", args[0]))
	}
	return usageError(d, fmt.Sprintf("unknown command %q (see promptpet --help)", args[0]))
}
