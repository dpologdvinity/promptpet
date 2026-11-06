package cli

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dpologdvinity/promptpet/internal/event"
	"github.com/dpologdvinity/promptpet/internal/pet"
	"github.com/dpologdvinity/promptpet/internal/repo"
	"github.com/dpologdvinity/promptpet/internal/store"
)

// parsePromptArgs reads the hook-supplied "--exit N" and "--duration-ms N"
// (or "--flag=N") arguments. Missing or invalid values yield -1 for exit
// (unknown) and 0 for duration (unknown, omitted from the event).
func parsePromptArgs(args []string) (exit int, durationMs int64) {
	exit = -1
	for i := 0; i < len(args); i++ {
		name, val, hasVal := strings.Cut(args[i], "=")
		if name != "--exit" && name != "--duration-ms" {
			continue
		}
		if !hasVal && i+1 < len(args) {
			i++
			val = args[i]
		}
		if name == "--exit" {
			if n, err := strconv.Atoi(val); err == nil {
				exit = n
			}
		} else if n, err := strconv.ParseInt(val, 10, 64); err == nil && n >= 0 {
			durationMs = n
		}
	}
	return exit, durationMs
}

// runPrompt implements `promptpet prompt`: it advances the pet, logs a
// command_finish event (with any git milestone that happened since the
// previous prompt) and prints the prompt segment.
func runPrompt(args []string, d Deps) int {
	if d.Getenv("PROMPTPET_DISABLE") != "" {
		return 0
	}
	exit, durationMs := parsePromptArgs(args)
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
			// The reflog signature changes after a git operation, so a milestone
			// fires once, on the first prompt after it happened. It also changes
			// when cd-ing into another repo, so the action must be newer than the
			// previous prompt or an old merge/branch there would replay.
			if sig, ok := repo.Sig(gd); ok && sig != s.Repo {
				info := repo.GetInfo(gd, now)
				gitBranch = info.Branch
				commitCount = info.CommitCount
				if prev.IsZero() || info.LastActionAt >= prev.Unix() {
					gitAction = info.LastAction
				}
			}
		}
	}

	if exit >= 0 {
		pet.ApplyExit(s, exit)
		ev := event.Event{
			Type:        "command_finish",
			ExitCode:    exit,
			DurationMs:  durationMs,
			Timestamp:   time.Now().Unix(),
			GitAction:   gitAction,
			GitBranch:   gitBranch,
			CommitCount: commitCount,
		}
		if d.Now != nil {
			ev.Timestamp = d.Now().Unix()
		}
		evFile := filepath.Join(dir, "events.jsonl")
		if d.Getenv != nil && d.Getenv("PROMPTPET_EVENT_FILE") != "" {
			evFile = d.Getenv("PROMPTPET_EVENT_FILE")
		}
		_ = event.AppendEventToFile(evFile, ev)
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
