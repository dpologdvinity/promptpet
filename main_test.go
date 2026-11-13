package main

import (
	"os/exec"
	"strings"
	"testing"
)

// deps lists every package linked into the given main package.
func deps(t *testing.T, pkg string) []string {
	t.Helper()
	out, err := exec.Command("go", "list", "-deps", pkg).Output()
	if err != nil {
		t.Fatalf("go list -deps %s: %v", pkg, err)
	}
	return strings.Fields(string(out))
}

// TestPromptBinaryDoesNotLinkEbitengine guards the prompt budget. Go runs the
// init code of every linked package on every start, and Ebitengine's init takes
// tens of milliseconds, so linking it into promptpet made every prompt slow even
// though `promptpet prompt` never opens a window. The window lives in its own
// binary (cmd/promptpet-gui) instead.
func TestPromptBinaryDoesNotLinkEbitengine(t *testing.T) {
	for _, p := range deps(t, ".") {
		if strings.Contains(p, "ebiten") || strings.HasSuffix(p, "/internal/gui") {
			t.Errorf("promptpet links %s; GUI code belongs in cmd/promptpet-gui", p)
		}
	}
}

func TestGUIBinaryLinksEbitengine(t *testing.T) {
	found := false
	for _, p := range deps(t, "./cmd/promptpet-gui") {
		if p == "github.com/hajimehoshi/ebiten/v2" {
			found = true
		}
	}
	if !found {
		t.Fatal("cmd/promptpet-gui should link Ebitengine")
	}
}
