package pet

import (
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
