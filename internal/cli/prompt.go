package cli

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dpologdvinity/promptpet/internal/event"
	"github.com/dpologdvinity/promptpet/internal/pet"
	"github.com/dpologdvinity/promptpet/internal/repo"
	"github.com/dpologdvinity/promptpet/internal/store"
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

	var gitAction string
	var gitBranch string
	var commitCount int
	var foundRepo bool
	var repoWt, repoGd string

	if cwd, err := d.Getwd(); err == nil {
		if wt, gd, ok := repo.Find(cwd); ok {
			foundRepo = true
			repoWt = wt
			repoGd = gd
			info := repo.GetInfo(gd)
			gitBranch = info.Branch
			commitCount = info.CommitCount
			if sig, ok := repo.Sig(gd); ok && sig != s.Repo {
				if info.LastAction == "commit" {
					gitAction = "commit"
				}
			}
		}
	}

	if exit >= 0 {
		pet.ApplyExit(s, exit)
		ev := event.Event{
			Type:        "command_finish",
			ExitCode:    exit,
			Timestamp:   time.Now().Unix(),
			GitAction:   gitAction,
			GitBranch:   gitBranch,
			CommitCount: commitCount,
		}
		if d.Now != nil {
			ev.Timestamp = d.Now().Unix()
		}
		if d.Getenv != nil && d.Getenv("PROMPTPET_EVENT_FILE") != "" {
			_ = event.AppendEventToFile(d.Getenv("PROMPTPET_EVENT_FILE"), ev)
		} else if d.Getenv != nil && d.Getenv("PROMPTPET_HOME") != "" {
			_ = event.AppendEventToFile(filepath.Join(d.Getenv("PROMPTPET_HOME"), "events.jsonl"), ev)
		} else {
			_ = event.LogEvent(ev)
		}
	}
	mood := pet.MoodOf(s, pet.MoodInput{Now: now, PrevPrompt: prev, Exit: exit, Ate: n > 0})
	s.LastPrompt = now
	segment := pet.Segment(s, mood)
	if lockErr == nil {
		if foundRepo {
			if sig, ok := repo.Sig(repoGd); ok && sig != s.Repo {
				s.Repo = sig
				_ = d.Spawn(repoWt)
			}
		}
		_ = store.SaveState(dir, s)
	}
	fmt.Fprintln(d.Stdout, segment)
	return 0
}
