package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestPromptBinaryLatency runs the real promptpet binary, the way a shell hook
// does, and checks the prompt budget end to end. In-process benchmarks such as
// BenchmarkPromptWithEventLogging cannot catch startup costs: Go runs package
// init once per test binary, so a dependency whose init took 12 ms on every
// prompt (median 13.9 ms per prompt in total) went unnoticed until v0.5.0 was timed with hyperfine.
//
// Wall-clock gates are noisy on a busy laptop, so this only runs when
// PROMPTPET_LATENCY_TEST=1 (CI sets it on Linux).
func TestPromptBinaryLatency(t *testing.T) {
	if os.Getenv("PROMPTPET_LATENCY_TEST") != "1" {
		t.Skip("set PROMPTPET_LATENCY_TEST=1 to time the built binary")
	}
	bin := filepath.Join(t.TempDir(), "promptpet")
	build := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", bin, "..")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	home := t.TempDir()
	env := append(os.Environ(), "PROMPTPET_HOME="+home)
	run := func(extraEnv ...string) (time.Duration, string) {
		cmd := exec.Command(bin, "prompt", "--exit", "0", "--duration-ms", "12")
		cmd.Env = append(env, extraEnv...)
		cmd.Dir = home // not a git repo, like most prompts
		start := time.Now()
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("promptpet prompt: %v\n%s", err, out)
		}
		return time.Since(start), string(out)
	}

	// Package init must stay small: this is what regressed. It is a few short
	// measurements inside one run, so it is less exposed to scheduling noise
	// than the whole-process timings below.
	_, trace := run("GODEBUG=inittrace=1")
	var initMs float64
	for _, line := range strings.Split(trace, "\n") {
		// "init <pkg> @0.1 ms, 0.02 ms clock, ..."
		if f := strings.Fields(line); len(f) > 6 && f[0] == "init" && f[6] == "clock," {
			ms, _ := strconv.ParseFloat(f[4], 64)
			initMs += ms
		}
	}
	if initMs > 5 {
		t.Errorf("package init takes %.1f ms per prompt (budget 5 ms):\n%s", initMs, trace)
	}

	const runs = 40
	times := make([]time.Duration, runs)
	for i := range times {
		times[i], _ = run()
	}
	slices.Sort(times)
	median := times[runs/2]
	t.Logf("promptpet prompt: median %v, min %v, max %v over %d runs", median, times[0], times[runs-1], runs)
	if median > 10*time.Millisecond {
		t.Errorf("median prompt time %v exceeds the 10 ms budget", median)
	}
}
