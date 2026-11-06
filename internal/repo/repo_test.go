package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindDirAndSig(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".git", "logs"), 0o755)
	sub := filepath.Join(root, "a", "b")
	os.MkdirAll(sub, 0o755)
	wt, gd, ok := Find(sub)
	if !ok || wt != root || gd != filepath.Join(root, ".git") {
		t.Fatal(wt, gd, ok)
	}
	if _, ok := Sig(gd); ok {
		t.Fatal("no logs/HEAD yet")
	}
	os.WriteFile(filepath.Join(gd, "logs", "HEAD"), []byte("x\n"), 0o644)
	sig, ok := Sig(gd)
	if !ok || sig.GitDir != gd || sig.Size != 2 || sig.MtimeNs == 0 {
		t.Fatal(sig)
	}
}
func TestFindGitFile(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "master", ".git", "worktrees", "wt")
	os.MkdirAll(real, 0o755)
	wt := filepath.Join(root, "wt")
	os.MkdirAll(wt, 0o755)
	os.WriteFile(filepath.Join(wt, ".git"), []byte("gitdir: ../master/.git/worktrees/wt\n"), 0o644)
	gotWt, gd, ok := Find(wt)
	if !ok || gotWt != wt || gd != real {
		t.Fatal(gotWt, gd, ok)
	}
}
func TestFindNone(t *testing.T) {
	if _, _, ok := Find(t.TempDir()); ok {
		t.Skip("a parent of the temp dir is a git repo")
	}
	if _, _, ok := Find(filepath.Join(t.TempDir(), "does", "not", "exist")); ok {
		t.Fatal("found repo in missing dir")
	}
}

// reflogLine builds one HEAD/branch reflog line with the given unix time and message.
func reflogLine(ts int64, msg string) string {
	return fmt.Sprintf("1111111 2222222 Some User <u@example.com> %d +0100\t%s\n", ts, msg)
}

// fakeGitDir creates a .git dir with HEAD on branch and the given HEAD reflog body.
func fakeGitDir(t *testing.T, branch, headLog string) string {
	t.Helper()
	gitDir := filepath.Join(t.TempDir(), ".git")
	os.MkdirAll(filepath.Join(gitDir, "logs", "refs", "heads"), 0o755)
	os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/"+branch+"\n"), 0o644)
	os.WriteFile(filepath.Join(gitDir, "logs", "HEAD"), []byte(headLog), 0o644)
	return gitDir
}

// now is a fixed local afternoon; "today" is its calendar day in time.Local.
var now = time.Date(2026, 11, 6, 15, 0, 0, 0, time.Local)

func TestGetInfo(t *testing.T) {
	gitDir := fakeGitDir(t, "feature-branch",
		reflogLine(now.Add(-3*time.Hour).Unix(), "commit (initial): first commit")+
			reflogLine(now.Add(-2*time.Hour).Unix(), "checkout: moving from a to b")+
			reflogLine(now.Add(-time.Hour).Unix(), "commit: second commit"))
	info := GetInfo(gitDir, now)
	if info.Branch != "feature-branch" {
		t.Errorf("expected branch 'feature-branch', got %q", info.Branch)
	}
	if info.CommitCount != 2 {
		t.Errorf("expected 2 commits, got %d", info.CommitCount)
	}
	if info.LastAction != "commit" {
		t.Errorf("expected lastAction 'commit', got %q", info.LastAction)
	}
}

func TestGetInfoCountsOnlyTodaysCommits(t *testing.T) {
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	gitDir := fakeGitDir(t, "main",
		reflogLine(midnight.Add(-time.Second).Unix(), "commit: yesterday")+
			reflogLine(midnight.Unix(), "commit: at midnight")+
			reflogLine(now.Add(-time.Hour).Unix(), "commit (amend): reworded")+
			reflogLine(now.Add(-50*time.Minute).Unix(), "commit (merge): Merge branch 'f'")+
			reflogLine(now.Add(-40*time.Minute).Unix(), "rebase (finish): returning to refs/heads/main")+
			"garbage line without tab\n"+
			reflogLine(now.Add(-time.Minute).Unix(), "commit: latest"))
	if got := GetInfo(gitDir, now).CommitCount; got != 3 {
		t.Fatalf("today's commits = %d, want 3 (midnight, merge, latest)", got)
	}
	if got := GetInfo(gitDir, now.AddDate(0, 0, 1)).CommitCount; got != 0 {
		t.Fatalf("tomorrow's commits = %d, want 0", got)
	}
}

func TestGetInfoLastAction(t *testing.T) {
	ts := now.Add(-time.Minute).Unix()
	cases := []struct {
		name, msg, branchLog, want string
	}{
		{"commit", "commit: x", "", "commit"},
		{"initial", "commit (initial): x", "", "commit"},
		{"amend", "commit (amend): x", "", ""},
		{"merge commit", "commit (merge): Merge branch 'f'", "", "merge"},
		{"merge ort", "merge feat: Merge made by the 'ort' strategy.", "", "merge"},
		{"merge ff", "merge feat: Fast-forward", "", "merge"},
		{"pull merge", "pull: Merge made by the 'ort' strategy.", "", "merge"},
		{"pull ff", "pull: Fast-forward", "", ""},
		{"pull rebase", "pull --rebase (finish): returning to refs/heads/feat", "", ""},
		{"new branch", "checkout: moving from main to feat", reflogLine(ts, "branch: Created from main"), "branch"},
		{"old branch", "checkout: moving from main to feat", reflogLine(ts-3600, "branch: Created from main") + reflogLine(ts-60, "commit: y"), ""},
		{"no branch log", "checkout: moving from main to feat", "", ""},
		{"reset", "reset: moving to HEAD~1", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gitDir := fakeGitDir(t, "feat", reflogLine(ts-7200, "commit: older")+reflogLine(ts, c.msg))
			if c.branchLog != "" {
				os.WriteFile(filepath.Join(gitDir, "logs", "refs", "heads", "feat"), []byte(c.branchLog), 0o644)
			}
			if got := GetInfo(gitDir, now).LastAction; got != c.want {
				t.Fatalf("LastAction = %q, want %q", got, c.want)
			}
		})
	}
}

func TestGetInfoNewBranchInLinkedWorktree(t *testing.T) {
	ts := now.Unix()
	common := fakeGitDir(t, "main", "")
	os.MkdirAll(filepath.Join(common, "logs", "refs", "heads", "fix"), 0o755)
	os.WriteFile(filepath.Join(common, "logs", "refs", "heads", "fix", "bug"), []byte(reflogLine(ts, "branch: Created from HEAD")), 0o644)
	wtGit := filepath.Join(common, "worktrees", "wt")
	os.MkdirAll(filepath.Join(wtGit, "logs"), 0o755)
	os.WriteFile(filepath.Join(wtGit, "commondir"), []byte("../..\n"), 0o644)
	os.WriteFile(filepath.Join(wtGit, "HEAD"), []byte("ref: refs/heads/fix/bug\n"), 0o644)
	os.WriteFile(filepath.Join(wtGit, "logs", "HEAD"), []byte(reflogLine(ts, "checkout: moving from main to fix/bug")), 0o644)
	if info := GetInfo(wtGit, now); info.LastAction != "branch" || info.Branch != "fix/bug" {
		t.Fatalf("%+v", info)
	}
}
