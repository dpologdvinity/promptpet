package gui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/dpologdvinity/promptpet/internal/companion"
	"github.com/dpologdvinity/promptpet/internal/gui/games"
)

// rgb builds an opaque colour.
func rgb(r, g, b uint8) color.RGBA { return color.RGBA{R: r, G: g, B: b, A: 255} }

// casingPalette holds the shell colours for one device casing.
type casingPalette struct {
	face, shadow, loop, speaker color.RGBA
}

// casingPalettes maps companion.Casings names to their colours.
var casingPalettes = map[string]casingPalette{
	"lavender": {face: rgb(216, 206, 242), shadow: rgb(165, 155, 198), loop: rgb(180, 175, 205), speaker: rgb(170, 160, 195)},
	"coral":    {face: rgb(248, 192, 182), shadow: rgb(205, 140, 130), loop: rgb(225, 170, 160), speaker: rgb(210, 150, 140)},
	"mint":     {face: rgb(190, 236, 210), shadow: rgb(135, 190, 160), loop: rgb(165, 210, 185), speaker: rgb(145, 195, 170)},
	"midnight": {face: rgb(58, 62, 96), shadow: rgb(34, 36, 60), loop: rgb(80, 84, 120), speaker: rgb(90, 95, 135)},
}

// paletteFor returns the palette for a casing name, falling back to lavender.
func paletteFor(casing string) casingPalette {
	if p, ok := casingPalettes[casing]; ok {
		return p
	}
	return casingPalettes["lavender"]
}

// faceButton is one A/B/C button; drawing and click hit-testing share this geometry.
type faceButton struct {
	button    games.Button
	x, y, r   float32
	label     string
	face, rim color.RGBA
}

var faceButtons = []faceButton{
	{games.ButtonA, 62, 276, 15, "A", rgb(243, 139, 168), rgb(195, 95, 125)},   // coral
	{games.ButtonB, 120, 286, 17, "B", rgb(166, 227, 161), rgb(110, 175, 110)}, // mint, larger action button
	{games.ButtonC, 178, 276, 15, "C", rgb(180, 190, 254), rgb(120, 135, 210)}, // blue
}

// buttonAt returns the face button under screen point (x, y), or games.ButtonNone.
func buttonAt(x, y float32) games.Button {
	for _, b := range faceButtons {
		dx, dy := x-b.x, y-b.y
		if dx*dx+dy*dy <= b.r*b.r {
			return b.button
		}
	}
	return games.ButtonNone
}

// DrawDevice renders the retro handheld console casing:
// outer pastel frame with rounded corners, top keychain loop,
// center LCD screen bevel (dark glass frame), and 3 circular buttons (A, B, C).
// casing selects the shell colour (see companion.Casings).
func DrawDevice(screen *ebiten.Image, pet companion.PetDef, casing string) {
	pal := paletteFor(casing)

	// 1. Overall window backdrop (soft dark vignette)
	backdropClr := color.RGBA{R: 20, G: 21, B: 30, A: 255}
	vector.DrawFilledRect(screen, 0, 0, 240, 340, backdropClr, false)

	// 2. Top keychain loop
	loopCenterX := float32(120)
	loopCenterY := float32(12)
	loopOuterRadius := float32(9)
	loopInnerRadius := float32(4.5)
	vector.DrawFilledCircle(screen, loopCenterX, loopCenterY, loopOuterRadius, pal.loop, true)
	vector.DrawFilledCircle(screen, loopCenterX, loopCenterY, loopInnerRadius, backdropClr, true)

	// 3. Outer pastel chassis frame with rounded corners
	chassisX := float32(8)
	chassisY := float32(16)
	chassisW := float32(224)
	chassisH := float32(314)
	chassisRadius := float32(26)

	// Drop shadow / 3D edge under chassis
	drawRoundedRect(screen, chassisX, chassisY+3, chassisW, chassisH, chassisRadius, pal.shadow)

	// Casing face
	drawRoundedRect(screen, chassisX, chassisY, chassisW, chassisH, chassisRadius, pal.face)

	// Console branding / header
	ebitenutil.DebugPrintAt(screen, "PROMPTPET", 86, 26)

	// 4. Center LCD screen bevel (dark glass frame)
	bevelX := float32(22)
	bevelY := float32(42)
	bevelW := float32(196)
	bevelH := float32(190)
	bevelRadius := float32(14)
	bevelClr := color.RGBA{R: 45, G: 48, B: 64, A: 255}
	drawRoundedRect(screen, bevelX, bevelY, bevelW, bevelH, bevelRadius, bevelClr)

	// Inner bevel rim
	rimX := float32(25)
	rimY := float32(45)
	rimW := float32(190)
	rimH := float32(184)
	rimRadius := float32(10)
	rimClr := color.RGBA{R: 32, G: 34, B: 46, A: 255}
	drawRoundedRect(screen, rimX, rimY, rimW, rimH, rimRadius, rimClr)

	// LCD display surface (dark glass retro LCD)
	lcdX := float32(28)
	lcdY := float32(48)
	lcdW := float32(184)
	lcdH := float32(178)
	lcdRadius := float32(8)
	lcdClr := color.RGBA{R: 22, G: 24, B: 35, A: 255}
	drawRoundedRect(screen, lcdX, lcdY, lcdW, lcdH, lcdRadius, lcdClr)

	// 5. Bottom face buttons (A, B, C)
	for _, b := range faceButtons {
		drawButton(screen, b.x, b.y, b.r, b.label, b.face, b.rim)
	}

	// Speaker grille dots (bottom right casing)
	speakerClr := pal.speaker
	vector.DrawFilledCircle(screen, 192, 246, 2, speakerClr, true)
	vector.DrawFilledCircle(screen, 200, 246, 2, speakerClr, true)
	vector.DrawFilledCircle(screen, 196, 252, 2, speakerClr, true)
	vector.DrawFilledCircle(screen, 204, 252, 2, speakerClr, true)
}

func drawButton(screen *ebiten.Image, cx, cy, r float32, label string, faceClr, shadowClr color.Color) {
	// 3D bottom bevel
	vector.DrawFilledCircle(screen, cx, cy+2.5, r, shadowClr, true)
	// Button face
	vector.DrawFilledCircle(screen, cx, cy, r, faceClr, true)
	// Inner highlight ring
	vector.StrokeCircle(screen, cx, cy, r-2, 1, color.RGBA{R: 255, G: 255, B: 255, A: 70}, true)
	// Button letter
	ebitenutil.DebugPrintAt(screen, label, int(cx-3), int(cy-8))
}

func drawRoundedRect(dst *ebiten.Image, x, y, width, height, radius float32, clr color.Color) {
	// Four rounded corners
	vector.DrawFilledCircle(dst, x+radius, y+radius, radius, clr, true)
	vector.DrawFilledCircle(dst, x+width-radius, y+radius, radius, clr, true)
	vector.DrawFilledCircle(dst, x+radius, y+height-radius, radius, clr, true)
	vector.DrawFilledCircle(dst, x+width-radius, y+height-radius, radius, clr, true)

	// Center body rectangles
	vector.DrawFilledRect(dst, x+radius, y, width-2*radius, height, clr, true)
	vector.DrawFilledRect(dst, x, y+radius, width, height-2*radius, clr, true)
}
