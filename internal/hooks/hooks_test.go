package hooks

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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

// argsBin is a fake promptpet that prints all of its arguments.
const argsBin = "#!/bin/sh\necho \"ARGS:$*\"\n"

var durationRe = regexp.MustCompile(`^ARGS:prompt --exit 7 --duration-ms=([0-9]+)$`)

// checkDuration asserts seg reports exit 7 and a duration of at least minMs.
func checkDuration(t *testing.T, seg string, minMs int) {
	t.Helper()
	m := durationRe.FindStringSubmatch(seg)
	if m == nil {
		t.Fatalf("segment %q lacks --duration-ms", seg)
	}
	if ms, _ := strconv.Atoi(m[1]); ms < minMs || ms > 60000 {
		t.Fatalf("duration %dms not in [%d, 60000]", ms, minMs)
	}
}

func TestBashHookDuration(t *testing.T) {
	hook, _ := Script("bash")
	f := filepath.Join(t.TempDir(), "h.bash")
	os.WriteFile(f, []byte(hook), 0o644)
	bin := fakeBin(t, argsBin)
	// ${PS0@P} performs the expansion an interactive bash does before running a command.
	out, errOut := runShell(t, "bash", []string{"--norc", "-c", `source "$1"; source "$1"; : "${PS0@P}"; sleep 0.2; (exit 7); _promptpet_precmd; echo "rc=$? seg=$PROMPTPET_SEGMENT"; (exit 7); _promptpet_precmd; echo "again seg=$PROMPTPET_SEGMENT"; echo "ps0=[${PS0@P}]"`, "x", f}, bin+":/usr/bin:/bin")
	lines := strings.Split(out, "\n")
	if len(lines) != 4 || !strings.HasPrefix(lines[0], "rc=7 seg=") || errOut != "" {
		t.Fatalf("out=%q err=%q", out, errOut)
	}
	checkDuration(t, strings.TrimPrefix(lines[0], "rc=7 seg="), 150)
	// No command ran since the last prompt (empty Enter): duration omitted.
	if lines[1] != "again seg=ARGS:prompt --exit 7" || lines[2] != "ps0=[]" {
		t.Fatalf("out=%q", out)
	}
}

// TestBashHookDurationInteractive drives a real interactive bash (under
// set -u) so PS0 and PROMPT_COMMAND fire as they do for a user.
func TestBashHookDurationInteractive(t *testing.T) {
	hook, _ := Script("bash")
	dir := t.TempDir()
	f := filepath.Join(dir, "h.bash")
	os.WriteFile(f, []byte(hook), 0o644)
	logf := filepath.Join(dir, "log")
	bin := fakeBin(t, "#!/bin/sh\necho \"ARGS:$*\" >> "+logf+"\n")
	cmd := exec.Command("bash", "--norc", "--noprofile", "-i")
	cmd.Env = append(os.Environ(), "PATH="+bin+":/usr/bin:/bin", "HOME="+dir)
	cmd.Stdin = strings.NewReader("set -u\nsource " + f + "\nsleep 0.2\n\nfalse\n")
	stderr, _ := cmd.CombinedOutput()
	log, _ := os.ReadFile(logf)
	lines := strings.Split(strings.TrimSpace(string(log)), "\n")
	// prompts after: source (PS0 not yet installed), sleep, empty line (untimed), false
	if strings.Contains(string(stderr), "unbound") || len(lines) != 4 {
		t.Fatalf("log=%q stderr=%q", log, stderr)
	}
	if !strings.HasPrefix(lines[1], "ARGS:prompt --exit 0 --duration-ms=") || lines[2] != "ARGS:prompt --exit 0" || !strings.HasPrefix(lines[3], "ARGS:prompt --exit 1 --duration-ms=") {
		t.Fatalf("log=%q", log)
	}
	if ms, _ := strconv.Atoi(strings.TrimPrefix(lines[1], "ARGS:prompt --exit 0 --duration-ms=")); ms < 150 {
		t.Fatalf("sleep 0.2 measured %dms", ms)
	}
}

func TestZshHookDuration(t *testing.T) {
	if _, err := exec.LookPath("zsh"); err != nil {
		t.Skip("zsh not installed")
	}
	hook, _ := Script("zsh")
	f := filepath.Join(t.TempDir(), "h.zsh")
	os.WriteFile(f, []byte(hook), 0o644)
	bin := fakeBin(t, argsBin)
	out, errOut := runShell(t, "zsh", []string{"-f", "-c", `source "$1"; source "$1"; for f in $preexec_functions; do $f; done; sleep 0.2; (exit 7); _promptpet_precmd; echo "rc=$? seg=$PROMPTPET_SEGMENT"; (exit 7); _promptpet_precmd; echo "again seg=$PROMPTPET_SEGMENT"; echo "hooks=$preexec_functions"`, "x", f}, bin+":/usr/bin:/bin")
	lines := strings.Split(out, "\n")
	if len(lines) != 4 || !strings.HasPrefix(lines[0], "rc=7 seg=") || errOut != "" {
		t.Fatalf("out=%q err=%q", out, errOut)
	}
	checkDuration(t, strings.TrimPrefix(lines[0], "rc=7 seg="), 150)
	if lines[1] != "again seg=ARGS:prompt --exit 7" || lines[2] != "hooks=_promptpet_preexec" {
		t.Fatalf("out=%q", out)
	}
}

func TestFishHookDuration(t *testing.T) {
	if _, err := exec.LookPath("fish"); err != nil {
		t.Skip("fish not installed")
	}
	hook, _ := Script("fish")
	f := filepath.Join(t.TempDir(), "h.fish")
	os.WriteFile(f, []byte(hook), 0o644)
	bin := fakeBin(t, argsBin)
	out, errOut := runShell(t, "fish", []string{"--no-config", "-c", `function fish_prompt; end; source $argv[1]; set CMD_DURATION 250; false; fish_prompt; echo; set -e CMD_DURATION; false; fish_prompt`, f}, bin+":/usr/bin:/bin")
	if out != "ARGS:prompt --exit 1 --duration-ms=250 \nARGS:prompt --exit 1 " || errOut != "" {
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

	// Test execution of completion scripts in subshells
	bashComp, _ := Completion("bash")
	fb := filepath.Join(t.TempDir(), "comp.bash")
	os.WriteFile(fb, []byte(bashComp), 0o644)
	out, errOut := runShell(t, "bash", []string{"--norc", "-c", `source "$1"; type _promptpet_completions`, "x", fb}, "/usr/bin:/bin")
	if !strings.Contains(out, "_promptpet_completions is a function") || errOut != "" {
		t.Fatalf("bash completion failed: out=%q err=%q", out, errOut)
	}

	if _, err := exec.LookPath("zsh"); err == nil {
		zshComp, _ := Completion("zsh")
		fz := filepath.Join(t.TempDir(), "comp.zsh")
		os.WriteFile(fz, []byte(zshComp), 0o644)
		out, errOut = runShell(t, "zsh", []string{"-f", "-c", `autoload -Uz compinit && compinit -D; source "$1"; which _promptpet`, "x", fz}, "/usr/bin:/bin")
		if !strings.Contains(out, "_promptpet ()") || errOut != "" {
			t.Fatalf("zsh completion failed: out=%q err=%q", out, errOut)
		}
	}
}
