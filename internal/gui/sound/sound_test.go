package sound

import (
	"testing"
)

func TestGenerateSquareSweep(t *testing.T) {
	buf := generateSquareSweep(440, 880, 0.05, 0.1)
	if len(buf) == 0 {
		t.Fatal("expected non-empty audio buffer")
	}
	expectedSamples := int(44100 * 0.05)
	expectedBytes := expectedSamples * 4
	if len(buf) != expectedBytes {
		t.Fatalf("expected %d bytes, got %d", expectedBytes, len(buf))
	}
}

func TestMuteToggling(t *testing.T) {
	orig := IsMuted()
	defer SetMuted(orig)

	SetMuted(false)
	if IsMuted() {
		t.Fatal("expected not muted")
	}

	muted := ToggleMute()
	if !muted || !IsMuted() {
		t.Fatal("expected muted after toggle")
	}

	unmuted := ToggleMute()
	if unmuted || IsMuted() {
		t.Fatal("expected unmuted after second toggle")
	}
}

func TestPlaySafety(t *testing.T) {
	// Calling play functions should never panic or crash even in test environment
	PlayBlip()
	PlayJump()
	PlaySuccess()
	PlayMiss()
}
