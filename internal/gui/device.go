package gui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/dpologdvinity/promptpet/internal/companion"
)

// DrawDevice renders the retro handheld console casing:
// outer pastel frame with rounded corners, top keychain loop,
// center LCD screen bevel (dark glass frame), and 3 circular buttons (A, B, C).
func DrawDevice(screen *ebiten.Image, pet companion.PetDef) {
	// 1. Overall window backdrop (soft dark vignette)
	backdropClr := color.RGBA{R: 20, G: 21, B: 30, A: 255}
	vector.DrawFilledRect(screen, 0, 0, 240, 340, backdropClr, false)

	// 2. Top keychain loop
	loopCenterX := float32(120)
	loopCenterY := float32(12)
	loopOuterRadius := float32(9)
	loopInnerRadius := float32(4.5)
	loopClr := color.RGBA{R: 180, G: 175, B: 205, A: 255}
	vector.DrawFilledCircle(screen, loopCenterX, loopCenterY, loopOuterRadius, loopClr, true)
	vector.DrawFilledCircle(screen, loopCenterX, loopCenterY, loopInnerRadius, backdropClr, true)

	// 3. Outer pastel chassis frame with rounded corners
	chassisX := float32(8)
	chassisY := float32(16)
	chassisW := float32(224)
	chassisH := float32(314)
	chassisRadius := float32(26)

	// Drop shadow / 3D edge under chassis
	chassisShadowClr := color.RGBA{R: 165, G: 155, B: 198, A: 255}
	drawRoundedRect(screen, chassisX, chassisY+3, chassisW, chassisH, chassisRadius, chassisShadowClr)

	// Pastel lavender casing face
	casingFaceClr := color.RGBA{R: 216, G: 206, B: 242, A: 255}
	drawRoundedRect(screen, chassisX, chassisY, chassisW, chassisH, chassisRadius, casingFaceClr)

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
	// Button A: Left
	btnAY := float32(276)
	drawButton(screen, 62, btnAY, 15, "A",
		color.RGBA{R: 243, G: 139, B: 168, A: 255}, // Coral face
		color.RGBA{R: 195, G: 95, B: 125, A: 255},  // Dark rim
	)

	// Button B: Middle (larger jump / action button)
	btnBY := float32(286)
	drawButton(screen, 120, btnBY, 17, "B",
		color.RGBA{R: 166, G: 227, B: 161, A: 255}, // Mint face
		color.RGBA{R: 110, G: 175, B: 110, A: 255}, // Dark rim
	)

	// Button C: Right
	btnCY := float32(276)
	drawButton(screen, 178, btnCY, 15, "C",
		color.RGBA{R: 180, G: 190, B: 254, A: 255}, // Blue face
		color.RGBA{R: 120, G: 135, B: 210, A: 255}, // Dark rim
	)

	// Speaker grille dots (bottom right casing)
	speakerClr := color.RGBA{R: 170, G: 160, B: 195, A: 255}
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
