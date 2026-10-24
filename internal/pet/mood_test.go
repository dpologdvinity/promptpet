package pet

import (
	"testing"
	"time"
)

func TestHearts(t *testing.T) {
	cases := []struct {
		f, fun, h float64
		want      int
	}{
		{0, 0, 50, 0},   // 16.67
		{0, 0, 51, 1},   // 17
		{50, 50, 49, 1}, // 49.67
		{50, 50, 50, 2}, // 50
		{80, 80, 88, 2}, // 82.67
		{80, 80, 89, 3}, // 83
	}
	for _, c := range cases {
		s := New(t0)
		s.Food, s.Fun, s.Health = c.f, c.fun, c.h
		if got := Hearts(s); got != c.want {
			t.Errorf("%v: got %d", c, got)
		}
	}
}
func TestMoodPriority(t *testing.T) {
	noon := t0 // 12:00 UTC
	base := func() *State { s := New(t0); s.Food, s.Fun, s.Health = 60, 60, 60; return s }
	in := MoodInput{Now: noon, Exit: -1}
	check := func(name string, s *State, in MoodInput, want Mood) {
		t.Helper()
		if got := MoodOf(s, in); got != want {
			t.Errorf("%s: got %d want %d", name, got, want)
		}
	}
	s := base()
	s.Away = true
	check("away beats eating", s, MoodInput{Now: noon, Exit: 1, Ate: true}, Away)
	s = base()
	s.Health = 10
	check("eating beats sick", s, MoodInput{Now: noon, Exit: -1, Ate: true}, Eating)
	s = base()
	s.Health = 39.9
	check("sick beats exit", s, MoodInput{Now: noon, Exit: 1}, Sick)
	check("127 confused", base(), MoodInput{Now: noon, Exit: 127}, Confused)
	check("1 startled", base(), MoodInput{Now: noon, Exit: 1}, Startled)
	check("130 no reaction", base(), MoodInput{Now: noon, Exit: 130}, Content)
	check("gap sleepy", base(), MoodInput{Now: noon, PrevPrompt: noon.Add(-20 * time.Minute), Exit: 0}, Sleepy)
	check("short gap", base(), MoodInput{Now: noon, PrevPrompt: noon.Add(-19 * time.Minute), Exit: 0}, Content)
	check("01:00 sleepy", base(), MoodInput{Now: time.Date(2026, 3, 10, 1, 0, 0, 0, time.UTC), Exit: -1}, Sleepy)
	check("06:00 awake", base(), MoodInput{Now: time.Date(2026, 3, 10, 6, 0, 0, 0, time.UTC), Exit: -1}, Content)
	s = base()
	s.Food = 24.9
	s.Fun = 10
	check("hungry beats bored", s, in, Hungry)
	s = base()
	s.Fun = 24.9
	check("bored", s, in, Bored)
	s = base()
	s.Food, s.Fun, s.Health = 90, 90, 90
	check("happy", s, in, Happy)
}
