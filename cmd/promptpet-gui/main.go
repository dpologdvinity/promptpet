// Command promptpet-gui opens the promptpet companion window. It is a separate
// binary so that promptpet itself, which runs on every shell prompt, does not
// link Ebitengine (see internal/window). Start it with `promptpet gui`, which
// finds this binary next to promptpet or on PATH, or run it directly.
package main

import (
	"os"

	"github.com/dpologdvinity/promptpet/internal/gui"
	"github.com/dpologdvinity/promptpet/internal/window"
)

// version is set by GoReleaser with -X main.version.
var version = "dev"

func main() {
	os.Exit(window.Main(os.Args[1:], os.Getenv, os.Stdout, os.Stderr, version, gui.Run))
}
