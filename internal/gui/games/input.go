package games

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Button is one of the device's three on-screen face buttons.
type Button int

const (
	ButtonNone Button = iota
	ButtonA           // Left / action
	ButtonB           // Select / jump
	ButtonC           // Right / cancel
)

// buttonKeys lists the keys each face button stands in for, so every game
// that listens for Left/Space/Right or 1/2/3 also responds to clicks.
var buttonKeys = map[Button][]ebiten.Key{
	ButtonA: {ebiten.KeyArrowLeft, ebiten.KeyDigit1},
	ButtonB: {ebiten.KeySpace, ebiten.KeyDigit2},
	ButtonC: {ebiten.KeyArrowRight, ebiten.KeyDigit3},
}

// virtual is the face button state for the current tick, set by the app before
// calling MiniGame.Update. Updates run on one goroutine, so no locking is needed.
var virtual struct {
	button      Button
	justPressed bool
}

// SetVirtualButton records which face button is held this tick and whether it was just pressed.
// Pass ButtonNone when no button is held.
func SetVirtualButton(b Button, justPressed bool) {
	virtual.button = b
	virtual.justPressed = justPressed && b != ButtonNone
}

func virtualHas(k ebiten.Key) bool {
	for _, bk := range buttonKeys[virtual.button] {
		if bk == k {
			return true
		}
	}
	return false
}

// justPressed reports a key press from the keyboard or a just-clicked face button.
func justPressed(k ebiten.Key) bool {
	return inpututil.IsKeyJustPressed(k) || (virtual.justPressed && virtualHas(k))
}

// pressed reports a key held on the keyboard or via a held face button.
func pressed(k ebiten.Key) bool {
	return ebiten.IsKeyPressed(k) || virtualHas(k)
}
