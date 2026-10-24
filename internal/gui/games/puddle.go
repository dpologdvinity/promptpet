package games

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/dpologdvinity/promptpet/internal/companion"
)

type obstacle struct {
	dist    float32
	isRock  bool
	hit     bool
}

// PuddleLeap is a side-scrolling runner where pet leaps over obstacles toward the finish flag.
type PuddleLeap struct {
	pet            companion.PetDef
	petX           float32
	petY           float32
	vy             float32
	groundY        float32
	gravity        float32
	jumpVelocity   float32
	progress       float32
	finishDist     float32
	scrollSpeed    float32
	trips          int
	maxTrips       int
	obstacles      []obstacle
	ticks          int64
	done           bool
	won            bool
}

// NewPuddleLeap creates a new PuddleLeap game.
func NewPuddleLeap() *PuddleLeap {
	g := &PuddleLeap{
		petX:         64,
		groundY:      196,
		gravity:      0.5,
		jumpVelocity: -7.0,
		finishDist:   500,
		scrollSpeed:  2.5,
		maxTrips:     3,
	}
	g.resetObstacles()
	return g
}

func (g *PuddleLeap) resetObstacles() {
	g.obstacles = []obstacle{
		{dist: 110, isRock: false},
		{dist: 190, isRock: true},
		{dist: 280, isRock: false},
		{dist: 360, isRock: true},
		{dist: 440, isRock: false},
	}
}

func (g *PuddleLeap) Name() string {
	return "Puddle Leap"
}

func (g *PuddleLeap) Init(pet companion.PetDef) {
	g.pet = pet
	g.petX = 64
	g.groundY = 196
	g.petY = g.groundY
	g.vy = 0
	g.progress = 0
	g.finishDist = 500
	g.scrollSpeed = 2.5
	g.trips = 0
	g.maxTrips = 3
	g.ticks = 0
	g.done = false
	g.won = false
	g.resetObstacles()
}

// SetProgress sets the current distance progress (for testing).
func (g *PuddleLeap) SetProgress(dist float32) {
	g.progress = dist
}

// FinishDistance returns the total distance required to reach the flag.
func (g *PuddleLeap) FinishDistance() float32 {
	return g.finishDist
}

// Trip manually increments trip count (for testing).
func (g *PuddleLeap) Trip() {
	g.trips++
}

func (g *PuddleLeap) Update() (done bool, won bool, score int) {
	if g.done {
		return true, g.won, g.calculateScore()
	}

	if g.trips >= g.maxTrips {
		g.done = true
		g.won = false
		return true, false, g.calculateScore()
	}

	if g.progress >= g.finishDist {
		g.done = true
		g.won = true
		return true, true, g.calculateScore()
	}

	g.ticks++

	// 1. Jump input
	onGround := g.petY >= g.groundY
	if onGround {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
			inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) ||
			inpututil.IsKeyJustPressed(ebiten.KeyW) {
			g.vy = g.jumpVelocity
		}
	}

	// 2. Physics
	g.petY += g.vy
	if g.petY < g.groundY {
		g.vy += g.gravity
	} else {
		g.petY = g.groundY
		g.vy = 0
	}

	// 3. Scroll progress
	g.progress += g.scrollSpeed

	// 4. Collision check with obstacles
	for i := range g.obstacles {
		obs := &g.obstacles[i]
		if obs.hit {
			continue
		}
		obsScreenX := (obs.dist - g.progress) + g.petX
		dx := obsScreenX - g.petX
		if dx < 0 {
			dx = -dx
		}
		// Pet bounding box check
		if dx < 14 && g.petY > g.groundY-14 {
			obs.hit = true
			g.trips++
		}
	}

	// 5. Check end conditions
	if g.trips >= g.maxTrips {
		g.done = true
		g.won = false
		return true, false, g.calculateScore()
	}

	if g.progress >= g.finishDist {
		g.done = true
		g.won = true
		return true, true, g.calculateScore()
	}

	return false, false, g.calculateScore()
}

func (g *PuddleLeap) calculateScore() int {
	if g.won {
		score := 1000 - (g.trips * 200)
		if score < 100 {
			score = 100
		}
		return score
	}
	return int(g.progress)
}

func (g *PuddleLeap) Draw(screen *ebiten.Image) {
	// HUD
	ebitenutil.DebugPrintAt(screen, "PUDDLE LEAP", 84, 52)
	hud := fmt.Sprintf("Dist: %d/%d  Trips: %d/%d", int(g.progress), int(g.finishDist), g.trips, g.maxTrips)
	ebitenutil.DebugPrintAt(screen, hud, 44, 68)

	// Ground line
	groundLineY := float32(208)
	groundClr := color.RGBA{R: 50, G: 55, B: 75, A: 255}
	vector.StrokeLine(screen, 32, groundLineY, 208, groundLineY, 2, groundClr, true)

	// Obstacles
	for _, obs := range g.obstacles {
		screenX := (obs.dist - g.progress) + g.petX
		if screenX < 26 || screenX > 214 {
			continue
		}
		if obs.isRock {
			// Gray rock
			rockClr := color.RGBA{R: 120, G: 125, B: 140, A: 255}
			vector.DrawFilledCircle(screen, screenX, groundLineY-4, 7, rockClr, true)
		} else {
			// Blue puddle
			puddleClr := color.RGBA{R: 90, G: 160, B: 240, A: 220}
			vector.DrawFilledRect(screen, screenX-8, groundLineY-2, 16, 4, puddleClr, true)
		}
	}

	// Finish flag
	flagScreenX := (g.finishDist - g.progress) + g.petX
	if flagScreenX >= 28 && flagScreenX <= 212 {
		vector.StrokeLine(screen, flagScreenX, groundLineY, flagScreenX, groundLineY-28, 2, color.RGBA{R: 240, G: 240, B: 250, A: 255}, true)
		flagClr := color.RGBA{R: 243, G: 139, B: 168, A: 255}
		vector.DrawFilledRect(screen, flagScreenX, groundLineY-28, 12, 8, flagClr, true)
	}

	// Animated Pet
	petClr := hexToRGBA(g.pet.PrimaryColor, 255)
	if petClr.R == 0 && petClr.G == 0 && petClr.B == 0 {
		petClr = color.RGBA{R: 250, G: 240, B: 230, A: 255}
	}

	runBob := float32(0)
	if g.petY >= g.groundY {
		runBob = float32(math.Sin(float64(g.ticks)*0.4) * 2.0)
	}
	drawPetY := g.petY + runBob

	// Body
	vector.DrawFilledCircle(screen, g.petX, drawPetY, 11, petClr, true)
	// Eyes
	vector.DrawFilledCircle(screen, g.petX+4, drawPetY-2, 2, color.RGBA{R: 30, G: 30, B: 40, A: 255}, true)
}
