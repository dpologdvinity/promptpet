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

type lookDirection int

const (
	dirCenter lookDirection = iota
	dirLeft
	dirRight
)

// LookAway is a Tamagotchi-style game: guess which direction pet looks (Left/Right).
// 5 quick rounds; win 3+ to win.
type LookAway struct {
	pet          companion.PetDef
	currentRound int
	maxRounds    int
	wins         int
	targetWins   int
	state        int // 0 = waiting for input, 1 = showing round result
	resultTimer  int
	playerGuess  lookDirection
	petLook      lookDirection
	lastMatch    bool
	done         bool
	won          bool
}

// NewLookAway creates a new LookAway game.
func NewLookAway() *LookAway {
	return &LookAway{
		maxRounds:  5,
		targetWins: 3,
	}
}

func (g *LookAway) Name() string {
	return "Look Away"
}

func (g *LookAway) Init(pet companion.PetDef) {
	g.pet = pet
	g.currentRound = 1
	g.maxRounds = 5
	g.wins = 0
	g.targetWins = 3
	g.state = 0
	g.resultTimer = 0
	g.playerGuess = dirCenter
	g.petLook = dirCenter
	g.lastMatch = false
	g.done = false
	g.won = false
}

// SimulateRound directly resolves a round with match (win) or miss (loss) for testing.
func (g *LookAway) SimulateRound(match bool) {
	if g.done {
		return
	}
	if match {
		g.wins++
		g.lastMatch = true
	} else {
		g.lastMatch = false
	}
	g.currentRound++
	if g.currentRound > g.maxRounds {
		g.done = true
		g.won = g.wins >= g.targetWins
	}
}

func (g *LookAway) Update() (done bool, won bool, score int) {
	if g.done {
		return true, g.won, g.calculateScore()
	}

	if g.currentRound > g.maxRounds {
		g.done = true
		g.won = g.wins >= g.targetWins
		return true, g.won, g.calculateScore()
	}

	if g.state == 1 {
		// Showing round result
		g.resultTimer--
		if g.resultTimer <= 0 {
			g.currentRound++
			if g.currentRound > g.maxRounds {
				g.done = true
				g.won = g.wins >= g.targetWins
				return true, g.won, g.calculateScore()
			}
			g.state = 0
			g.petLook = dirCenter
			g.playerGuess = dirCenter
		}
		return false, false, g.calculateScore()
	}

	// State 0: waiting for player input
	guess := dirCenter
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		guess = dirLeft
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		guess = dirRight
	}

	if guess != dirCenter {
		g.playerGuess = guess
		// Pet chooses direction (Left or Right)
		if rand.Intn(2) == 0 {
			g.petLook = dirLeft
		} else {
			g.petLook = dirRight
		}

		if g.playerGuess == g.petLook {
			g.wins++
			g.lastMatch = true
		} else {
			g.lastMatch = false
		}

		g.state = 1
		g.resultTimer = 25 // show for ~25 frames
	}

	return false, false, g.calculateScore()
}

func (g *LookAway) calculateScore() int {
	return g.wins * 200
}

func (g *LookAway) Draw(screen *ebiten.Image) {
	// HUD
	ebitenutil.DebugPrintAt(screen, "LOOK AWAY", 90, 52)
	hud := fmt.Sprintf("Round: %d/%d  Wins: %d/%d", g.displayRound(), g.maxRounds, g.wins, g.targetWins)
	ebitenutil.DebugPrintAt(screen, hud, 44, 68)

	// Pet rendering at center
	petCenterX := float32(120)
	petCenterY := float32(138)
	petRadius := float32(24)

	petClr := hexToRGBA(g.pet.PrimaryColor, 255)
	if petClr.R == 0 && petClr.G == 0 && petClr.B == 0 {
		petClr = color.RGBA{R: 250, G: 240, B: 230, A: 255}
	}
	vector.DrawFilledCircle(screen, petCenterX, petCenterY, petRadius, petClr, true)

	// Blush cheeks
	blushClr := hexToRGBA(g.pet.BlushColor, 200)
	vector.DrawFilledCircle(screen, petCenterX-15, petCenterY+5, 4, blushClr, true)
	vector.DrawFilledCircle(screen, petCenterX+15, petCenterY+5, 4, blushClr, true)

	// Eye offset based on petLook
	eyeOffsetX := float32(0)
	if g.petLook == dirLeft {
		eyeOffsetX = -7
	} else if g.petLook == dirRight {
		eyeOffsetX = 7
	}

	eyeClr := color.RGBA{R: 30, G: 30, B: 40, A: 255}
	vector.DrawFilledCircle(screen, petCenterX-9+eyeOffsetX, petCenterY-2, 3.5, eyeClr, true)
	vector.DrawFilledCircle(screen, petCenterX+9+eyeOffsetX, petCenterY-2, 3.5, eyeClr, true)
	// Glints
	vector.DrawFilledCircle(screen, petCenterX-8+eyeOffsetX, petCenterY-3, 1.2, color.RGBA{R: 255, G: 255, B: 255, A: 255}, true)
	vector.DrawFilledCircle(screen, petCenterX+10+eyeOffsetX, petCenterY-3, 1.2, color.RGBA{R: 255, G: 255, B: 255, A: 255}, true)

	// Bottom instructions / Feedback
	if g.state == 1 {
		if g.lastMatch {
			ebitenutil.DebugPrintAt(screen, "*** MATCH! +1 *** ", 62, 190)
		} else {
			ebitenutil.DebugPrintAt(screen, "--- MISMATCH ---", 64, 190)
		}
	} else {
		ebitenutil.DebugPrintAt(screen, "<- Press Left / Right ->", 46, 190)
	}
}

func (g *LookAway) displayRound() int {
	if g.currentRound > g.maxRounds {
		return g.maxRounds
	}
	return g.currentRound
}
