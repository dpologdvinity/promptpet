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
	// Every pet sees every mood, reminder, hour and tick phase once (dimensions
	// cycle together), plus one pet gets every mood x reminder pair. A full
	// cross product took minutes under -race in CI.
	screen := ebiten.NewImage(240, 340)
	hours := []int{8, 14, 23}
	ticks := []int64{0, 37, 1001}
	steps := max(len(allMoods), len(allWellness), len(hours), len(ticks))
	for _, pet := range companion.GetAllPets() {
		for i := 0; i < steps; i++ {
			DrawPet(screen, pet, ticks[i%len(ticks)], Scene{
				Mood:     allMoods[i%len(allMoods)],
				Wellness: allWellness[i%len(allWellness)],
				Hour:     hours[i%len(hours)],
			})
		}
	}
	pet := companion.GetAllPets()[0]
	for _, mood := range allMoods {
		for _, w := range allWellness {
			DrawPet(screen, pet, 37, Scene{Mood: mood, Wellness: w, Hour: 23})
		}
	}
}

// TestSceneHours checks the day/night prop boundaries from docs/agents/specs/2026-10-24-companion-window-design.md §6.
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
