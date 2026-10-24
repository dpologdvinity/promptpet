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

type simonPhase int

const (
	phaseWatch simonPhase = iota
	phaseInput
)

// MemorySimon is a mini-game where a sequence of 3-5 button flashes is shown,
// and the player must repeat the pattern using number keys 1, 2, 3.
type MemorySimon struct {
	pet           companion.PetDef
	sequence      []int
	playerInput   []int
	phase         simonPhase
	flashIndex    int
	flashTimer    int
	activeLight   int
	feedbackTimer int
	ticks         int64
	done          bool
	won           bool
}

// NewMemorySimon creates a new MemorySimon game.
func NewMemorySimon() *MemorySimon {
	g := &MemorySimon{}
	g.Init(companion.PetDef{})
	return g
}

func (g *MemorySimon) Name() string {
	return "Memory Simon"
}

func (g *MemorySimon) Init(pet companion.PetDef) {
	g.pet = pet
	// Generate default 4-step sequence
	g.sequence = make([]int, 4)
	for i := range g.sequence {
		g.sequence[i] = 1 + rand.Intn(3)
	}
	g.playerInput = nil
	g.phase = phaseWatch
	g.flashIndex = 0
	g.flashTimer = 25
	g.activeLight = 0
	g.feedbackTimer = 0
	g.ticks = 0
	g.done = false
	g.won = false
}

// SetSequence overrides the sequence for deterministic testing or custom length.
func (g *MemorySimon) SetSequence(seq []int) {
	g.sequence = make([]int, len(seq))
	copy(g.sequence, seq)
	g.playerInput = nil
	g.phase = phaseInput // allow immediate input in tests
	g.activeLight = 0
	g.done = false
	g.won = false
}

// InputStep manually inputs a button choice (1, 2, or 3). Returns true if correct.
func (g *MemorySimon) InputStep(btn int) bool {
	if g.done {
		return false
	}
	expectedIdx := len(g.playerInput)
	if expectedIdx >= len(g.sequence) {
		return false
	}

	expected := g.sequence[expectedIdx]
	g.activeLight = btn
	g.feedbackTimer = 10

	if btn == expected {
		g.playerInput = append(g.playerInput, btn)
		if len(g.playerInput) == len(g.sequence) {
			g.done = true
			g.won = true
		}
		return true
	}

	// Wrong input
	g.done = true
	g.won = false
	return false
}

func (g *MemorySimon) Update() (done bool, won bool, score int) {
	if g.done {
		return true, g.won, g.calculateScore()
	}

	g.ticks++

	if g.feedbackTimer > 0 {
		g.feedbackTimer--
		if g.feedbackTimer == 0 && g.phase == phaseInput {
			g.activeLight = 0
		}
	}

	// 1. Showing phase: flash sequence
	if g.phase == phaseWatch {
		g.flashTimer--
		if g.flashTimer <= 0 {
			if g.activeLight == 0 {
				// Light up current step
				if g.flashIndex < len(g.sequence) {
					g.activeLight = g.sequence[g.flashIndex]
					g.flashTimer = 20
				} else {
					// Done showing sequence
					g.phase = phaseInput
					g.activeLight = 0
					g.flashTimer = 0
				}
			} else {
				// Turn off light (gap)
				g.activeLight = 0
				g.flashIndex++
				g.flashTimer = 10
				if g.flashIndex >= len(g.sequence) {
					g.phase = phaseInput
					g.flashTimer = 0
				}
			}
		}
		return false, false, g.calculateScore()
	}

	// 2. Input phase: listen for keys 1, 2, 3
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit1) || inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.InputStep(1)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit2) || inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.InputStep(2)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit3) || inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.InputStep(3)
	}

	if g.done {
		return true, g.won, g.calculateScore()
	}

	return false, false, g.calculateScore()
}

func (g *MemorySimon) calculateScore() int {
	score := len(g.playerInput) * 150
	if g.won {
		score += 500
	}
	return score
}

func (g *MemorySimon) Draw(screen *ebiten.Image) {
	// LCD Area: x in [28, 212], y in [48, 226]
	ebitenutil.DebugPrintAt(screen, "MEMORY SIMON", 78, 52)

	var statusText string
	if g.phase == phaseWatch {
		statusText = "Watch Pattern!"
	} else {
		statusText = fmt.Sprintf("Step %d/%d (Press 1-3)", len(g.playerInput)+1, len(g.sequence))
	}
	ebitenutil.DebugPrintAt(screen, statusText, 52, 68)

	// Pet mascot representation above buttons
	petClr := hexToRGBA(g.pet.PrimaryColor, 255)
	if petClr.R == 0 && petClr.G == 0 && petClr.B == 0 {
		petClr = color.RGBA{R: 250, G: 240, B: 230, A: 255}
	}
	vector.DrawFilledCircle(screen, 120, 102, 14, petClr, true)
	vector.DrawFilledCircle(screen, 115, 100, 2, color.RGBA{R: 30, G: 30, B: 40, A: 255}, true)
	vector.DrawFilledCircle(screen, 125, 100, 2, color.RGBA{R: 30, G: 30, B: 40, A: 255}, true)

	// 3 Button positions
	btnPositions := [3]float32{65, 120, 175}
	btnColors := [3]color.RGBA{
		{R: 78, G: 204, B: 163, A: 255},  // 1: Teal / Green
		{R: 246, G: 201, B: 14, A: 255},  // 2: Amber / Yellow
		{R: 255, G: 107, B: 107, A: 255}, // 3: Coral / Red
	}

	for i := 0; i < 3; i++ {
		bx := btnPositions[i]
		by := float32(158)
		clr := btnColors[i]
		btnNum := i + 1

		isLit := (g.activeLight == btnNum)

		if isLit {
			// Glow halo
			haloClr := color.RGBA{R: clr.R, G: clr.G, B: clr.B, A: 100}
			vector.DrawFilledCircle(screen, bx, by, 22, haloClr, true)
			vector.DrawFilledCircle(screen, bx, by, 16, clr, true)
			vector.DrawFilledCircle(screen, bx, by, 8, color.RGBA{R: 255, G: 255, B: 255, A: 220}, true)
		} else {
			// Dim button
			dimClr := color.RGBA{R: clr.R / 2, G: clr.G / 2, B: clr.B / 2, A: 255}
			vector.DrawFilledCircle(screen, bx, by, 16, dimClr, true)
			vector.StrokeCircle(screen, bx, by, 16, 2, clr, true)
		}

		label := fmt.Sprintf("%d", btnNum)
		ebitenutil.DebugPrintAt(screen, label, int(bx)-4, int(by)+20)
	}
}
