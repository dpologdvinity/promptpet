package gui

import (
	"errors"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/dpologdvinity/promptpet/internal/companion"
	"github.com/dpologdvinity/promptpet/internal/event"
)

// ErrHeadless is returned when attempting to run the GUI in an environment without a display server.
var ErrHeadless = errors.New("gui: no display available (headless environment)")

// Config holds initialization paths for companion state and event IPC.
type Config struct {
	StatePath string
	EventPath string
}

// App implements ebiten.Game to render the handheld virtual pet companion.
type App struct {
	mu            sync.Mutex
	cfg           Config
	state         *companion.State
	pets          []companion.PetDef
	petIndex      int
	activePet     companion.PetDef
	ticks         int64
	speech        string
	mood          companion.RhythmMood
	lastReadTs    int64
	lastWellness  time.Time
	idleThrottled bool
	throttleSleep time.Duration
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

	app := &App{
		cfg:           cfg,
		state:         state,
		pets:          allPets,
		petIndex:      petIdx,
		activePet:     active,
		speech:        initialSpeech,
		mood:          companion.MoodResting,
		lastReadTs:    time.Now().Unix() - 60,
		lastWellness:  time.Now(),
		idleThrottled: true,
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

// Update handles input, event polling, rhythm evaluation, and throttling.
func (a *App) Update() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.ticks++

	// Tab key cycles active pet
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		a.cyclePetLocked()
	}

	now := time.Now()

	// Inter-process event polling: read newly arrived shell events
	events, err := event.ReadEventsSince(a.cfg.EventPath, a.lastReadTs)
	if err == nil && len(events) > 0 {
		mood, _ := a.state.EvaluateRhythm(events, now)
		a.mood = mood
		a.speech = a.activePet.SpeechFor(string(mood))
		for _, ev := range events {
			if ev.Timestamp > a.lastReadTs {
				a.lastReadTs = ev.Timestamp
			}
		}
	}

	// Wellness check: evaluate hydration, stretch, and eye rest intervals
	if now.Sub(a.lastWellness) >= 2*time.Second {
		a.lastWellness = now
		reminder := a.state.CheckWellness(now)
		if reminder != companion.WellnessNone {
			a.speech = a.activePet.SpeechFor(string(reminder))
			_ = companion.SaveState(a.cfg.StatePath, a.state)
		}
	}

	// Frame rate sleep throttling: keeps CPU <1.5% when idle
	if a.idleThrottled && a.throttleSleep > 0 {
		time.Sleep(a.throttleSleep)
	}

	return nil
}

// Draw renders the retro handheld device casing, LCD screen, pet, and speech bubble.
func (a *App) Draw(screen *ebiten.Image) {
	a.mu.Lock()
	pet := a.activePet
	ticks := a.ticks
	speech := a.speech
	a.mu.Unlock()

	DrawDevice(screen, pet)
	DrawPet(screen, pet, ticks)
	DrawSpeechBubble(screen, speech, pet)
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
