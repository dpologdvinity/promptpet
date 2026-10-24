package games

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/dpologdvinity/promptpet/internal/companion"
)

type fallingSnack struct {
	x      float32
	y      float32
	speed  float32
	active bool
}

// SnackCatch is a mini-game where the pet scurries left and right to catch falling snacks.
// Catch 5 snacks to win before 3 misses.
type SnackCatch struct {
	pet        companion.PetDef
	petX       float32
	petY       float32
	petSpeed   float32
	caught     int
	misses     int
	targetWins int
	maxMisses  int
	snack      fallingSnack
	spawnTimer int
	ticks      int64
	done       bool
	won        bool
}

// NewSnackCatch creates a new instance of SnackCatch.
func NewSnackCatch() *SnackCatch {
	return &SnackCatch{
		petX:       120,
		petY:       196,
		petSpeed:   3.5,
		targetWins: 5,
		maxMisses:  3,
		spawnTimer: 10,
	}
}

func (g *SnackCatch) Name() string {
	return "Snack Catch"
}

func (g *SnackCatch) Init(pet companion.PetDef) {
	g.pet = pet
	g.petX = 120
	g.petY = 196
	g.caught = 0
	g.misses = 0
	g.targetWins = 5
	g.maxMisses = 3
	g.snack = fallingSnack{active: false}
	g.spawnTimer = 10
	g.ticks = 0
	g.done = false
	g.won = false
}

// CatchSnack manually registers a caught snack (used in testing or automation).
func (g *SnackCatch) CatchSnack() {
	g.caught++
	g.snack.active = false
	g.spawnTimer = 15
}

// MissSnack manually registers a missed snack (used in testing or automation).
func (g *SnackCatch) MissSnack() {
	g.misses++
	g.snack.active = false
	g.spawnTimer = 15
}

func (g *SnackCatch) Update() (done bool, won bool, score int) {
	if g.done {
		return true, g.won, g.calculateScore()
	}

	if g.caught >= g.targetWins {
		g.done = true
		g.won = true
		return true, true, g.calculateScore()
	}
	if g.misses >= g.maxMisses {
		g.done = true
		g.won = false
		return true, false, g.calculateScore()
	}

	g.ticks++

	// 1. Controls
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.petX -= g.petSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.petX += g.petSpeed
	}

	// Clamp pet inside LCD
	if g.petX < 45 {
		g.petX = 45
	}
	if g.petX > 195 {
		g.petX = 195
	}

	// 2. Snack spawning
	if !g.snack.active {
		g.spawnTimer--
		if g.spawnTimer <= 0 {
			snackX := float32(50 + rand.Intn(140))
			g.snack = fallingSnack{
				x:      snackX,
				y:      55,
				speed:  2.2,
				active: true,
			}
		}
	} else {
		// Update snack
		g.snack.y += g.snack.speed

		// Collision check with pet basket
		dx := g.snack.x - g.petX
		if dx < 0 {
			dx = -dx
		}
		if dx < 18 && g.snack.y >= 186 && g.snack.y <= 204 {
			g.CatchSnack()
		} else if g.snack.y > 212 {
			g.MissSnack()
		}
	}

	if g.caught >= g.targetWins {
		g.done = true
		g.won = true
		return true, true, g.calculateScore()
	}
	if g.misses >= g.maxMisses {
		g.done = true
		g.won = false
		return true, false, g.calculateScore()
	}

	return false, false, g.calculateScore()
}

func (g *SnackCatch) calculateScore() int {
	score := g.caught * 100
	if g.won {
		score += (g.maxMisses - g.misses) * 50
	}
	return score
}

func (g *SnackCatch) Draw(screen *ebiten.Image) {
	// LCD Area: x in [28, 212], y in [48, 226]

	// 1. Title and HUD
	ebitenutil.DebugPrintAt(screen, "SNACK CATCH", 82, 52)
	statusText := fmt.Sprintf("Catch: %d/%d  Miss: %d/%d", g.caught, g.targetWins, g.misses, g.maxMisses)
	ebitenutil.DebugPrintAt(screen, statusText, 44, 68)

	// Ground line
	groundClr := color.RGBA{R: 50, G: 55, B: 75, A: 255}
	vector.StrokeLine(screen, 32, 210, 208, 210, 2, groundClr, true)

	// 2. Falling snack
	if g.snack.active {
		snackClr := hexToRGBA(g.pet.AccentColor, 255)
		if snackClr.R == 0 && snackClr.G == 0 && snackClr.B == 0 {
			snackClr = color.RGBA{R: 255, G: 180, B: 100, A: 255}
		}
		vector.DrawFilledCircle(screen, g.snack.x, g.snack.y, 6, snackClr, true)
		vector.DrawFilledCircle(screen, g.snack.x+1, g.snack.y-1, 2, color.RGBA{R: 255, G: 255, B: 255, A: 200}, true)
	}

	// 3. Pet and basket
	petClr := hexToRGBA(g.pet.PrimaryColor, 255)
	if petClr.R == 0 && petClr.G == 0 && petClr.B == 0 {
		petClr = color.RGBA{R: 250, G: 240, B: 230, A: 255}
	}
	// Pet body
	vector.DrawFilledCircle(screen, g.petX, g.petY, 12, petClr, true)
	// Pet eye dots
	vector.DrawFilledCircle(screen, g.petX-4, g.petY-2, 2, color.RGBA{R: 30, G: 30, B: 40, A: 255}, true)
	vector.DrawFilledCircle(screen, g.petX+4, g.petY-2, 2, color.RGBA{R: 30, G: 30, B: 40, A: 255}, true)

	// Basket bowl on pet's head/hands
	basketClr := hexToRGBA(g.pet.AccentColor, 255)
	drawRoundedRect(screen, g.petX-14, g.petY-12, 28, 6, 2, basketClr)
}
