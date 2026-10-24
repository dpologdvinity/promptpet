package games

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/dpologdvinity/promptpet/internal/companion"
)

// WhackABug is a reflex mini-game with 3 holes.
// Pests pop up to be whacked with keys 1, 2, 3. Ladybugs must be avoided.
// Whack 6 pests to win before 3 misses/penalties.
type WhackABug struct {
	pet          companion.PetDef
	targetHits   int
	hits         int
	misses       int
	maxMisses    int
	currentHole  int // 1, 2, 3 or 0 (none)
	isLadybug    bool
	bugTimer     int
	spawnTimer   int
	whackedHole  int
	whackTimer   int
	ticks        int64
	done         bool
	won          bool
}

// NewWhackABug creates a new WhackABug game.
func NewWhackABug() *WhackABug {
	g := &WhackABug{
		targetHits: 6,
		maxMisses:  3,
		spawnTimer: 15,
	}
	return g
}

func (g *WhackABug) Name() string {
	return "Whack-a-Bug"
}

func (g *WhackABug) Init(pet companion.PetDef) {
	g.pet = pet
	g.targetHits = 6
	g.hits = 0
	g.misses = 0
	g.maxMisses = 3
	g.currentHole = 0
	g.isLadybug = false
	g.bugTimer = 0
	g.spawnTimer = 15
	g.whackedHole = 0
	g.whackTimer = 0
	g.ticks = 0
	g.done = false
	g.won = false
}

// HitPest registers a pest whack directly.
func (g *WhackABug) HitPest() {
	if g.done {
		return
	}
	g.hits++
	g.currentHole = 0
	g.bugTimer = 0
	g.spawnTimer = 15
	if g.hits >= g.targetHits {
		g.done = true
		g.won = true
	}
}

// HitLadybug registers a ladybug strike penalty directly.
func (g *WhackABug) HitLadybug() {
	if g.done {
		return
	}
	g.misses++
	g.currentHole = 0
	g.bugTimer = 0
	g.spawnTimer = 15
	if g.misses >= g.maxMisses {
		g.done = true
		g.won = false
	}
}

// MissPest registers an escaped pest directly.
func (g *WhackABug) MissPest() {
	if g.done {
		return
	}
	g.misses++
	g.currentHole = 0
	g.bugTimer = 0
	g.spawnTimer = 15
	if g.misses >= g.maxMisses {
		g.done = true
		g.won = false
	}
}

// SpawnBug places a bug in the designated hole (1, 2, or 3) manually for testing.
func (g *WhackABug) SpawnBug(hole int, isLadybug bool) {
	g.currentHole = hole
	g.isLadybug = isLadybug
	g.bugTimer = 45
}

// WhackHole simulates a player whack on hole 1, 2, or 3.
func (g *WhackABug) WhackHole(hole int) {
	if g.done {
		return
	}
	g.whackedHole = hole
	g.whackTimer = 10

	if g.currentHole == hole {
		if g.isLadybug {
			g.HitLadybug()
		} else {
			g.HitPest()
		}
	}
}

func (g *WhackABug) Update() (done bool, won bool, score int) {
	if g.done {
		return true, g.won, g.calculateScore()
	}

	if g.hits >= g.targetHits {
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

	if g.whackTimer > 0 {
		g.whackTimer--
		if g.whackTimer == 0 {
			g.whackedHole = 0
		}
	}

	// 1. Controls (Keys 1, 2, 3)
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit1) || inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.WhackHole(1)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit2) || inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.WhackHole(2)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit3) || inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.WhackHole(3)
	}

	if g.done {
		return true, g.won, g.calculateScore()
	}

	// 2. Bug lifecycle & Spawning
	if g.currentHole == 0 {
		g.spawnTimer--
		if g.spawnTimer <= 0 {
			g.currentHole = 1 + rand.Intn(3)
			// 25% chance of ladybug (obstacle to avoid)
			g.isLadybug = rand.Float32() < 0.25
			g.bugTimer = 42
		}
	} else {
		g.bugTimer--
		if g.bugTimer <= 0 {
			// Bug retreated
			if !g.isLadybug {
				g.MissPest()
			} else {
				// Ladybug escaped safely (player avoided hitting it)
				g.currentHole = 0
				g.spawnTimer = 15
			}
		}
	}

	if g.hits >= g.targetHits {
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

func (g *WhackABug) calculateScore() int {
	score := g.hits * 150 - g.misses * 50
	if score < 0 {
		score = 0
	}
	if g.won {
		score += 500
	}
	return score
}

func (g *WhackABug) Draw(screen *ebiten.Image) {
	// LCD Area: x in [28, 212], y in [48, 226]
	ebitenutil.DebugPrintAt(screen, "WHACK-A-BUG", 82, 52)
	statusText := fmt.Sprintf("Pests: %d/%d  Miss: %d/%d", g.hits, g.targetHits, g.misses, g.maxMisses)
	ebitenutil.DebugPrintAt(screen, statusText, 48, 68)

	holeX := [3]float32{65, 120, 175}
	holeY := float32(165)

	// Draw 3 hole mounds
	dirtRimClr := color.RGBA{R: 70, G: 55, B: 45, A: 255}
	holeInsideClr := color.RGBA{R: 25, G: 22, B: 20, A: 255}

	for i := 0; i < 3; i++ {
		hx := holeX[i]
		vector.DrawFilledCircle(screen, hx, holeY+5, 18, dirtRimClr, true)
		vector.DrawFilledCircle(screen, hx, holeY+5, 14, holeInsideClr, true)

		label := fmt.Sprintf("%d", i+1)
		ebitenutil.DebugPrintAt(screen, label, int(hx)-4, int(holeY)+22)
	}

	// Draw current bug if visible
	if g.currentHole >= 1 && g.currentHole <= 3 {
		bx := holeX[g.currentHole-1]
		by := holeY - 6

		if g.isLadybug {
			// Ladybug: Red body with black spots
			bodyClr := color.RGBA{R: 235, G: 50, B: 50, A: 255}
			vector.DrawFilledCircle(screen, bx, by, 11, bodyClr, true)
			// Head
			vector.DrawFilledCircle(screen, bx, by-9, 5, color.RGBA{R: 20, G: 20, B: 20, A: 255}, true)
			// Spots
			vector.DrawFilledCircle(screen, bx-4, by-2, 2.5, color.RGBA{R: 20, G: 20, B: 20, A: 255}, true)
			vector.DrawFilledCircle(screen, bx+4, by-2, 2.5, color.RGBA{R: 20, G: 20, B: 20, A: 255}, true)
			vector.DrawFilledCircle(screen, bx, by+4, 2, color.RGBA{R: 20, G: 20, B: 20, A: 255}, true)
			// Center dividing line
			vector.StrokeLine(screen, bx, by-8, bx, by+10, 1.5, color.RGBA{R: 20, G: 20, B: 20, A: 255}, true)
		} else {
			// Pest: Green/Brown beetle bug with antennas
			pestClr := color.RGBA{R: 120, G: 190, B: 60, A: 255}
			vector.DrawFilledCircle(screen, bx, by, 11, pestClr, true)
			// Eyes
			vector.DrawFilledCircle(screen, bx-4, by-4, 3, color.RGBA{R: 255, G: 255, B: 255, A: 255}, true)
			vector.DrawFilledCircle(screen, bx+4, by-4, 3, color.RGBA{R: 255, G: 255, B: 255, A: 255}, true)
			vector.DrawFilledCircle(screen, bx-4, by-4, 1.5, color.RGBA{R: 20, G: 20, B: 20, A: 255}, true)
			vector.DrawFilledCircle(screen, bx+4, by-4, 1.5, color.RGBA{R: 20, G: 20, B: 20, A: 255}, true)
			// Antennae
			vector.StrokeLine(screen, bx-3, by-9, bx-7, by-15, 2, pestClr, true)
			vector.StrokeLine(screen, bx+3, by-9, bx+7, by-15, 2, pestClr, true)
		}
	}

	// Whack impact animation
	if g.whackedHole >= 1 && g.whackedHole <= 3 {
		wx := holeX[g.whackedHole-1]
		wy := holeY - 10
		sparkClr := color.RGBA{R: 255, G: 230, B: 80, A: 230}
		vector.DrawFilledCircle(screen, wx, wy, 16, sparkClr, true)
		vector.DrawFilledCircle(screen, wx, wy, 8, color.RGBA{R: 255, G: 255, B: 255, A: 255}, true)
	}

	// Bottom warning
	ebitenutil.DebugPrintAt(screen, "Whack PEST! Avoid LADYBUG!", 36, 206)
}
