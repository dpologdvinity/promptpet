package repo

import (
	"os"
	"path/filepath"
	"testing"
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

func TestGetInfo(t *testing.T) {
	root := t.TempDir()
	gitDir := filepath.Join(root, ".git")
	os.MkdirAll(filepath.Join(gitDir, "logs"), 0o755)

	// Set branch
	os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/feature-branch\n"), 0o644)

	// Set logs/HEAD with commits and checkout
	logContent := "000000 111111 User <u@example.com> 1000 +0000\tcommit (initial): first commit\n" +
		"111111 222222 User <u@example.com> 2000 +0000\tcheckout: moving from a to b\n" +
		"222222 333333 User <u@example.com> 3000 +0000\tcommit: second commit\n"
	os.WriteFile(filepath.Join(gitDir, "logs", "HEAD"), []byte(logContent), 0o644)

	info := GetInfo(gitDir)
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
