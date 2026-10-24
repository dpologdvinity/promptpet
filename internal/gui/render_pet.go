package gui

import (
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/dpologdvinity/promptpet/internal/companion"
)

func hexToRGBA(c uint32, alpha uint8) color.RGBA {
	return color.RGBA{
		R: uint8(c >> 16),
		G: uint8(c >> 8),
		B: uint8(c),
		A: alpha,
	}
}

// DrawPet renders the animated mascot inside the LCD screen area.
func DrawPet(screen *ebiten.Image, pet companion.PetDef, ticks int64) {
	centerX := float32(120)
	centerY := float32(152)

	// Subtle idle breathing/bobbing animation
	bobY := float32(math.Sin(float64(ticks)*0.1) * 3.0)
	cy := centerY + bobY
	cx := centerX

	primClr := hexToRGBA(pet.PrimaryColor, 255)
	accentClr := hexToRGBA(pet.AccentColor, 255)
	blushClr := hexToRGBA(pet.BlushColor, 255)
	darkEyeClr := color.RGBA{R: 24, G: 24, B: 36, A: 255}
	whiteGlintClr := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// 1. Soft floor shadow on the LCD base
	shadowW := float32(40.0) - bobY*1.5
	if shadowW < 24 {
		shadowW = 24
	}
	shadowClr := color.RGBA{R: 14, G: 15, B: 22, A: 160}
	drawRoundedRect(screen, cx-shadowW/2, 192, shadowW, 6, 3, shadowClr)

	// 2. Render mascot based on PetID
	switch pet.ID {
	case companion.PetUsagi:
		drawUsagi(screen, cx, cy, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr)
	case companion.PetKochi:
		drawKochi(screen, cx, cy, ticks, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr)
	case companion.PetPuff:
		drawPuff(screen, cx, cy, ticks, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr)
	case companion.PetMochi:
		drawMochi(screen, cx, cy, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr)
	case companion.PetSpook:
		drawSpook(screen, cx, cy, ticks, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr)
	case companion.PetPippin:
		drawPippin(screen, cx, cy, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr)
	case companion.PetGromit:
		drawGromit(screen, cx, cy, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr)
	case companion.PetBoba:
		drawBoba(screen, cx, cy, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr)
	default:
		// Fallback generic cute blob
		vector.DrawFilledCircle(screen, cx, cy, 26, primClr, true)
		vector.DrawFilledCircle(screen, cx-8, cy-2, 3, darkEyeClr, true)
		vector.DrawFilledCircle(screen, cx+8, cy-2, 3, darkEyeClr, true)
		vector.DrawFilledCircle(screen, cx-14, cy+5, 4, blushClr, true)
		vector.DrawFilledCircle(screen, cx+14, cy+5, 4, blushClr, true)
	}
}

// 1. Usagi: Dumpling Bunny
func drawUsagi(screen *ebiten.Image, cx, cy float32, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr color.RGBA) {
	// Floppy bunny ears
	// Outer ears
	drawRoundedRect(screen, cx-18, cy-32, 9, 22, 4.5, primClr)
	drawRoundedRect(screen, cx+9, cy-32, 9, 22, 4.5, primClr)
	// Inner ear pastel accent
	drawRoundedRect(screen, cx-16.5, cy-28, 6, 15, 3, accentClr)
	drawRoundedRect(screen, cx+10.5, cy-28, 6, 15, 3, accentClr)

	// Round dumpling body
	vector.DrawFilledCircle(screen, cx, cy, 26, primClr, true)

	// Cute dumpling paws
	vector.DrawFilledCircle(screen, cx-7, cy+18, 4, primClr, true)
	vector.DrawFilledCircle(screen, cx+7, cy+18, 4, primClr, true)

	// Eyes with shine
	drawEyes(screen, cx, cy-1, 3, 8, darkEyeClr, whiteGlintClr)

	// Tiny pink mouth / nose
	vector.DrawFilledCircle(screen, cx, cy+3, 1.8, accentClr, true)

	// Cheeks
	vector.DrawFilledCircle(screen, cx-14, cy+5, 4.5, blushClr, true)
	vector.DrawFilledCircle(screen, cx+14, cy+5, 4.5, blushClr, true)
}

// 2. Kochi: Shiba Pup
func drawKochi(screen *ebiten.Image, cx, cy float32, ticks int64, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr color.RGBA) {
	// Pointy triangular ears
	vector.DrawFilledCircle(screen, cx-16, cy-18, 8, primClr, true)
	vector.DrawFilledCircle(screen, cx-16, cy-18, 4.5, color.RGBA{R: 255, G: 245, B: 235, A: 255}, true)
	vector.DrawFilledCircle(screen, cx+16, cy-18, 8, primClr, true)
	vector.DrawFilledCircle(screen, cx+16, cy-18, 4.5, color.RGBA{R: 255, G: 245, B: 235, A: 255}, true)

	// Curled shiba tail with wagging animation
	tailWag := float32(math.Sin(float64(ticks)*0.25) * 3.0)
	vector.DrawFilledCircle(screen, cx+26, cy+4+tailWag, 7, primClr, true)
	vector.DrawFilledCircle(screen, cx+26, cy+4+tailWag, 3, color.RGBA{R: 255, G: 245, B: 235, A: 255}, true)

	// Round golden head/body
	vector.DrawFilledCircle(screen, cx, cy, 26, primClr, true)

	// White dumpling cheeks / muzzle
	muzzleClr := color.RGBA{R: 255, G: 245, B: 235, A: 255}
	drawRoundedRect(screen, cx-11, cy+2, 22, 13, 6, muzzleClr)

	// Green bandana around neck
	vector.DrawFilledRect(screen, cx-14, cy+15, 28, 6, accentClr, true)
	// Bandana knot
	vector.DrawFilledCircle(screen, cx+12, cy+18, 4, accentClr, true)

	// Pup nose
	vector.DrawFilledCircle(screen, cx, cy+4, 2.5, darkEyeClr, true)

	// Eyes with shine
	drawEyes(screen, cx, cy-2, 3, 9, darkEyeClr, whiteGlintClr)

	// Cheeks
	vector.DrawFilledCircle(screen, cx-16, cy+6, 4.5, blushClr, true)
	vector.DrawFilledCircle(screen, cx+16, cy+6, 4.5, blushClr, true)
}

// 3. Puff: Baby Cloud Dragon
func drawPuff(screen *ebiten.Image, cx, cy float32, ticks int64, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr color.RGBA) {
	// Tiny golden horns
	vector.DrawFilledCircle(screen, cx-11, cy-24, 4.5, accentClr, true)
	vector.DrawFilledCircle(screen, cx+11, cy-24, 4.5, accentClr, true)

	// Soft flapping baby wings
	wingFlap := float32(math.Sin(float64(ticks)*0.15) * 2.0)
	vector.DrawFilledCircle(screen, cx-25, cy+wingFlap, 7, primClr, true)
	vector.DrawFilledCircle(screen, cx+25, cy+wingFlap, 7, primClr, true)

	// Baby dragon body
	vector.DrawFilledCircle(screen, cx, cy, 26, primClr, true)

	// Golden tummy belly
	vector.DrawFilledCircle(screen, cx, cy+9, 14, accentClr, true)

	// Eyes
	drawEyes(screen, cx, cy-2, 3, 8, darkEyeClr, whiteGlintClr)

	// Cute nostrils
	vector.DrawFilledCircle(screen, cx-2.5, cy+3, 1.2, darkEyeClr, true)
	vector.DrawFilledCircle(screen, cx+2.5, cy+3, 1.2, darkEyeClr, true)

	// Tiny heart spark puff near mouth
	sparkleY := float32(math.Sin(float64(ticks)*0.1) * 2.0)
	vector.DrawFilledCircle(screen, cx+18, cy-10+sparkleY, 3, blushClr, true)

	// Cheeks
	vector.DrawFilledCircle(screen, cx-14, cy+5, 4, blushClr, true)
	vector.DrawFilledCircle(screen, cx+14, cy+5, 4, blushClr, true)
}

// 4. Mochi: Strawberry Cat
func drawMochi(screen *ebiten.Image, cx, cy float32, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr color.RGBA) {
	// Cat ears
	vector.DrawFilledCircle(screen, cx-16, cy-20, 7.5, primClr, true)
	vector.DrawFilledCircle(screen, cx-16, cy-20, 4, accentClr, true)
	vector.DrawFilledCircle(screen, cx+16, cy-20, 7.5, primClr, true)
	vector.DrawFilledCircle(screen, cx+16, cy-20, 4, accentClr, true)

	// Calico strawberry patch over right ear
	vector.DrawFilledCircle(screen, cx+14, cy-12, 6, accentClr, true)

	// Cat round body
	vector.DrawFilledCircle(screen, cx, cy, 26, primClr, true)

	// Whiskers (left & right)
	vector.StrokeLine(screen, cx-24, cy+4, cx-15, cy+3, 1.2, darkEyeClr, true)
	vector.StrokeLine(screen, cx-24, cy+8, cx-15, cy+7, 1.2, darkEyeClr, true)
	vector.StrokeLine(screen, cx+15, cy+3, cx+24, cy+4, 1.2, darkEyeClr, true)
	vector.StrokeLine(screen, cx+15, cy+7, cx+24, cy+8, 1.2, darkEyeClr, true)

	// Cat eyes
	drawEyes(screen, cx, cy-1, 3, 8, darkEyeClr, whiteGlintClr)

	// Tiny nose
	vector.DrawFilledCircle(screen, cx, cy+3, 1.5, accentClr, true)

	// Cheeks
	vector.DrawFilledCircle(screen, cx-14, cy+6, 4.5, blushClr, true)
	vector.DrawFilledCircle(screen, cx+14, cy+6, 4.5, blushClr, true)
}

// 5. Spook: Cloud Ghostie
func drawSpook(screen *ebiten.Image, cx, cy float32, ticks int64, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr color.RGBA) {
	// Outer translucent ethereal glow
	glowClr := color.RGBA{R: accentClr.R, G: accentClr.G, B: accentClr.B, A: 65}
	vector.DrawFilledCircle(screen, cx, cy-2, 32, glowClr, true)

	// Ghost rounded top head
	vector.DrawFilledCircle(screen, cx, cy-4, 24, primClr, true)
	// Ghost body rectangle
	vector.DrawFilledRect(screen, cx-24, cy-4, 48, 18, primClr, true)

	// Floating ghost hands
	vector.DrawFilledCircle(screen, cx-23, cy+3, 5, primClr, true)
	vector.DrawFilledCircle(screen, cx+23, cy+3, 5, primClr, true)

	// 3 wavy bottom wiggles oscillating
	for i := -1; i <= 1; i++ {
		wave := float32(math.Sin(float64(ticks)*0.15+float64(i)*1.5) * 2.5)
		scallopX := cx + float32(i)*15
		scallopY := cy + 14 + wave
		vector.DrawFilledCircle(screen, scallopX, scallopY, 8, primClr, true)
	}

	// Tall curious ghost eyes
	drawEyes(screen, cx, cy-4, 4, 8, darkEyeClr, whiteGlintClr)

	// Cheeks
	vector.DrawFilledCircle(screen, cx-14, cy+3, 4, blushClr, true)
	vector.DrawFilledCircle(screen, cx+14, cy+3, 4, blushClr, true)
}

// 6. Pippin: Cozy Penguin
func drawPippin(screen *ebiten.Image, cx, cy float32, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr color.RGBA) {
	// Penguin dark body
	vector.DrawFilledCircle(screen, cx, cy, 26, primClr, true)

	// Cream white egg tummy
	tummyClr := color.RGBA{R: 245, G: 245, B: 250, A: 255}
	drawRoundedRect(screen, cx-13, cy-10, 26, 32, 12, tummyClr)

	// Cozy knitted scarf
	vector.DrawFilledRect(screen, cx-16, cy+8, 32, 7, accentClr, true)
	// Scarf dangling tail
	vector.DrawFilledRect(screen, cx+7, cy+14, 6, 10, accentClr, true)

	// Flippers on sides
	vector.DrawFilledCircle(screen, cx-22, cy+4, 6, primClr, true)
	vector.DrawFilledCircle(screen, cx+22, cy+4, 6, primClr, true)

	// Orange beak
	beakClr := color.RGBA{R: 250, G: 160, B: 60, A: 255}
	vector.DrawFilledCircle(screen, cx, cy+1, 3, beakClr, true)

	// Orange feet
	vector.DrawFilledCircle(screen, cx-7, cy+23, 4, beakClr, true)
	vector.DrawFilledCircle(screen, cx+7, cy+23, 4, beakClr, true)

	// Eyes with shine
	drawEyes(screen, cx, cy-3, 3, 8, darkEyeClr, whiteGlintClr)

	// Cheeks
	vector.DrawFilledCircle(screen, cx-14, cy+3, 4, blushClr, true)
	vector.DrawFilledCircle(screen, cx+14, cy+3, 4, blushClr, true)
}

// 7. Gromit: Rainy Froggy
func drawGromit(screen *ebiten.Image, cx, cy float32, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr color.RGBA) {
	// Big elevated froggy eye bumps
	vector.DrawFilledCircle(screen, cx-13, cy-18, 9, primClr, true)
	vector.DrawFilledCircle(screen, cx+13, cy-18, 9, primClr, true)

	// Green frog head/body
	vector.DrawFilledCircle(screen, cx, cy, 25, primClr, true)

	// Yellow tummy
	vector.DrawFilledCircle(screen, cx, cy+8, 13, accentClr, true)

	// Big eyes inside the bumps
	vector.DrawFilledCircle(screen, cx-13, cy-18, 5, darkEyeClr, true)
	vector.DrawFilledCircle(screen, cx-14.5, cy-19.5, 1.8, whiteGlintClr, true)
	vector.DrawFilledCircle(screen, cx+13, cy-18, 5, darkEyeClr, true)
	vector.DrawFilledCircle(screen, cx+11.5, cy-19.5, 1.8, whiteGlintClr, true)

	// Wide froggy smile line
	smileClr := color.RGBA{R: 50, G: 100, B: 60, A: 255}
	vector.StrokeLine(screen, cx-8, cy+2, cx+8, cy+2, 1.5, smileClr, true)

	// Leaf umbrella stem above head
	leafClr := color.RGBA{R: 120, G: 200, B: 100, A: 255}
	vector.StrokeLine(screen, cx+2, cy-22, cx+6, cy-32, 1.5, leafClr, true)
	drawRoundedRect(screen, cx-2, cy-36, 18, 6, 3, leafClr)

	// Cheeks
	vector.DrawFilledCircle(screen, cx-15, cy+4, 4.5, blushClr, true)
	vector.DrawFilledCircle(screen, cx+15, cy+4, 4.5, blushClr, true)
}

// 8. Boba: Tea Panda
func drawBoba(screen *ebiten.Image, cx, cy float32, primClr, accentClr, blushClr, darkEyeClr, whiteGlintClr color.RGBA) {
	// Panda round black ears
	vector.DrawFilledCircle(screen, cx-18, cy-18, 8.5, accentClr, true)
	vector.DrawFilledCircle(screen, cx+18, cy-18, 8.5, accentClr, true)

	// Milk-tea beige head
	vector.DrawFilledCircle(screen, cx, cy, 26, primClr, true)

	// Dark slanted eye patches
	drawRoundedRect(screen, cx-16, cy-5, 12, 9, 4.5, accentClr)
	drawRoundedRect(screen, cx+4, cy-5, 12, 9, 4.5, accentClr)

	// Eyes inside patches
	vector.DrawFilledCircle(screen, cx-10, cy-1, 2.5, whiteGlintClr, true)
	vector.DrawFilledCircle(screen, cx-10, cy-1, 1.5, darkEyeClr, true)
	vector.DrawFilledCircle(screen, cx+10, cy-1, 2.5, whiteGlintClr, true)
	vector.DrawFilledCircle(screen, cx+10, cy-1, 1.5, darkEyeClr, true)

	// Tiny nose
	vector.DrawFilledCircle(screen, cx, cy+4, 2, accentClr, true)

	// Mini boba cup held between paws
	cupClr := color.RGBA{R: 245, G: 240, B: 235, A: 240}
	drawRoundedRect(screen, cx-7, cy+14, 14, 13, 3, cupClr)
	// Boba straw
	vector.DrawFilledRect(screen, cx+1, cy+8, 2, 8, accentClr, true)
	// Boba tapioca pearls
	vector.DrawFilledCircle(screen, cx-3, cy+22, 1.8, accentClr, true)
	vector.DrawFilledCircle(screen, cx+2, cy+22, 1.8, accentClr, true)

	// Cheeks
	vector.DrawFilledCircle(screen, cx-16, cy+5, 4, blushClr, true)
	vector.DrawFilledCircle(screen, cx+16, cy+5, 4, blushClr, true)
}

func drawEyes(screen *ebiten.Image, cx, cy, radius, spacing float32, darkClr, glintClr color.RGBA) {
	leftX := cx - spacing
	rightX := cx + spacing

	// Left eye
	vector.DrawFilledCircle(screen, leftX, cy, radius, darkClr, true)
	vector.DrawFilledCircle(screen, leftX-1, cy-1, radius*0.4, glintClr, true)

	// Right eye
	vector.DrawFilledCircle(screen, rightX, cy, radius, darkClr, true)
	vector.DrawFilledCircle(screen, rightX-1, cy-1, radius*0.4, glintClr, true)
}

// DrawSpeechBubble renders a dialogue bubble inside the LCD area.
func DrawSpeechBubble(screen *ebiten.Image, speech string, pet companion.PetDef) {
	if strings.TrimSpace(speech) == "" {
		speech = "..."
	}

	bubbleX := float32(34)
	bubbleY := float32(52)
	bubbleW := float32(172)
	bubbleH := float32(48)
	bubbleRadius := float32(8)

	// Bubble background (sleek dark glass with translucent depth)
	bubbleClr := color.RGBA{R: 34, G: 37, B: 52, A: 240}
	drawRoundedRect(screen, bubbleX, bubbleY, bubbleW, bubbleH, bubbleRadius, bubbleClr)

	// Accent border
	accentBorderClr := hexToRGBA(pet.AccentColor, 200)
	vector.StrokeRect(screen, bubbleX, bubbleY, bubbleW, bubbleH, 1.2, accentBorderClr, true)

	// Bubble tail pointing down toward pet
	tailClr := bubbleClr
	vector.DrawFilledCircle(screen, 120, bubbleY+bubbleH+2, 4, tailClr, true)

	// Format and render text
	wrapped := wrapText(speech, 24, 2)
	ebitenutil.DebugPrintAt(screen, wrapped, int(bubbleX+8), int(bubbleY+8))
}

func wrapText(text string, maxChars, maxLines int) string {
	if len(text) <= maxChars {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	cur := ""
	for _, w := range words {
		if cur == "" {
			cur = w
		} else if len(cur)+1+len(w) <= maxChars {
			cur += " " + w
		} else {
			lines = append(lines, cur)
			cur = w
			if len(lines) == maxLines {
				break
			}
		}
	}
	if len(lines) < maxLines && cur != "" {
		lines = append(lines, cur)
	}
	return strings.Join(lines, "\n")
}
