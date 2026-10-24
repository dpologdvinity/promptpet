package pet

import (
	"math/rand/v2"
	"strings"
	"testing"
	"time"
)

var t0 = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

func TestNew(t *testing.T) {
	s := New(t0)
	if s.Version != 1 || s.Name != "Pip" || s.Species != "cat" || s.Food != 80 || s.Fun != 80 || s.Health != 100 || !s.LastTick.Equal(t0) || !s.Born.Equal(t0) {
		t.Fatalf("%+v", s)
	}
}
func TestTickWorkedExample(t *testing.T) {
	s := New(t0)
	s.Health = 50
	Tick(s, t0.Add(10*time.Hour))
	if s.Food != 60 || s.Fun != 50 || s.Health != 90 || s.Away {
		t.Fatalf("%+v", s)
	}
}
func TestTickLeavesAfter40Hours(t *testing.T) {
	s := New(t0)
	Tick(s, t0.Add(40*time.Hour-time.Second))
	if s.Away {
		t.Fatal("left early")
	}
	Tick(s, t0.Add(40*time.Hour))
	if !s.Away || s.Food != 0 || !s.AwaySince.Equal(t0.Add(40*time.Hour)) {
		t.Fatalf("%+v", s)
	}
}
func TestTickClampsAndCaps(t *testing.T) {
	s := New(t0)
	s.Health = 10
	Tick(s, t0.Add(1000*24*time.Hour)) // capped at 720 h
	if s.Food != 0 || s.Fun != 0 || s.Health != 100 {
		t.Fatalf("%+v", s)
	}
}
func TestTickClockBackwards(t *testing.T) {
	s := New(t0)
	Tick(s, t0.Add(-5*time.Hour))
	if s.Food != 80 || s.Fun != 80 || !s.LastTick.Equal(t0.Add(-5*time.Hour)) {
		t.Fatalf("%+v", s)
	}
}

func rng() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

func TestApplyExit(t *testing.T) {
	s := New(t0)
	s.Health = 50
	ApplyExit(s, 0)
	if s.Health != 51 {
		t.Fatal(s.Health)
	}
	ApplyExit(s, 130)
	ApplyExit(s, 148)
	if s.Health != 51 {
		t.Fatal(s.Health)
	}
	ApplyExit(s, 1)
	if s.Health != 45 {
		t.Fatal(s.Health)
	}
	ApplyExit(s, 127)
	if s.Health != 39 {
		t.Fatal(s.Health)
	}
}
func TestMealCaps(t *testing.T) {
	s := New(t0)
	s.Food = 90
	s.Fun = 98
	if msg := Meal(s, rng()); msg != "" {
		t.Fatal(msg)
	}
	if s.Food != 100 || s.Fun != 100 || s.Meals != 1 {
		t.Fatalf("%+v", s)
	}
}
func TestMealBringsPetBack(t *testing.T) {
	s := New(t0)
	s.Away = true
	s.Food = 0
	s.Fun = 10
	s.Health = 30
	msg := Meal(s, rng())
	want := "Pip came back from an adventure with a " + Souvenirs[rng().IntN(6)] + "!"
	if msg != want || s.Message != want {
		t.Fatalf("%q", msg)
	}
	if s.Away || s.Food != 60 || s.Fun != 45 || s.Health != 60 || s.Adventures != 1 {
		t.Fatalf("%+v", s)
	}
}
func TestPetAndPlay(t *testing.T) {
	s := New(t0)
	s.Fun = 50
	s.Health = 50
	if got := PetIt(s, rng()); len(got) != 1 || got[0] != "Pip purrs and headbutts your hand." {
		t.Fatal(got)
	}
	if s.Fun != 60 || s.Health != 52 {
		t.Fatalf("%+v", s)
	}
	s.Species = "dragon"
	if got := PetIt(s, rng()); got[0] != "Pip puffs a happy little smoke ring." {
		t.Fatal(got)
	}
	s.Species = "blob"
	if got := PetIt(s, rng()); got[0] != "Pip wobbles contentedly." {
		t.Fatal(got)
	}
	got := Play(s, rng())
	if got[0] != "Pip chases the "+Toys[rng().IntN(5)]+"!" || s.Fun != 100 {
		t.Fatal(got, s.Fun)
	}
	s.Away = true
	if got := Play(s, rng()); len(got) != 2 || !strings.HasPrefix(got[0], "Pip came back") {
		t.Fatal(got)
	}
}
func TestValidName(t *testing.T) {
	for _, ok := range []string{"Mochi", " Sir Pounce ", "Ünïcödé", "R2-D2", "O'Malley", "a.b_c"} {
		if _, v := ValidName(ok); !v {
			t.Errorf("%q rejected", ok)
		}
	}
	for _, bad := range []string{"", "   ", "123456789012345678901", "a%b", "a$b", "a`b", "a\\b", "a\x1b[31m", "a\nb", `a"b`} {
		if _, v := ValidName(bad); v {
			t.Errorf("%q accepted", bad)
		}
	}
	if n, _ := ValidName("  Pip  "); n != "Pip" {
		t.Fatal(n)
	}
	if !ValidSpecies("dragon") || ValidSpecies("dog") {
		t.Fatal("species")
	}
}
