package gui

import (
	"testing"

	"github.com/dpologdvinity/promptpet/internal/companion"
	"github.com/dpologdvinity/promptpet/internal/gui/games"
)

func TestButtonAt(t *testing.T) {
	cases := []struct {
		x, y float32
		want games.Button
	}{
		{62, 276, games.ButtonA},
		{120, 286, games.ButtonB},
		{178, 276, games.ButtonC},
		{120 + 16, 286, games.ButtonB}, // inside B's larger radius
		{62 + 16, 276, games.ButtonNone},
		{120, 150, games.ButtonNone}, // LCD, not a button
	}
	for _, c := range cases {
		if got := buttonAt(c.x, c.y); got != c.want {
			t.Errorf("buttonAt(%v,%v) = %v, want %v", c.x, c.y, got, c.want)
		}
	}
}

func TestEveryCasingHasPalette(t *testing.T) {
	for _, c := range companion.Casings {
		if _, ok := casingPalettes[c]; !ok {
			t.Errorf("casing %q has no palette", c)
		}
	}
	if paletteFor("") != casingPalettes["lavender"] || paletteFor("plaid") != casingPalettes["lavender"] {
		t.Error("empty/unknown casing should fall back to lavender")
	}
}

func TestCornerPosition(t *testing.T) {
	if x, y := cornerPosition(1920, 1080); x != 1920-240-24 || y != 1080-340-72 {
		t.Errorf("1920x1080: got (%d,%d)", x, y)
	}
	if x, y := cornerPosition(100, 100); x != 0 || y != 0 {
		t.Errorf("tiny monitor must clamp to (0,0), got (%d,%d)", x, y)
	}
}
