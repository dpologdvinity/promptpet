package games

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/dpologdvinity/promptpet/internal/companion"
)

type balloonObj struct {
	x      float32
	y      float32
	speed  float32
	radius float32
	clr    color.RGBA
	active bool
}

type dartObj struct {
	x      float32
	y      float32
	vx     float32
	vy     float32
	active bool
}

// BalloonPop is an arcade mini-game where the player aims darts to pop rising balloons.
// Pop 5 balloons to win before 4 escape.
type BalloonPop struct {
	pet        companion.PetDef
	targetPops int
	popped     int
	misses     int
	maxMisses  int
	aimAngle   float64 // Radians from horizontal
	originX    float32
	originY    float32
	balloons   []balloonObj
	darts      []dartObj
	spawnTimer int
	ticks      int64
	done       bool
	won        bool
}

// NewBalloonPop creates a new BalloonPop game.
func NewBalloonPop() *BalloonPop {
	g := &BalloonPop{
		targetPops: 5,
		maxMisses:  4,
		aimAngle:   0,
		originX:    44,
		originY:    140,
		spawnTimer: 10,
	}
	return g
}

func (g *BalloonPop) Name() string {
	return "Balloon Pop"
}

func (g *BalloonPop) Init(pet companion.PetDef) {
	g.pet = pet
	g.targetPops = 5
	g.popped = 0
	g.misses = 0
	g.maxMisses = 4
	g.aimAngle = 0
	g.originX = 44
	g.originY = 140
	g.balloons = nil
	g.darts = nil
	g.spawnTimer = 10
	g.ticks = 0
	g.done = false
	g.won = false
}

// SetAimAngle sets the dart launcher angle in radians.
func (g *BalloonPop) SetAimAngle(rad float64) {
	g.aimAngle = rad
}

// ShootDart launches a dart along the current aim angle.
func (g *BalloonPop) ShootDart() {
	if g.done {
		return
	}
	speed := float32(6.0)
	vx := float32(math.Cos(g.aimAngle)) * speed
	vy := float32(math.Sin(g.aimAngle)) * speed

	g.darts = append(g.darts, dartObj{
		x:      g.originX + 12,
		y:      g.originY,
		vx:     vx,
		vy:     vy,
		active: true,
	})
}

// PopBalloon increments popped count directly.
func (g *BalloonPop) PopBalloon() {
	if g.done {
		return
	}
	g.popped++
	if g.popped >= g.targetPops {
		g.done = true
		g.won = true
	}
}

// MissBalloon increments missed count directly.
func (g *BalloonPop) MissBalloon() {
	if g.done {
		return
	}
	g.misses++
	if g.misses >= g.maxMisses {
		g.done = true
		g.won = false
	}
}

func (g *BalloonPop) Update() (done bool, won bool, score int) {
	if g.done {
		return true, g.won, g.calculateScore()
	}

	if g.popped >= g.targetPops {
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
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.aimAngle -= 0.04
		if g.aimAngle < -0.9 {
			g.aimAngle = -0.9
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.aimAngle += 0.04
		if g.aimAngle > 0.9 {
			g.aimAngle = 0.9
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.ShootDart()
	}

	// 2. Spawn balloons
	g.spawnTimer--
	if g.spawnTimer <= 0 {
		g.spawnTimer = 35 + rand.Intn(20)
		palette := []color.RGBA{
			{R: 255, G: 120, B: 150, A: 255}, // Pink
			{R: 100, G: 190, B: 255, A: 255}, // Cyan
			{R: 255, G: 215, B: 80, A: 255},  // Gold
			{R: 180, G: 130, B: 240, A: 255}, // Lavender
			{R: 120, G: 220, B: 150, A: 255}, // Mint
		}
		clr := palette[rand.Intn(len(palette))]
		bx := float32(85 + rand.Intn(105))
		speed := float32(0.9 + rand.Float64()*0.6)

		g.balloons = append(g.balloons, balloonObj{
			x:      bx,
			y:      215,
			speed:  speed,
			radius: 9,
			clr:    clr,
			active: true,
		})
	}

	// 3. Update balloons
	for i := range g.balloons {
		b := &g.balloons[i]
		if !b.active {
			continue
		}
		b.y -= b.speed
		if b.y < 52 {
			b.active = false
			g.MissBalloon()
		}
	}

	// 4. Update darts & collision detection
	for di := range g.darts {
		d := &g.darts[di]
		if !d.active {
			continue
		}
		d.x += d.vx
		d.y += d.vy

		// Out of LCD bounds check (x: [28, 212], y: [48, 226])
		if d.x > 212 || d.y < 48 || d.y > 226 {
			d.active = false
			continue
		}

		// Collision check with balloons
		for bi := range g.balloons {
			b := &g.balloons[bi]
			if !b.active {
				continue
			}
			dx := d.x - b.x
			dy := d.y - b.y
			distSq := dx*dx + dy*dy
			radSq := (b.radius + 3) * (b.radius + 3)

			if distSq <= radSq {
				b.active = false
				d.active = false
				g.PopBalloon()
				break
			}
		}
	}

	if g.popped >= g.targetPops {
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

func (g *BalloonPop) calculateScore() int {
	score := g.popped * 200
	if g.won {
		score += (g.maxMisses - g.misses) * 100 + 400
	}
	return score
}

func (g *BalloonPop) Draw(screen *ebiten.Image) {
	// LCD Area: x in [28, 212], y in [48, 226]
	ebitenutil.DebugPrintAt(screen, "BALLOON POP", 82, 52)
	statusText := fmt.Sprintf("Pop: %d/%d  Miss: %d/%d", g.popped, g.targetPops, g.misses, g.maxMisses)
	ebitenutil.DebugPrintAt(screen, statusText, 48, 68)

	// 1. Draw Balloons
	for _, b := range g.balloons {
		if !b.active {
			continue
		}
		// Balloon body
		vector.DrawFilledCircle(screen, b.x, b.y, b.radius, b.clr, true)
		// Highlight shine
		vector.DrawFilledCircle(screen, b.x-2.5, b.y-2.5, 2, color.RGBA{R: 255, G: 255, B: 255, A: 200}, true)
		// Balloon tie knot
		knotClr := color.RGBA{R: b.clr.R / 2, G: b.clr.G / 2, B: b.clr.B / 2, A: 255}
		vector.DrawFilledCircle(screen, b.x, b.y+b.radius+1, 1.5, knotClr, true)
		// Dangling string
		stringClr := color.RGBA{R: 120, G: 125, B: 140, A: 200}
		vector.StrokeLine(screen, b.x, b.y+b.radius+2, b.x-1, b.y+b.radius+9, 1, stringClr, true)
	}

	// 2. Draw Darts
	dartClr := color.RGBA{R: 255, G: 80, B: 80, A: 255}
	for _, d := range g.darts {
		if !d.active {
			continue
		}
		vector.StrokeLine(screen, d.x-d.vx*1.5, d.y-d.vy*1.5, d.x, d.y, 2, dartClr, true)
		vector.DrawFilledCircle(screen, d.x, d.y, 2, color.RGBA{R: 255, G: 255, B: 255, A: 255}, true)
	}

	// 3. Draw Player Pet and Launcher
	petClr := hexToRGBA(g.pet.PrimaryColor, 255)
	if petClr.R == 0 && petClr.G == 0 && petClr.B == 0 {
		petClr = color.RGBA{R: 250, G: 240, B: 230, A: 255}
	}
	vector.DrawFilledCircle(screen, g.originX, g.originY, 11, petClr, true)
	vector.DrawFilledCircle(screen, g.originX+2, g.originY-2, 1.5, color.RGBA{R: 30, G: 30, B: 40, A: 255}, true)

	// Aim line / barrel
	barrelLen := float32(14)
	endX := g.originX + float32(math.Cos(g.aimAngle))*barrelLen
	endY := g.originY + float32(math.Sin(g.aimAngle))*barrelLen
	barrelClr := color.RGBA{R: 200, G: 190, B: 220, A: 255}
	vector.StrokeLine(screen, g.originX, g.originY, endX, endY, 2.5, barrelClr, true)

	// 4. Instructions
	ebitenutil.DebugPrintAt(screen, "[Up/Dn] Aim  [Space] Fire", 36, 206)
}
