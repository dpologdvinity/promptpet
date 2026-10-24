package hooks

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func fakeBin(t *testing.T, script string) string {
	t.Helper()
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "promptpet"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return d
}
func runShell(t *testing.T, shell string, args []string, path string) (string, string) {
	t.Helper()
	cmd := exec.Command(shell, args...)
	cmd.Env = append(os.Environ(), "PATH="+path, "HOME="+t.TempDir())
	out, err := cmd.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return string(out), string(e.Stderr)
		}
		t.Fatal(err)
	}
	return string(out), ""
}
func TestBashHook(t *testing.T) {
	hook, _ := Script("bash")
	f := filepath.Join(t.TempDir(), "h.bash")
	os.WriteFile(f, []byte(hook), 0o644)
	bin := fakeBin(t, "#!/bin/sh\necho \"SEG:$3\"\n")
	out, errOut := runShell(t, "bash", []string{"--norc", "-c", `PS1='$ '; source "$1"; source "$1"; (exit 7); _promptpet_precmd; echo "rc=$? seg=$PROMPTPET_SEGMENT pc=$PROMPT_COMMAND"; echo "$PS1"`, "x", f}, bin+":/usr/bin:/bin")
	want := "rc=7 seg=SEG:7 pc=_promptpet_precmd\n${PROMPTPET_SEGMENT:+$PROMPTPET_SEGMENT }$ \n"
	if out != want || errOut != "" {
		t.Fatalf("out=%q err=%q", out, errOut)
	}
}
func TestBashHookMissingBinary(t *testing.T) {
	hook, _ := Script("bash")
	f := filepath.Join(t.TempDir(), "h.bash")
	os.WriteFile(f, []byte(hook), 0o644)
	out, errOut := runShell(t, "bash", []string{"--norc", "-c", `source "$1"; (exit 3); _promptpet_precmd; echo "rc=$? seg=[$PROMPTPET_SEGMENT]"`, "x", f}, t.TempDir())
	if out != "rc=3 seg=[]\n" || errOut != "" {
		t.Fatalf("out=%q err=%q", out, errOut)
	}
}
func TestZshHook(t *testing.T) {
	if _, err := exec.LookPath("zsh"); err != nil {
		t.Skip("zsh not installed")
	}
	hook, _ := Script("zsh")
	f := filepath.Join(t.TempDir(), "h.zsh")
	os.WriteFile(f, []byte(hook), 0o644)
	bin := fakeBin(t, "#!/bin/sh\necho \"SEG:$3\"\n")
	out, errOut := runShell(t, "zsh", []string{"-f", "-c", `PROMPT='%# '; source "$1"; source "$1"; (exit 7); _promptpet_precmd; echo "rc=$? seg=$PROMPTPET_SEGMENT"; echo "$PROMPT"`, "x", f}, bin+":/usr/bin:/bin")
	if !strings.Contains(out, "rc=7 seg=SEG:7\n") || !strings.Contains(out, "${PROMPTPET_SEGMENT:+$PROMPTPET_SEGMENT }%# \n") || errOut != "" {
		t.Fatalf("out=%q err=%q", out, errOut)
	}
}
func TestFishHook(t *testing.T) {
	if _, err := exec.LookPath("fish"); err != nil {
		t.Skip("fish not installed")
	}
	hook, _ := Script("fish")
	f := filepath.Join(t.TempDir(), "h.fish")
	os.WriteFile(f, []byte(hook), 0o644)
	bin := fakeBin(t, "#!/bin/sh\necho \"SEG:$3\"\n")
	out, errOut := runShell(t, "fish", []string{"--no-config", "-c", `function fish_prompt; echo -n "orig:$status"; end; source $argv[1]; source $argv[1]; false; fish_prompt`, f}, bin+":/usr/bin:/bin")
	if out != "SEG:1 orig:1" || errOut != "" {
		t.Fatalf("out=%q err=%q", out, errOut)
	}
}
func TestUnknownShell(t *testing.T) {
	if s, ok := Script("tcsh"); ok || s != "" {
		t.Fatal(s, ok)
	}
}

func TestCompletion(t *testing.T) {
	for _, sh := range []string{"zsh", "bash", "fish"} {
		s, ok := Completion(sh)
		if !ok || len(s) == 0 {
			t.Fatalf("shell %s completion failed", sh)
		}
	}
	if s, ok := Completion("tcsh"); ok || s != "" {
		t.Fatal(s, ok)
	}
}
