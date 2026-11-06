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
	wellness        companion.WellnessType // reminder currently shown, cleared at wellnessUntil
	wellnessUntil   time.Time
	lastEventOffset int64
	lastEventPoll   time.Time
	lastWellness    time.Time
	idleThrottled   bool
	dragging        bool // left mouse held on the casing (not a button): moving the window
	dragX, dragY    int  // cursor position where the drag started
	activeGame      games.MiniGame
}

// wellnessPropDuration is how long a reminder prop (water glass, stretch, binoculars) stays on screen.
const wellnessPropDuration = 30 * time.Second

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
	a.cyclePetLocked(1)
}

// PressButton handles a click on face button A, B or C in the companion view:
// A and C switch to the previous/next pet, B cycles the casing colour.
// While a game runs, buttons go to the game instead (see Update).
func (a *App) PressButton(b games.Button) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pressButtonLocked(b)
}

func (a *App) pressButtonLocked(b games.Button) {
	switch b {
	case games.ButtonA:
		a.cyclePetLocked(-1)
	case games.ButtonB:
		a.cycleCasingLocked()
	case games.ButtonC:
		a.cyclePetLocked(1)
	}
}

// Casing returns the current device casing colour name.
func (a *App) Casing() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state.Casing
}

func (a *App) cycleCasingLocked() {
	sound.PlayBlip()
	a.speech = "New look: " + a.state.CycleCasing() + "!"
	_ = companion.SaveState(a.cfg.StatePath, a.state)
}

func (a *App) cyclePetLocked(step int) {
	if len(a.pets) == 0 {
		return
	}
	sound.PlayBlip()
	n := len(a.pets)
	a.petIndex = ((a.petIndex+step)%n + n) % n
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

	button, buttonJust := a.pollMouseLocked()

	// 1. If mini-game is active, prioritize game loop and handle Esc early exit
	if a.activeGame != nil {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			a.exitGameLocked()
			a.mu.Unlock()
			return nil
		}

		games.SetVirtualButton(button, buttonJust)
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

	if buttonJust {
		a.pressButtonLocked(button)
	}

	// Tab key cycles active pet
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		a.cyclePetLocked(1)
	}

	// K key cycles the casing colour
	if inpututil.IsKeyJustPressed(ebiten.KeyK) {
		a.cycleCasingLocked()
	}

	// Q quits: the borderless window has no close button
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		a.mu.Unlock()
		return ebiten.Termination
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
				if mood == companion.MoodFanfare && rSpeech != "" {
					a.speech = rSpeech
				} else {
					a.speech = a.activePet.SpeechFor(string(mood))
				}
			}
		}
	}

	if a.wellness != companion.WellnessNone && now.After(a.wellnessUntil) {
		a.wellness = companion.WellnessNone
	}

	// 4. Wellness check: evaluate hydration, stretch, and eye rest intervals
	if now.Sub(a.lastWellness) >= 2*time.Second {
		a.lastWellness = now
		reminder := a.state.CheckWellness(now)
		if reminder != companion.WellnessNone {
			a.speech = a.activePet.SpeechFor(string(reminder))
			a.wellness = reminder
			a.wellnessUntil = now.Add(wellnessPropDuration)
			_ = companion.SaveState(a.cfg.StatePath, a.state)
		}
	}

	a.mu.Unlock()
	return nil
}

// pollMouseLocked returns the face button held by the left mouse button (and whether
// it was just pressed), and drags the borderless window when the press starts elsewhere.
func (a *App) pollMouseLocked() (games.Button, bool) {
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		a.dragging = false
		return games.ButtonNone, false
	}
	mx, my := ebiten.CursorPosition()
	b := buttonAt(float32(mx), float32(my))
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if b != games.ButtonNone {
			return b, true
		}
		a.dragging, a.dragX, a.dragY = true, mx, my
	}
	if a.dragging {
		// The cursor is window-relative, so shifting the window by the cursor's
		// offset from the grab point keeps the grab point under the cursor.
		wx, wy := ebiten.WindowPosition()
		ebiten.SetWindowPosition(wx+mx-a.dragX, wy+my-a.dragY)
		return games.ButtonNone, false
	}
	return b, false
}

// Draw renders the retro handheld device casing, LCD screen, pet, and speech bubble.
func (a *App) Draw(screen *ebiten.Image) {
	a.mu.Lock()
	pet := a.activePet
	ticks := a.ticks
	speech := a.speech
	activeGame := a.activeGame
	mood := a.mood
	wellness := a.wellness
	casing := a.state.Casing
	a.mu.Unlock()

	DrawDevice(screen, pet, casing)
	if activeGame != nil {
		activeGame.Draw(screen)
	} else {
		DrawPet(screen, pet, ticks, Scene{Mood: mood, Wellness: wellness, Hour: time.Now().Hour()})
		DrawSpeechBubble(screen, speech, pet)
	}
}

// Layout returns the fixed virtual screen size of 240x340 px.
func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 240, 340
}

// cornerPosition places the 240x340 window at the bottom-right of a monitor of the given size,
// leaving room for a taskbar or dock.
func cornerPosition(monitorW, monitorH int) (x, y int) {
	const marginX, marginY = 24, 72
	return max(0, monitorW-240-marginX), max(0, monitorH-340-marginY)
}

// Headless reports whether no display server is available for the companion window.
func Headless() bool {
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
	if Headless() {
		return ErrHeadless
	}

	app, err := NewApp(cfg)
	if err != nil {
		return err
	}

	// 20 TPS when idle keeps CPU low; games raise it to 60 (see launchGameLocked).
	ebiten.SetWindowSize(240, 340)
	ebiten.SetWindowTitle("PromptPet")
	// Floating, borderless and pinned to the corner (DESIGN §3); drag the casing to move it.
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	if m := ebiten.Monitor(); m != nil {
		ebiten.SetWindowPosition(cornerPosition(m.Size()))
	}
	ebiten.SetTPS(20)

	return ebiten.RunGame(app)
}
