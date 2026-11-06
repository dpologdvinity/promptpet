package gui

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/dpologdvinity/promptpet/internal/companion"
)

var (
	allMoods    = []companion.RhythmMood{"", companion.MoodEnergetic, companion.MoodResting, companion.MoodCheer, companion.MoodSympathy, companion.MoodFanfare}
	allWellness = []companion.WellnessType{companion.WellnessNone, companion.WellnessWater, companion.WellnessStretch, companion.WellnessLookAway}
)

// TestDrawPetAllScenes renders every pet in every scene combination at a
// morning, afternoon and night hour, across several animation ticks.
func TestDrawPetAllScenes(t *testing.T) {
	screen := ebiten.NewImage(240, 340)
	for _, pet := range companion.GetAllPets() {
		for _, mood := range allMoods {
			for _, w := range allWellness {
				for _, hour := range []int{8, 14, 23} {
					for _, ticks := range []int64{0, 37, 1001} {
						DrawPet(screen, pet, ticks, Scene{Mood: mood, Wellness: w, Hour: hour})
					}
				}
			}
		}
	}
}

// TestSceneHours checks the day/night prop boundaries from DESIGN.md §6.
// Prop pixels can't be asserted: ebiten's ReadPixels/At panic before
// RunGame starts, and RunGame needs a window, so tests stay headless.
func TestSceneHours(t *testing.T) {
	for hour := 0; hour < 24; hour++ {
		wantNight := hour >= 22 || hour <= 4
		wantMorning := hour >= 6 && hour <= 11
		if got := isNight(hour); got != wantNight {
			t.Errorf("isNight(%d) = %v, want %v", hour, got, wantNight)
		}
		if got := isMorning(hour); got != wantMorning {
			t.Errorf("isMorning(%d) = %v, want %v", hour, got, wantMorning)
		}
	}
}
