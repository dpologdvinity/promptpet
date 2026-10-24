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
