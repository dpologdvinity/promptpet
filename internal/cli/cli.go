package cli

import (
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type Deps struct {
	Version string
	Now     func() time.Time
	Rand    *rand.Rand
	Getenv  func(string) string
	Getwd   func() (string, error)
	Spawn   func(string) error
	Stdout  io.Writer
	Stderr  io.Writer
}

func RealDeps(version string) Deps {
	seed := uint64(time.Now().UnixNano())
	return Deps{version, time.Now, rand.New(rand.NewPCG(seed, seed+1)), os.Getenv, os.Getwd, realSpawn, os.Stdout, os.Stderr}
}
func realSpawn(worktree string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "_feed", worktree)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err = cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

const usage = `promptpet - a tiny pet that lives in your shell prompt

usage:
  promptpet                  show your pet
  promptpet init <shell>     print the shell hook (zsh, bash, fish)
  promptpet pet              pet your pet
  promptpet play             play with your pet
  promptpet name <name>      rename your pet
  promptpet species <kind>   change species (cat, dragon, blob)
  promptpet prompt           print the prompt segment (used by the hook)
  promptpet --version        print the version

Your pet eats when you make git commits.
Setup (zsh):  add  eval "$(promptpet init zsh)"  at the end of ~/.zshrc
Setup (bash): add  eval "$(promptpet init bash)"  at the end of ~/.bashrc
Setup (fish): add  promptpet init fish | source  to ~/.config/fish/config.fish
`

func Run(args []string, d Deps) int {
	if len(args) > 0 && args[0] == "prompt" {
		return runPrompt(args[1:], d)
	}
	if len(args) == 1 {
		switch args[0] {
		case "--version", "-V":
			fmt.Fprintf(d.Stdout, "promptpet %s\n", d.Version)
			return 0
		case "--help", "-h", "help":
			fmt.Fprint(d.Stdout, usage)
			return 0
		}
	}
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	}
	fmt.Fprintf(d.Stderr, "promptpet: unknown command %q (see promptpet --help)\n", cmd)
	return 2
}
