package cli

import (
	"fmt"
	"github.com/dpologdvinity/promptpet/internal/pet"
	"github.com/dpologdvinity/promptpet/internal/repo"
	"github.com/dpologdvinity/promptpet/internal/store"
	"strconv"
	"time"
)

func runPrompt(args []string, d Deps) int {
	if d.Getenv("PROMPTPET_DISABLE") != "" {
		return 0
	}
	exit := -1
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--exit" {
			if n, err := strconv.Atoi(args[i+1]); err == nil {
				exit = n
			}
			break
		}
	}
	now := d.Now()
	dir, err := store.Dir(d.Getenv)
	if err != nil {
		return 0
	}
	unlock, lockErr := store.Lock(dir, "state.lock", 5*time.Millisecond)
	if lockErr == nil {
		defer unlock()
	}
	var s *pet.State
	if lockErr == nil {
		s, _, err = store.LoadState(dir, now)
	} else {
		s, err = store.LoadStateReadOnly(dir, now)
	}
	if err != nil {
		return 0
	}
	prev := s.LastPrompt
	pet.Tick(s, now)
	n, off, err := store.ReadMeals(dir, s.EventOffset)
	if err != nil {
		return 0
	}
	for i := 0; i < n; i++ {
		pet.Meal(s, d.Rand)
	}
	s.EventOffset = off
	if exit >= 0 {
		pet.ApplyExit(s, exit)
	}
	mood := pet.MoodOf(s, pet.MoodInput{Now: now, PrevPrompt: prev, Exit: exit, Ate: n > 0})
	s.LastPrompt = now
	segment := pet.Segment(s, mood)
	if lockErr == nil {
		if cwd, err := d.Getwd(); err == nil {
			if wt, gd, ok := repo.Find(cwd); ok {
				if sig, ok := repo.Sig(gd); ok && sig != s.Repo {
					s.Repo = sig
					_ = d.Spawn(wt)
				}
			}
		}
		_ = store.SaveState(dir, s)
	}
	fmt.Fprintln(d.Stdout, segment)
	return 0
}
