package feed

import (
	"context"
	"github.com/dpologdvinity/promptpet/internal/pet"
	"github.com/dpologdvinity/promptpet/internal/store"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func Run(parent context.Context, dir, worktree string, now time.Time) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	unlock, err := store.Lock(dir, "feed.lock", 5*time.Second)
	if err != nil {
		return err
	}
	defer unlock()
	git := func(args ...string) (string, error) {
		cmd := exec.CommandContext(ctx, "git", append([]string{"-C", worktree}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0", "GIT_TERMINAL_PROMPT=0", "LC_ALL=C")
		b, e := cmd.Output()
		return strings.TrimSpace(string(b)), e
	}
	key, err := git("rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil || key == "" {
		return err
	}
	key = filepath.Clean(key)
	email, err := git("config", "--get", "user.email")
	if err != nil || email == "" {
		return nil
	}
	log, err := git("log", "-n", "50", "--no-merges", "--format=%H%x09%at%x09%ae", "HEAD")
	if err != nil {
		// An unborn HEAD has nothing to feed.
		return nil
	}
	f, err := store.LoadFeed(dir)
	if err != nil {
		return err
	}
	last, ok := f.Repos[key]
	if !ok {
		last = now.Unix() - 600
	}
	matches := []store.MealLine{}
	maxSeen := last
	for _, line := range strings.Split(log, "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) != 3 || len(parts[0]) < 7 || !strings.EqualFold(parts[2], email) {
			continue
		}
		unix, e := strconv.ParseInt(parts[1], 10, 64)
		if e != nil || unix <= last {
			continue
		}
		matches = append(matches, store.MealLine{Unix: unix, SHA: parts[0][:7]})
		if unix > maxSeen {
			maxSeen = unix
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Unix == matches[j].Unix {
			return matches[i].SHA < matches[j].SHA
		}
		return matches[i].Unix < matches[j].Unix
	})
	if len(matches) > pet.MaxMealsPerCheck {
		matches = matches[len(matches)-pet.MaxMealsPerCheck:]
	}
	if err := store.AppendMeals(dir, matches); err != nil {
		return err
	}
	f.Repos[key] = maxSeen
	return store.SaveFeed(dir, f)
}
