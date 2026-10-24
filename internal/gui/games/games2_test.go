package games_test

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/dpologdvinity/promptpet/internal/companion"
	"github.com/dpologdvinity/promptpet/internal/gui/games"
)

func TestAllEightMiniGamesAvailable(t *testing.T) {
	all := games.GetAllMiniGames()
	if len(all) != 8 {
		t.Fatalf("expected 8 mini games, got %d", len(all))
	}

	p, _ := companion.GetPet(companion.PetUsagi)
	screen := ebiten.NewImage(240, 340)

	expectedNames := []string{
		"Snack Catch",
		"Puddle Leap",
		"Look Away",
		"Rhythm Dance",
		"Memory Simon",
		"Whack-a-Bug",
		"Balloon Pop",
		"Word Snack",
	}

	for i, g := range all {
		if g == nil {
			t.Fatalf("game at index %d is nil", i)
		}
		if g.Name() != expectedNames[i] {
			t.Errorf("game %d: expected name %q, got %q", i, expectedNames[i], g.Name())
		}
		g.Init(p)
		g.Draw(screen)
	}
}

func TestMemorySimonWin(t *testing.T) {
	p, _ := companion.GetPet(companion.PetUsagi)
	g := games.NewMemorySimon()
	g.Init(p)
	if g.Name() != "Memory Simon" {
		t.Errorf("unexpected name: %s", g.Name())
	}

	screen := ebiten.NewImage(240, 340)
	g.Draw(screen)

	// Set deterministic sequence: [1, 2, 3, 2]
	seq := []int{1, 2, 3, 2}
	g.SetSequence(seq)

	for _, step := range seq {
		ok := g.InputStep(step)
		if !ok {
			t.Fatalf("expected step %d to be correct", step)
		}
	}

	done, won, score := g.Update()
	if !done || !won {
		t.Fatalf("expected game done and won after full sequence, got done=%v won=%v", done, won)
	}
	if score <= 0 {
		t.Errorf("expected positive score on win, got %d", score)
	}
}

func TestMemorySimonFail(t *testing.T) {
	p, _ := companion.GetPet(companion.PetUsagi)
	g := games.NewMemorySimon()
	g.Init(p)

	g.SetSequence([]int{1, 2, 3})
	// Wrong step
	ok := g.InputStep(3)
	if ok {
		t.Fatalf("expected step 3 to be incorrect for first step 1")
	}

	done, won, _ := g.Update()
	if !done || won {
		t.Fatalf("expected game done and lost on wrong input, got done=%v won=%v", done, won)
	}
}

func TestWhackABugWin(t *testing.T) {
	p, _ := companion.GetPet(companion.PetKochi)
	g := games.NewWhackABug()
	g.Init(p)
	if g.Name() != "Whack-a-Bug" {
		t.Errorf("unexpected name: %s", g.Name())
	}

	screen := ebiten.NewImage(240, 340)
	g.Draw(screen)

	// Whack 6 pests to win
	for i := 0; i < 6; i++ {
		g.HitPest()
	}

	done, won, score := g.Update()
	if !done || !won {
		t.Fatalf("expected game done and won after 6 pests, got done=%v won=%v", done, won)
	}
	if score <= 0 {
		t.Errorf("expected positive score on win, got %d", score)
	}
}

func TestWhackABugLadybugPenaltyAndLoss(t *testing.T) {
	p, _ := companion.GetPet(companion.PetKochi)
	g := games.NewWhackABug()
	g.Init(p)

	// Hit ladybugs / miss 3 times
	for i := 0; i < 3; i++ {
		g.HitLadybug()
	}

	done, won, _ := g.Update()
	if !done || won {
		t.Fatalf("expected game done and lost after 3 penalties, got done=%v won=%v", done, won)
	}
}

func TestBalloonPopWin(t *testing.T) {
	p, _ := companion.GetPet(companion.PetMochi)
	g := games.NewBalloonPop()
	g.Init(p)
	if g.Name() != "Balloon Pop" {
		t.Errorf("unexpected name: %s", g.Name())
	}

	screen := ebiten.NewImage(240, 340)
	g.Draw(screen)

	// Pop 5 balloons to win
	for i := 0; i < 5; i++ {
		g.PopBalloon()
	}

	done, won, score := g.Update()
	if !done || !won {
		t.Fatalf("expected game done and won after 5 balloons popped, got done=%v won=%v", done, won)
	}
	if score <= 0 {
		t.Errorf("expected positive score on win, got %d", score)
	}
}

func TestBalloonPopLoss(t *testing.T) {
	p, _ := companion.GetPet(companion.PetMochi)
	g := games.NewBalloonPop()
	g.Init(p)

	// Miss 4 balloons to lose
	for i := 0; i < 4; i++ {
		g.MissBalloon()
	}

	done, won, _ := g.Update()
	if !done || won {
		t.Fatalf("expected game done and lost after misses, got done=%v won=%v", done, won)
	}
}

func TestWordSnackWin(t *testing.T) {
	p, _ := companion.GetPet(companion.PetPippin)
	g := games.NewWordSnack()
	g.Init(p)
	if g.Name() != "Word Snack" {
		t.Errorf("unexpected name: %s", g.Name())
	}

	screen := ebiten.NewImage(240, 340)
	g.Draw(screen)

	// Spell 3 words to win
	g.SetCurrentWord("TEA")
	if !g.TypeLetter('T') || !g.TypeLetter('E') || !g.TypeLetter('A') {
		t.Fatalf("expected TEA to be successfully typed")
	}

	g.SetCurrentWord("JOY")
	if !g.TypeLetter('J') || !g.TypeLetter('O') || !g.TypeLetter('Y') {
		t.Fatalf("expected JOY to be successfully typed")
	}

	g.SetCurrentWord("PET")
	if !g.TypeLetter('P') || !g.TypeLetter('E') || !g.TypeLetter('T') {
		t.Fatalf("expected PET to be successfully typed")
	}

	done, won, score := g.Update()
	if !done || !won {
		t.Fatalf("expected game done and won after 3 words, got done=%v won=%v", done, won)
	}
	if score <= 0 {
		t.Errorf("expected positive score on win, got %d", score)
	}
}

func TestWordSnackLoss(t *testing.T) {
	p, _ := companion.GetPet(companion.PetPippin)
	g := games.NewWordSnack()
	g.Init(p)

	for i := 0; i < 3; i++ {
		g.MissWord()
	}

	done, won, _ := g.Update()
	if !done || won {
		t.Fatalf("expected game done and lost after 3 missed words, got done=%v won=%v", done, won)
	}
}
