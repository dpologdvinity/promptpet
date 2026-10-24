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

type rhythmBeat struct {
	lane   int
	y      float32
	speed  float32
	active bool
}

// RhythmDance is a musical mini-game where arrows fall down; tap keys when in target zone.
// Hit 8 beats to win before 4 misses.
type RhythmDance struct {
	pet           companion.PetDef
	targetHits    int
	maxMisses     int
	hits          int
	misses        int
	targetY       float32
	laneX         [4]float32
	beats         []rhythmBeat
	spawnTimer    int
	feedbackTimer int
	ticks         int64
	done          bool
	won           bool
}

// NewRhythmDance creates a new RhythmDance game.
func NewRhythmDance() *RhythmDance {
	return &RhythmDance{
		targetHits: 8,
		maxMisses:  4,
		targetY:    185,
		laneX:      [4]float32{60, 100, 140, 180},
		spawnTimer: 15,
	}
}

func (g *RhythmDance) Name() string {
	return "Rhythm Dance"
}

func (g *RhythmDance) Init(pet companion.PetDef) {
	g.pet = pet
	g.targetHits = 8
	g.maxMisses = 4
	g.hits = 0
	g.misses = 0
	g.targetY = 185
	g.laneX = [4]float32{60, 100, 140, 180}
	g.beats = nil
	g.spawnTimer = 15
	g.feedbackTimer = 0
	g.ticks = 0
	g.done = false
	g.won = false
}

// HitBeat manually registers a hit beat (for testing).
func (g *RhythmDance) HitBeat() {
	g.hits++
	if g.hits >= g.targetHits {
		g.done = true
		g.won = true
	}
}

// MissBeat manually registers a missed beat (for testing).
func (g *RhythmDance) MissBeat() {
	g.misses++
	if g.misses >= g.maxMisses {
		g.done = true
		g.won = false
	}
}

func (g *RhythmDance) Update() (done bool, won bool, score int) {
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
	if g.feedbackTimer > 0 {
		g.feedbackTimer--
	}

	// 1. Spawning
	g.spawnTimer--
	if g.spawnTimer <= 0 {
		g.spawnTimer = 22
		g.beats = append(g.beats, rhythmBeat{
			lane:   rand.Intn(4),
			y:      55,
			speed:  2.2,
			active: true,
		})
	}

	// 2. Input handling
	inputLanes := [4]bool{
		inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA),
		inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW),
		inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS),
		inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD),
	}

	for laneIdx, pressed := range inputLanes {
		if !pressed {
			continue
		}
		// Look for closest active beat in this lane
		closestIdx := -1
		minDist := float32(9999)
		for i := range g.beats {
			if !g.beats[i].active || g.beats[i].lane != laneIdx {
				continue
			}
			dist := g.beats[i].y - g.targetY
			if dist < 0 {
				dist = -dist
			}
			if dist < minDist {
				minDist = dist
				closestIdx = i
			}
		}

		if closestIdx != -1 && minDist <= 18 {
			g.beats[closestIdx].active = false
			g.HitBeat()
			g.feedbackTimer = 12
		}
	}

	// 3. Move beats and check misses
	for i := range g.beats {
		if !g.beats[i].active {
			continue
		}
		g.beats[i].y += g.beats[i].speed
		if g.beats[i].y > 206 {
			g.beats[i].active = false
			g.MissBeat()
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

func (g *RhythmDance) calculateScore() int {
	score := g.hits * 100
	if g.won {
		score += (g.maxMisses - g.misses) * 50
	}
	return score
}

func (g *RhythmDance) Draw(screen *ebiten.Image) {
	// HUD
	ebitenutil.DebugPrintAt(screen, "RHYTHM DANCE", 82, 52)
	hud := fmt.Sprintf("Hits: %d/%d  Miss: %d/%d", g.hits, g.targetHits, g.misses, g.maxMisses)
	ebitenutil.DebugPrintAt(screen, hud, 44, 68)

	// Target line
	targetLineClr := color.RGBA{R: 60, G: 65, B: 85, A: 255}
	vector.StrokeLine(screen, 36, g.targetY, 204, g.targetY, 1.5, targetLineClr, true)

	// Lane target receptors
	laneLabels := []string{"<", "^", "v", ">"}
	for i, lx := range g.laneX {
		receptorClr := color.RGBA{R: 70, G: 75, B: 100, A: 255}
		vector.StrokeCircle(screen, lx, g.targetY, 10, 1.5, receptorClr, true)
		ebitenutil.DebugPrintAt(screen, laneLabels[i], int(lx-3), int(g.targetY-7))
	}

	// Falling beats
	beatClr := hexToRGBA(g.pet.AccentColor, 255)
	if beatClr.R == 0 && beatClr.G == 0 && beatClr.B == 0 {
		beatClr = color.RGBA{R: 245, G: 194, B: 231, A: 255}
	}

	for _, b := range g.beats {
		if !b.active {
			continue
		}
		lx := g.laneX[b.lane]
		vector.DrawFilledCircle(screen, lx, b.y, 8, beatClr, true)
		innerClr := color.RGBA{R: 255, G: 255, B: 255, A: 200}
		vector.DrawFilledCircle(screen, lx, b.y, 3, innerClr, true)
		ebitenutil.DebugPrintAt(screen, laneLabels[b.lane], int(lx-3), int(b.y-7))
	}

	// Hit feedback
	if g.feedbackTimer > 0 {
		ebitenutil.DebugPrintAt(screen, "PERFECT!", 92, 160)
	}

	// Tiny dancing mascot in corner
	petClr := hexToRGBA(g.pet.PrimaryColor, 255)
	if petClr.R == 0 && petClr.G == 0 && petClr.B == 0 {
		petClr = color.RGBA{R: 250, G: 240, B: 230, A: 255}
	}
	danceWiggle := float32(math.Sin(float64(g.ticks)*0.5) * 4.0)
	vector.DrawFilledCircle(screen, 120+danceWiggle, 110, 14, petClr, true)
	vector.DrawFilledCircle(screen, 120+danceWiggle-4, 108, 2, color.RGBA{R: 30, G: 30, B: 40, A: 255}, true)
	vector.DrawFilledCircle(screen, 120+danceWiggle+4, 108, 2, color.RGBA{R: 30, G: 30, B: 40, A: 255}, true)
}
