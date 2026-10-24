package cli

import (
	"bytes"
	"math/rand/v2"
	"strings"
	"testing"
	"time"
)

var t0 = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

func testDeps(t *testing.T) (Deps, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	out, errb := new(bytes.Buffer), new(bytes.Buffer)
	home, cwd := t.TempDir(), t.TempDir()
	d := Deps{Version: "dev", Now: func() time.Time { return t0 }, Rand: rand.New(rand.NewPCG(1, 2)), Getenv: func(k string) string {
		if k == "PROMPTPET_HOME" {
			return home
		}
		return ""
	}, Getwd: func() (string, error) { return cwd, nil }, Spawn: func(string) error { return nil }, Stdout: out, Stderr: errb}
	return d, out, errb
}
func TestVersion(t *testing.T) {
	d, out, _ := testDeps(t)
	d.Version = "1.2.3"
	if code := Run([]string{"--version"}, d); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if out.String() != "promptpet 1.2.3\n" {
		t.Fatalf("got %q", out.String())
	}
}
func TestHelpAndUnknown(t *testing.T) {
	d, out, errb := testDeps(t)
	if Run([]string{"--help"}, d) != 0 || !strings.HasPrefix(out.String(), "promptpet - a tiny pet that lives in your shell prompt\n") {
		t.Fatal(out.String())
	}
	if Run([]string{"bogus"}, d) != 2 || errb.String() != "promptpet: unknown command \"bogus\" (see promptpet --help)\n" {
		t.Fatal(errb.String())
	}
}
