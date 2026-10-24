package games_test

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/dpologdvinity/promptpet/internal/companion"
	"github.com/dpologdvinity/promptpet/internal/gui/games"
)

func TestSnackCatch(t *testing.T) {
	p, _ := companion.GetPet(companion.PetUsagi)
	g := games.NewSnackCatch()
	g.Init(p)
	if g.Name() != "Snack Catch" {
		t.Errorf("unexpected name: %s", g.Name())
	}

	screen := ebiten.NewImage(240, 340)
	g.Draw(screen)

	// Simulate catching snacks directly or updating
	for i := 0; i < 5; i++ {
		g.CatchSnack()
	}
	done, won, score := g.Update()
	if !done || !won {
		t.Fatalf("expected game to be done and won after 5 catches, got done=%v won=%v", done, won)
	}
	if score <= 0 {
		t.Errorf("expected positive score, got %d", score)
	}
}

func TestSnackCatchMissCondition(t *testing.T) {
	p, _ := companion.GetPet(companion.PetUsagi)
	g := games.NewSnackCatch()
	g.Init(p)

	for i := 0; i < 3; i++ {
		g.MissSnack()
	}
	done, won, _ := g.Update()
	if !done || won {
		t.Fatalf("expected game done and lost after 3 misses, got done=%v won=%v", done, won)
	}
}

func TestPuddleLeap(t *testing.T) {
	p, _ := companion.GetPet(companion.PetKochi)
	g := games.NewPuddleLeap()
	g.Init(p)
	if g.Name() != "Puddle Leap" {
		t.Errorf("unexpected name: %s", g.Name())
	}

	screen := ebiten.NewImage(240, 340)
	g.Draw(screen)

	// Simulate completing the distance to flag
	g.SetProgress(g.FinishDistance())
	done, won, score := g.Update()
	if !done || !won {
		t.Fatalf("expected game done and won after reaching finish, got done=%v won=%v", done, won)
	}
	if score <= 0 {
		t.Errorf("expected positive score on win, got %d", score)
	}
}

func TestPuddleLeapTrips(t *testing.T) {
	p, _ := companion.GetPet(companion.PetKochi)
	g := games.NewPuddleLeap()
	g.Init(p)

	for i := 0; i < 3; i++ {
		g.Trip()
	}
	done, won, _ := g.Update()
	if !done || won {
		t.Fatalf("expected game done and lost after 3 trips, got done=%v won=%v", done, won)
	}
}

func TestLookAway(t *testing.T) {
	p, _ := companion.GetPet(companion.PetKochi)
	g := games.NewLookAway()
	g.Init(p)
	if g.Name() != "Look Away" {
		t.Errorf("unexpected name: %s", g.Name())
	}

	screen := ebiten.NewImage(240, 340)
	g.Draw(screen)

	// Simulate 5 rounds with 3 wins
	g.SimulateRound(true)
	g.SimulateRound(true)
	g.SimulateRound(true)
	g.SimulateRound(false)
	g.SimulateRound(false)

	done, won, score := g.Update()
	if !done || !won {
		t.Fatalf("expected game done and won with 3/5 wins, got done=%v won=%v", done, won)
	}
	if score < 600 {
		t.Errorf("expected score >= 600, got %d", score)
	}
}

func TestLookAwayLoss(t *testing.T) {
	p, _ := companion.GetPet(companion.PetMochi)
	g := games.NewLookAway()
	g.Init(p)

	// Simulate 5 rounds with only 2 wins (loss condition: < 3 wins)
	g.SimulateRound(true)
	g.SimulateRound(true)
	g.SimulateRound(false)
	g.SimulateRound(false)
	g.SimulateRound(false)

	done, won, _ := g.Update()
	if !done || won {
		t.Fatalf("expected game done and lost with 2/5 wins, got done=%v won=%v", done, won)
	}
}

func TestRhythmDance(t *testing.T) {
	p, _ := companion.GetPet(companion.PetPippin)
	g := games.NewRhythmDance()
	g.Init(p)
	if g.Name() != "Rhythm Dance" {
		t.Errorf("unexpected name: %s", g.Name())
	}

	screen := ebiten.NewImage(240, 340)
	g.Draw(screen)

	// Hit 8 beats to win
	for i := 0; i < 8; i++ {
		g.HitBeat()
	}
	done, won, score := g.Update()
	if !done || !won {
		t.Fatalf("expected game done and won after 8 hits, got done=%v won=%v", done, won)
	}
	if score <= 0 {
		t.Errorf("expected positive score, got %d", score)
	}
}

func TestRhythmDanceLoss(t *testing.T) {
	p, _ := companion.GetPet(companion.PetPippin)
	g := games.NewRhythmDance()
	g.Init(p)

	for i := 0; i < 4; i++ {
		g.MissBeat()
	}
	done, won, _ := g.Update()
	if !done || won {
		t.Fatalf("expected game done and lost after 4 misses, got done=%v won=%v", done, won)
	}
}
