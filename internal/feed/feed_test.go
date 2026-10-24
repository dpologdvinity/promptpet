package feed

import (
	"context"
	"fmt"
	"github.com/dpologdvinity/promptpet/internal/store"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func gitRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "global"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, args := range [][]string{{"init", "-q", "-b", "master", root}, {"-C", root, "config", "user.email", "me@example.com"}, {"-C", root, "config", "user.name", "Me"}, {"-C", root, "config", "commit.gpgsign", "false"}} {
		cmd := exec.Command("git", args...)
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s: %v", args, b, err)
		}
	}
	return root
}
func commit(t *testing.T, dir string, unix int64, email string) {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "commit", "-q", "--allow-empty", "-m", "c")
	stamp := fmt.Sprintf("@%d +0000", unix)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+stamp, "GIT_COMMITTER_DATE="+stamp, "GIT_AUTHOR_NAME=Me", "GIT_AUTHOR_EMAIL="+email, "GIT_COMMITTER_NAME=Me", "GIT_COMMITTER_EMAIL="+email)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("commit: %s: %v", b, err)
	}
}

func TestFeedNewRepoOnlyRecent(t *testing.T) {
	wt := gitRepo(t)
	state := t.TempDir()
	now := time.Unix(1_800_000_000, 0)
	commit(t, wt, now.Unix()-3600, "me@example.com")  // older than 10 min: not fed
	commit(t, wt, now.Unix()-60, "ME@example.com")    // case-insensitive match: fed
	commit(t, wt, now.Unix()-30, "other@example.com") // someone else: not fed
	if err := Run(context.Background(), state, wt, now); err != nil {
		t.Fatal(err)
	}
	if n, _, _ := store.ReadMeals(state, 0); n != 1 {
		t.Fatal(n)
	}
	f, _ := store.LoadFeed(state)
	for _, v := range f.Repos {
		if v != now.Unix()-60 {
			t.Fatal(v)
		}
	}
}
func TestFeedNoDoubleCountAndCap(t *testing.T) {
	wt := gitRepo(t)
	state := t.TempDir()
	now := time.Unix(1_800_000_000, 0)
	Run(context.Background(), state, wt, now) // no commits yet: no error, nothing fed
	for i := 0; i < 7; i++ {
		commit(t, wt, now.Unix()-300+int64(i), "me@example.com")
	}
	Run(context.Background(), state, wt, now)
	Run(context.Background(), state, wt, now) // same commits again
	if n, _, _ := store.ReadMeals(state, 0); n != 5 {
		t.Fatal(n)
	}
	b, _ := os.ReadFile(filepath.Join(state, "events.log"))
	if !strings.HasPrefix(string(b), fmt.Sprintf("meal %d ", now.Unix()-298)) {
		t.Fatalf("oldest-first of newest five: %s", b)
	}
}
func TestFeedNoEmail(t *testing.T) {
	wt := gitRepo(t)
	state := t.TempDir()
	exec.Command("git", "-C", wt, "config", "--unset", "user.email").Run()
	commit(t, wt, 1_800_000_000-10, "me@example.com")
	if err := Run(context.Background(), state, wt, time.Unix(1_800_000_000, 0)); err != nil {
		t.Fatal(err)
	}
	if n, _, _ := store.ReadMeals(state, 0); n != 0 {
		t.Fatal(n)
	}
}
