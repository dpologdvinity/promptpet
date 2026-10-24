package games

import (
	"fmt"
	"image/color"
	"math/rand"
	"strings"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/dpologdvinity/promptpet/internal/companion"
)

// WordSnack is an educational/typing mini-game where cute 3-4 letter words float
// across the LCD screen and the player types them using keyboard letter keys.
// Spell 3 words to win before 3 misses.
type WordSnack struct {
	pet            companion.PetDef
	targetWords    int
	wordsCompleted int
	misses         int
	maxMisses      int
	currentWord    string
	typedIndex     int
	wordX          float32
	wordY          float32
	wordSpeed      float32
	wordList       []string
	ticks          int64
	done           bool
	won            bool
}

// NewWordSnack creates a new WordSnack game.
func NewWordSnack() *WordSnack {
	g := &WordSnack{
		targetWords: 3,
		maxMisses:   3,
		wordSpeed:   0.45,
		wordList: []string{
			"TEA", "JOY", "SUN", "PET", "DOG", "CAT",
			"PIE", "GEM", "BEE", "HUG", "STAR", "FISH",
		},
	}
	g.Init(companion.PetDef{})
	return g
}

func (g *WordSnack) Name() string {
	return "Word Snack"
}

func (g *WordSnack) Init(pet companion.PetDef) {
	g.pet = pet
	g.targetWords = 3
	g.wordsCompleted = 0
	g.misses = 0
	g.maxMisses = 3
	g.wordSpeed = 0.45
	if len(g.wordList) == 0 {
		g.wordList = []string{"TEA", "JOY", "SUN", "PET", "DOG", "CAT"}
	}
	g.ticks = 0
	g.done = false
	g.won = false
	g.spawnNextWord()
}

func (g *WordSnack) spawnNextWord() {
	idx := rand.Intn(len(g.wordList))
	next := g.wordList[idx]
	if next == g.currentWord && len(g.wordList) > 1 {
		idx = (idx + 1) % len(g.wordList)
		next = g.wordList[idx]
	}
	g.currentWord = next
	g.typedIndex = 0
	g.wordX = 165
	g.wordY = 135
}

// SetCurrentWord sets the current word directly for tests or custom word sequences.
func (g *WordSnack) SetCurrentWord(w string) {
	g.currentWord = strings.ToUpper(w)
	g.typedIndex = 0
	g.wordX = 165
	g.wordY = 135
}

// TypeLetter simulates or inputs a character typed by the user.
func (g *WordSnack) TypeLetter(r rune) bool {
	if g.done {
		return false
	}
	rUpper := unicode.ToUpper(r)
	if g.typedIndex < len(g.currentWord) {
		expected := rune(g.currentWord[g.typedIndex])
		if rUpper == expected {
			g.typedIndex++
			if g.typedIndex >= len(g.currentWord) {
				g.CompleteWord()
			}
			return true
		}
	}
	return false
}

// CompleteWord manually marks the current word completed.
func (g *WordSnack) CompleteWord() {
	if g.done {
		return
	}
	g.wordsCompleted++
	if g.wordsCompleted >= g.targetWords {
		g.done = true
		g.won = true
		return
	}
	g.spawnNextWord()
}

// MissWord manually marks the current word missed.
func (g *WordSnack) MissWord() {
	if g.done {
		return
	}
	g.misses++
	if g.misses >= g.maxMisses {
		g.done = true
		g.won = false
		return
	}
	g.spawnNextWord()
}

func (g *WordSnack) Update() (done bool, won bool, score int) {
	if g.done {
		return true, g.won, g.calculateScore()
	}

	if g.wordsCompleted >= g.targetWords {
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

	// 1. Keyboard Input: Check letters A-Z
	for k := ebiten.KeyA; k <= ebiten.KeyZ; k++ {
		if inpututil.IsKeyJustPressed(k) {
			char := rune('A' + (k - ebiten.KeyA))
			g.TypeLetter(char)
			break
		}
	}

	if g.done {
		return true, g.won, g.calculateScore()
	}

	// 2. Drift word across screen from right to left
	g.wordX -= g.wordSpeed
	if g.wordX < 45 {
		// Word slipped past without being typed
		g.MissWord()
	}

	if g.wordsCompleted >= g.targetWords {
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

func (g *WordSnack) calculateScore() int {
	score := g.wordsCompleted * 300
	if g.won {
		score += (g.maxMisses-g.misses)*100 + 400
	}
	return score
}

func (g *WordSnack) Draw(screen *ebiten.Image) {
	// LCD Area: x in [28, 212], y in [48, 226]
	ebitenutil.DebugPrintAt(screen, "WORD SNACK", 84, 52)
	statusText := fmt.Sprintf("Words: %d/%d  Miss: %d/%d", g.wordsCompleted, g.targetWords, g.misses, g.maxMisses)
	ebitenutil.DebugPrintAt(screen, statusText, 48, 68)

	// 1. Pet on the left waiting for snack
	petClr := hexToRGBA(g.pet.PrimaryColor, 255)
	if petClr.R == 0 && petClr.G == 0 && petClr.B == 0 {
		petClr = color.RGBA{R: 250, G: 240, B: 230, A: 255}
	}
	petX := float32(46)
	petY := float32(142)
	vector.DrawFilledCircle(screen, petX, petY, 13, petClr, true)
	// Eyes
	vector.DrawFilledCircle(screen, petX+5, petY-3, 2, color.RGBA{R: 30, G: 30, B: 40, A: 255}, true)
	// Open mouth waiting for snack
	vector.DrawFilledCircle(screen, petX+10, petY+2, 3, color.RGBA{R: 240, G: 120, B: 120, A: 255}, true)

	// 2. Floating snack word bubble
	bubbleW := float32(len(g.currentWord)*12 + 16)
	bubbleH := float32(24)
	bubbleX := g.wordX - bubbleW/2
	bubbleY := g.wordY - bubbleH/2

	bubbleBgClr := color.RGBA{R: 42, G: 46, B: 62, A: 230}
	drawRoundedRect(screen, bubbleX, bubbleY, bubbleW, bubbleH, 6, bubbleBgClr)
	drawRoundedRect(screen, bubbleX, bubbleY, bubbleW, bubbleH, 6, color.RGBA{R: 85, G: 95, B: 130, A: 180})

	// 3. Render word letters
	startX := int(bubbleX) + 8
	for i, ch := range g.currentWord {
		letterStr := string(ch)
		lx := startX + i*12
		ly := int(bubbleY) + 5

		if i < g.typedIndex {
			// Already typed letter: green highlight
			vector.DrawFilledCircle(screen, float32(lx+4), float32(ly+5), 5, color.RGBA{R: 78, G: 204, B: 163, A: 120}, true)
			ebitenutil.DebugPrintAt(screen, letterStr, lx, ly)
		} else if i == g.typedIndex {
			// Current active letter: underline & bright yellow/white
			vector.StrokeLine(screen, float32(lx-1), float32(ly+14), float32(lx+9), float32(ly+14), 1.5, color.RGBA{R: 255, G: 215, B: 80, A: 255}, true)
			ebitenutil.DebugPrintAt(screen, letterStr, lx, ly)
		} else {
			// Remaining letter
			ebitenutil.DebugPrintAt(screen, letterStr, lx, ly)
		}
	}

	// 4. Instructions
	ebitenutil.DebugPrintAt(screen, "Type letters to feed!", 48, 206)
}
