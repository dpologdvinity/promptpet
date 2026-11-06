package games

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/dpologdvinity/promptpet/internal/companion"
)

// TestSimonSequenceLength checks DESIGN §5: Simon shows a 3-to-5 step sequence.
func TestSimonSequenceLength(t *testing.T) {
	p, _ := companion.GetPet(companion.PetUsagi)
	seen := map[int]bool{}
	for i := 0; i < 300; i++ {
		g := NewMemorySimon()
		g.Init(p)
		n := len(g.sequence)
		if n < 3 || n > 5 {
			t.Fatalf("sequence length %d outside 3..5", n)
		}
		seen[n] = true
	}
	if len(seen) != 3 {
		t.Errorf("expected lengths 3, 4 and 5 to all occur, saw %v", seen)
	}
}

// TestVirtualButtonKeys checks that face buttons stand in for their mapped keys.
func TestVirtualButtonKeys(t *testing.T) {
	t.Cleanup(func() { SetVirtualButton(ButtonNone, false) })

	SetVirtualButton(ButtonA, true)
	if !justPressed(ebiten.KeyArrowLeft) || !justPressed(ebiten.KeyDigit1) {
		t.Error("button A should press Left and 1")
	}
	if justPressed(ebiten.KeyArrowRight) || justPressed(ebiten.KeySpace) {
		t.Error("button A must not press Right or Space")
	}

	SetVirtualButton(ButtonB, false) // held, not just pressed
	if justPressed(ebiten.KeySpace) {
		t.Error("held button B must not count as just pressed")
	}
	if !pressed(ebiten.KeySpace) || !pressed(ebiten.KeyDigit2) {
		t.Error("held button B should hold Space and 2")
	}

	SetVirtualButton(ButtonNone, true)
	if pressed(ebiten.KeyArrowRight) || justPressed(ebiten.KeyDigit3) {
		t.Error("no button held must press nothing")
	}
}
