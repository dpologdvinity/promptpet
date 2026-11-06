package gui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/dpologdvinity/promptpet/internal/companion"
)

// Prop layout on the LCD. The speech bubble (incl. tail) covers y 52..106,
// so props stay below propTopY; held props sit in slots left/right of the pet.
const (
	propTopY   = float32(110)
	leftSlotX  = float32(52)
	rightSlotX = float32(188)
	slotY      = float32(178)
)

var (
	confettiColors = [...]color.RGBA{
		{R: 255, G: 150, B: 180, A: 255}, // pink
		{R: 140, G: 230, B: 200, A: 255}, // mint
		{R: 255, G: 230, B: 120, A: 255}, // lemon
		{R: 190, G: 160, B: 255, A: 255}, // lavender
		{R: 255, G: 190, B: 140, A: 255}, // peach
	}
	sparkleClr = color.RGBA{R: 255, G: 245, B: 180, A: 255}
	steamClr   = color.RGBA{R: 230, G: 230, B: 240, A: 150}
	outlineClr = color.RGBA{R: 60, G: 60, B: 80, A: 255}
	// sparkleOffsets places twinkles in a ring around the pet, clear of the bubble.
	sparkleOffsets = [...][2]float32{{-44, -18}, {42, -24}, {-50, 14}, {48, 10}, {-30, 34}, {34, 36}}
)

// drawScene draws the props around the pet for the current Scene:
// mood effects, wellness reminder props, and day/night accessories.
// All motion is derived from ticks so frames are deterministic.
func drawScene(screen *ebiten.Image, scene Scene, ticks int64, cx, cy float32) {
	if isNight(scene.Hour) {
		drawNightcap(screen, ticks, cx, cy)
	}

	switch scene.Mood {
	case companion.MoodEnergetic:
		drawSparkles(screen, ticks, cx, cy)
	case companion.MoodResting:
		drawZzz(screen, ticks, cx, cy)
		drawBook(screen, leftSlotX, slotY)
	case companion.MoodCheer:
		drawConfetti(screen, ticks)
	case companion.MoodSympathy:
		drawMug(screen, ticks, leftSlotX, slotY, color.RGBA{R: 250, G: 240, B: 225, A: 255}, color.RGBA{R: 200, G: 150, B: 90, A: 255})
		drawBandAid(screen, leftSlotX-2, slotY-30)
	case companion.MoodFanfare:
		drawStarBurst(screen, ticks, cx, cy)
		drawConfetti(screen, ticks)
	}

	switch scene.Wellness {
	case companion.WellnessWater:
		drawWaterGlass(screen, ticks, rightSlotX, slotY)
	case companion.WellnessStretch:
		drawStretchArrows(screen, ticks, cx, cy)
	case companion.WellnessLookAway:
		drawBinoculars(screen, ticks, rightSlotX, slotY)
	case companion.WellnessNone:
		// The morning coffee shares the right slot, so a reminder prop wins.
		if isMorning(scene.Hour) {
			drawMug(screen, ticks, rightSlotX, slotY, color.RGBA{R: 255, G: 200, B: 210, A: 255}, color.RGBA{R: 120, G: 80, B: 50, A: 255})
		}
	}
}

// isNight reports late-night hours (22:00-04:59), when the pet wears a nightcap.
func isNight(hour int) bool { return hour >= 22 || hour < 5 }

// isMorning reports morning hours (06:00-11:59), when the pet has a coffee.
func isMorning(hour int) bool { return hour >= 6 && hour < 12 }

// drawSparkle draws a four-point twinkle of the given half-size.
func drawSparkle(screen *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(screen, x-size, y, x+size, y, 1.5, clr, true)
	vector.StrokeLine(screen, x, y-size, x, y+size, 1.5, clr, true)
	vector.DrawFilledCircle(screen, x, y, size*0.35, clr, true)
}

// drawSparkles twinkles particles in a ring around the pet (energetic mood).
func drawSparkles(screen *ebiten.Image, ticks int64, cx, cy float32) {
	for i, off := range sparkleOffsets {
		size := float32(1 + 3*math.Abs(math.Sin(float64(ticks)*0.15+float64(i)*1.3)))
		drawSparkle(screen, cx+off[0], cy+off[1], size, sparkleClr)
	}
}

// drawZzz floats three staggered "z"s up and away from the pet's head (resting mood).
func drawZzz(screen *ebiten.Image, ticks int64, cx, cy float32) {
	for i := int64(0); i < 3; i++ {
		phase := (ticks + i*20) % 60
		x := cx + 26 + float32(i)*8 + float32(phase)/6
		y := cy - 30 - float32(phase)/3
		if y < propTopY {
			continue
		}
		ebitenutil.DebugPrintAt(screen, "z", int(x), int(y))
	}
}

// drawBook draws a tiny open book (resting mood).
func drawBook(screen *ebiten.Image, x, y float32) {
	coverClr := color.RGBA{R: 150, G: 180, B: 240, A: 255}
	pageClr := color.RGBA{R: 255, G: 252, B: 240, A: 255}
	drawRoundedRect(screen, x-13, y-1, 26, 14, 2, coverClr)
	vector.DrawFilledRect(screen, x-11, y, 10, 11, pageClr, true)
	vector.DrawFilledRect(screen, x+1, y, 10, 11, pageClr, true)
	for row := float32(3); row < 10; row += 3 {
		vector.StrokeLine(screen, x-9, y+row, x-3, y+row, 0.8, coverClr, true)
		vector.StrokeLine(screen, x+3, y+row, x+9, y+row, 0.8, coverClr, true)
	}
}

// drawConfetti rains small coloured flakes over the pet area (cheer, fanfare).
func drawConfetti(screen *ebiten.Image, ticks int64) {
	const n = 14
	for i := int64(0); i < n; i++ {
		x := 40 + float32(i*53%160) + float32(math.Sin(float64(ticks)*0.1+float64(i))*3)
		y := propTopY + float32((ticks*(1+i%3)+i*29)%100)
		w, h := float32(3), float32(2)
		if (ticks/8+i)%2 == 0 { // flip as it tumbles
			w, h = h, w
		}
		vector.DrawFilledRect(screen, x, y, w, h, confettiColors[i%int64(len(confettiColors))], true)
	}
}

// drawStarBurst pulses rays and stars out from behind the pet (fanfare).
func drawStarBurst(screen *ebiten.Image, ticks int64, cx, cy float32) {
	pulse := float32(math.Abs(math.Sin(float64(ticks) * 0.12)))
	inner, outer := float32(34), 40+8*pulse
	for i := 0; i < 8; i++ {
		a := float64(i)*math.Pi/4 + 0.2
		dx, dy := float32(math.Cos(a)), float32(math.Sin(a))
		if cy+dy*(outer+6) < propTopY { // keep clear of the speech bubble
			continue
		}
		clr := confettiColors[i%len(confettiColors)]
		vector.StrokeLine(screen, cx+dx*inner, cy+dy*inner, cx+dx*outer, cy+dy*outer, 2, clr, true)
		drawSparkle(screen, cx+dx*(outer+4), cy+dy*(outer+4), 2+2*pulse, sparkleClr)
	}
}

// drawMug draws a small mug with rising steam wisps (sympathy tea, morning coffee).
func drawMug(screen *ebiten.Image, ticks int64, x, y float32, mugClr, drinkClr color.RGBA) {
	vector.StrokeCircle(screen, x+8, y+6, 4, 2, mugClr, true) // handle
	drawRoundedRect(screen, x-8, y, 16, 14, 3, mugClr)
	vector.DrawFilledRect(screen, x-6, y+1, 12, 3, drinkClr, true)
	for i := 0; i < 2; i++ {
		rise := float32((ticks + int64(i)*15) % 30)
		sx := x - 3 + float32(i)*6 + float32(math.Sin(float64(ticks)*0.2+float64(i)))*1.5
		vector.DrawFilledCircle(screen, sx, y-3-rise/3, 1.6, steamClr, true)
	}
}

// drawBandAid draws a tiny band-aid (sympathy).
func drawBandAid(screen *ebiten.Image, x, y float32) {
	drawRoundedRect(screen, x-10, y, 20, 7, 3.5, color.RGBA{R: 245, G: 200, B: 160, A: 255})
	vector.DrawFilledRect(screen, x-4, y+1, 8, 5, color.RGBA{R: 255, G: 230, B: 210, A: 255}, true)
}

// drawWaterGlass draws a glass of water with drops dripping off it (water reminder).
func drawWaterGlass(screen *ebiten.Image, ticks int64, x, y float32) {
	vector.DrawFilledRect(screen, x-7, y-4, 14, 18, color.RGBA{R: 220, G: 240, B: 255, A: 120}, true)
	vector.DrawFilledRect(screen, x-6, y+2, 12, 11, color.RGBA{R: 120, G: 190, B: 255, A: 220}, true)
	vector.StrokeRect(screen, x-7, y-4, 14, 18, 1, outlineClr, true)
	dropClr := color.RGBA{R: 140, G: 200, B: 255, A: 255}
	for i := int64(0); i < 2; i++ {
		fall := float32((ticks + i*12) % 24)
		vector.DrawFilledCircle(screen, x-10+float32(i)*20, y-2+fall, 1.8, dropClr, true)
	}
}

// drawStretchArrows rises little up-arrows beside the pet's paws (stretch reminder).
func drawStretchArrows(screen *ebiten.Image, ticks int64, cx, cy float32) {
	clr := color.RGBA{R: 140, G: 230, B: 200, A: 255}
	for i, side := range [...]float32{-1, 1} {
		rise := float32((ticks + int64(i)*10) % 20)
		x, y := cx+side*36, cy+14-rise
		vector.StrokeLine(screen, x, y, x, y+10, 1.5, clr, true)
		vector.StrokeLine(screen, x, y, x-3, y+4, 1.5, clr, true)
		vector.StrokeLine(screen, x, y, x+3, y+4, 1.5, clr, true)
	}
}

// drawBinoculars draws tiny binoculars scanning side to side (look-away reminder).
func drawBinoculars(screen *ebiten.Image, ticks int64, x, y float32) {
	x += float32(math.Sin(float64(ticks)*0.08) * 3)
	bodyClr := color.RGBA{R: 90, G: 90, B: 120, A: 255}
	lensClr := color.RGBA{R: 170, G: 220, B: 255, A: 255}
	vector.DrawFilledRect(screen, x-3, y+2, 6, 4, bodyClr, true)
	for _, side := range [...]float32{-1, 1} {
		vector.DrawFilledCircle(screen, x+side*6, y+4, 5.5, bodyClr, true)
		vector.DrawFilledCircle(screen, x+side*6, y+4, 3, lensClr, true)
	}
}

// drawNightcap perches a slouchy striped nightcap on the pet's head (late night).
func drawNightcap(screen *ebiten.Image, ticks int64, cx, cy float32) {
	capClr := color.RGBA{R: 150, G: 160, B: 240, A: 255}
	trimClr := color.RGBA{R: 250, G: 250, B: 255, A: 255}
	// Cone built from shrinking circles drooping to the right.
	for i := float32(0); i < 5; i++ {
		vector.DrawFilledCircle(screen, cx+i*4, cy-30-i*2+i*i*0.5, 9-i*1.5, capClr, true)
	}
	tipSwing := float32(math.Sin(float64(ticks)*0.1) * 1.5)
	vector.DrawFilledCircle(screen, cx+20+tipSwing, cy-26, 3, trimClr, true)
	drawRoundedRect(screen, cx-13, cy-27, 26, 5, 2.5, trimClr)
}
