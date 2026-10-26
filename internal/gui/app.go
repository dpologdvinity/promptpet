package gui

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/dpologdvinity/promptpet/internal/companion"
	"github.com/dpologdvinity/promptpet/internal/event"
	"github.com/dpologdvinity/promptpet/internal/gui/games"
	"github.com/dpologdvinity/promptpet/internal/gui/sound"
)

// ErrHeadless is returned when attempting to run the GUI in an environment without a display server.
var ErrHeadless = errors.New("gui: no display available (headless environment)")

// Config holds initialization paths for companion state and event IPC.
type Config struct {
	StatePath          string
	EventPath          string
	InitialEventOffset *int64
}

// App implements ebiten.Game to render the handheld virtual pet companion.
type App struct {
	mu              sync.Mutex
	cfg             Config
	state           *companion.State
	pets            []companion.PetDef
	petIndex        int
	activePet       companion.PetDef
	ticks           int64
	speech          string
	mood            companion.RhythmMood
	lastEventOffset int64
	lastEventPoll   time.Time
	lastWellness    time.Time
	idleThrottled   bool
	throttleSleep   time.Duration
	activeGame      games.MiniGame
}

// NewApp initializes a new companion application.
func NewApp(cfg Config) (*App, error) {
	if cfg.StatePath == "" {
		cfg.StatePath = companion.DefaultStatePath()
	}
	if cfg.EventPath == "" {
		cfg.EventPath = event.GetEventFilePath()
	}

	state, err := companion.LoadState(cfg.StatePath)
	if err != nil {
		state = companion.NewDefaultState()
	}

	allPets := companion.GetAllPets()
	if len(allPets) == 0 {
		return nil, errors.New("gui: no pets available in roster")
	}

	petIdx := 0
	for i, p := range allPets {
		if string(p.ID) == state.ActivePet {
			petIdx = i
			break
		}
	}

	active := allPets[petIdx]
	initialSpeech := active.SpeechFor(string(companion.MoodResting))

	var initialOffset int64
	if cfg.InitialEventOffset != nil {
		initialOffset = *cfg.InitialEventOffset
	} else if fi, err := os.Stat(cfg.EventPath); err == nil {
		initialOffset = fi.Size()
	}

	app := &App{
		cfg:             cfg,
		state:           state,
		pets:            allPets,
		petIndex:        petIdx,
		activePet:       active,
		speech:          initialSpeech,
		mood:            companion.MoodResting,
		lastEventOffset: initialOffset,
		lastEventPoll:   time.Time{}, // allow immediate first read
		lastWellness:    time.Time{}, // allow immediate first check
		idleThrottled:   true,
	}

	return app, nil
}

// ActivePet returns the currently active pet mascot definition.
func (a *App) ActivePet() companion.PetDef {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.activePet
}

// Speech returns the current dialogue line being spoken by the pet.
func (a *App) Speech() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.speech
}

// State returns a reference to the active companion state.
func (a *App) State() *companion.State {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state
}

// LastEventOffset returns the byte offset in the event stream that has been processed.
func (a *App) LastEventOffset() int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.lastEventOffset
}

// CyclePet cycles to the next mascot in the 8-pet roster and saves state.
func (a *App) CyclePet() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cyclePetLocked()
}

func (a *App) cyclePetLocked() {
	if len(a.pets) == 0 {
		return
	}
	sound.PlayBlip()
	a.petIndex = (a.petIndex + 1) % len(a.pets)
	a.activePet = a.pets[a.petIndex]
	a.state.ActivePet = string(a.activePet.ID)
	_ = companion.SaveState(a.cfg.StatePath, a.state)

	if a.mood != "" {
		a.speech = a.activePet.SpeechFor(string(a.mood))
	} else {
		a.speech = a.activePet.SpeechFor("resting")
	}
}

// ActiveGame returns the currently active mini-game, or nil if in companion view.
func (a *App) ActiveGame() games.MiniGame {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.activeGame
}

// LaunchGameByID launches mini-game 1 through 8.
func (a *App) LaunchGameByID(id int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.launchGameLocked(id)
}

// ExitGame exits the active mini-game back to companion view without losing state.
func (a *App) ExitGame() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.exitGameLocked()
}

func (a *App) launchGameLocked(id int) {
	var g games.MiniGame
	switch id {
	case 1:
		g = games.NewSnackCatch()
	case 2:
		g = games.NewPuddleLeap()
	case 3:
		g = games.NewLookAway()
	case 4:
		g = games.NewRhythmDance()
	case 5:
		g = games.NewMemorySimon()
	case 6:
		g = games.NewWhackABug()
	case 7:
		g = games.NewBalloonPop()
	case 8:
		g = games.NewWordSnack()
	default:
		return
	}
	sound.PlayBlip()
	g.Init(a.activePet)
	a.activeGame = g
	ebiten.SetTPS(60)
}

func (a *App) exitGameLocked() {
	sound.PlayBlip()
	a.activeGame = nil
	ebiten.SetTPS(20)
}

func (a *App) finishGameLocked(won bool, score int) {
	if a.activeGame == nil {
		return
	}
	gameName := a.activeGame.Name()
	if won {
		sound.PlaySuccess()
		a.state.FriendshipHearts++
		a.speech = "You won!"
	} else {
		sound.PlayMiss()
		a.speech = "Good try!"
	}

	if a.state.HighScores == nil {
		a.state.HighScores = make(map[string]int)
	}
	if score > a.state.HighScores[gameName] {
		a.state.HighScores[gameName] = score
	}

	_ = companion.SaveState(a.cfg.StatePath, a.state)
	a.activeGame = nil
	ebiten.SetTPS(20)
}

// Update handles input, event polling, rhythm evaluation, and throttling.
func (a *App) Update() error {
	a.mu.Lock()

	a.ticks++

	// 1. If mini-game is active, prioritize game loop and handle Esc early exit
	if a.activeGame != nil {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			a.exitGameLocked()
			a.mu.Unlock()
			return nil
		}

		done, won, score := a.activeGame.Update()
		if done {
			a.finishGameLocked(won, score)
		}
		a.mu.Unlock()
		return nil
	}

	// 2. Companion view controls:
	// Mini-game launcher keys: 1, 2, 3, 4, 5, 6, 7, 8
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit1) || inpututil.IsKeyJustPressed(ebiten.Key1) {
		a.launchGameLocked(1)
		a.mu.Unlock()
		return nil
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit2) || inpututil.IsKeyJustPressed(ebiten.Key2) {
		a.launchGameLocked(2)
		a.mu.Unlock()
		return nil
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit3) || inpututil.IsKeyJustPressed(ebiten.Key3) {
		a.launchGameLocked(3)
		a.mu.Unlock()
		return nil
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit4) || inpututil.IsKeyJustPressed(ebiten.Key4) {
		a.launchGameLocked(4)
		a.mu.Unlock()
		return nil
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit5) || inpututil.IsKeyJustPressed(ebiten.Key5) {
		a.launchGameLocked(5)
		a.mu.Unlock()
		return nil
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit6) || inpututil.IsKeyJustPressed(ebiten.Key6) {
		a.launchGameLocked(6)
		a.mu.Unlock()
		return nil
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit7) || inpututil.IsKeyJustPressed(ebiten.Key7) {
		a.launchGameLocked(7)
		a.mu.Unlock()
		return nil
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDigit8) || inpututil.IsKeyJustPressed(ebiten.Key8) {
		a.launchGameLocked(8)
		a.mu.Unlock()
		return nil
	}

	// Tab key cycles active pet
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		a.cyclePetLocked()
	}

	// M key toggles sound mute
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		muted := sound.ToggleMute()
		if muted {
			a.speech = "(sound muted)"
		} else {
			a.speech = "(sound on)"
			sound.PlayBlip()
		}
	}

	now := time.Now()

	// 3. Inter-process event polling: throttled to 1s to prevent 20Hz disk I/O
	if now.Sub(a.lastEventPoll) >= 1*time.Second {
		a.lastEventPoll = now
		events, newOffset, err := event.ReadEventsFromOffset(a.cfg.EventPath, a.lastEventOffset)
		if err == nil {
			a.lastEventOffset = newOffset
			if len(events) > 0 {
				mood, rSpeech := a.state.EvaluateRhythm(events, now)
				a.mood = mood
				if rSpeech != "" && (strings.Contains(rSpeech, "commit") || strings.Contains(rSpeech, "High five!")) {
					a.speech = rSpeech
				} else {
					a.speech = a.activePet.SpeechFor(string(mood))
				}
			}
		}
	}

	// 4. Wellness check: evaluate hydration, stretch, and eye rest intervals
	if now.Sub(a.lastWellness) >= 2*time.Second {
		a.lastWellness = now
		reminder := a.state.CheckWellness(now)
		if reminder != companion.WellnessNone {
			a.speech = a.activePet.SpeechFor(string(reminder))
			_ = companion.SaveState(a.cfg.StatePath, a.state)
		}
	}

	a.mu.Unlock()
	return nil
}

// Draw renders the retro handheld device casing, LCD screen, pet, and speech bubble.
func (a *App) Draw(screen *ebiten.Image) {
	a.mu.Lock()
	pet := a.activePet
	ticks := a.ticks
	speech := a.speech
	activeGame := a.activeGame
	a.mu.Unlock()

	DrawDevice(screen, pet)
	if activeGame != nil {
		activeGame.Draw(screen)
	} else {
		DrawPet(screen, pet, ticks)
		DrawSpeechBubble(screen, speech, pet)
	}
}

// Layout returns the fixed virtual screen size of 240x340 px.
func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 240, 340
}

func isHeadless() bool {
	if os.Getenv("PROMPTPET_HEADLESS") == "1" {
		return true
	}
	if runtime.GOOS == "linux" {
		return os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == ""
	}
	return false
}

// Run configures the Ebitengine window and launches the companion application loop.
func Run(cfg Config) error {
	if isHeadless() {
		return ErrHeadless
	}

	app, err := NewApp(cfg)
	if err != nil {
		return err
	}

	// 20 FPS throttling when idle to maintain <1.5% CPU
	app.throttleSleep = 20 * time.Millisecond
	ebiten.SetWindowSize(240, 340)
	ebiten.SetWindowTitle("PromptPet")
	ebiten.SetTPS(20)

	return ebiten.RunGame(app)
}
