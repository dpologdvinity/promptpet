package games

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/dpologdvinity/promptpet/internal/companion"
)

// MiniGame defines the standard interface for all interactive handheld games.
type MiniGame interface {
	Name() string
	Init(pet companion.PetDef)
	Update() (done bool, won bool, score int)
	Draw(screen *ebiten.Image)
}

func hexToRGBA(c uint32, alpha uint8) color.RGBA {
	return color.RGBA{
		R: uint8(c >> 16),
		G: uint8(c >> 8),
		B: uint8(c),
		A: alpha,
	}
}

func drawRoundedRect(dst *ebiten.Image, x, y, width, height, radius float32, clr color.Color) {
	vector.DrawFilledCircle(dst, x+radius, y+radius, radius, clr, true)
	vector.DrawFilledCircle(dst, x+width-radius, y+radius, radius, clr, true)
	vector.DrawFilledCircle(dst, x+radius, y+height-radius, radius, clr, true)
	vector.DrawFilledCircle(dst, x+width-radius, y+height-radius, radius, clr, true)
	vector.DrawFilledRect(dst, x+radius, y, width-2*radius, height, clr, true)
	vector.DrawFilledRect(dst, x, y+radius, width, height-2*radius, clr, true)
}
