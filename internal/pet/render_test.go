package pet

import (
	"strings"
	"testing"
)

func TestSegmentExamples(t *testing.T) {
	s := New(t0)
	s.Food, s.Fun, s.Health = 60, 60, 60 // hearts 2
	if got := Segment(s, Content); got != "(=^･ω･^=) ♥♥♡" {
		t.Fatalf("%q", got)
	}
	s.Species = "dragon"
	if got := Segment(s, Eating); got != "<(°ᴗ°)> nom ♥♥♡" {
		t.Fatalf("%q", got)
	}
	s.Species = "blob"
	s.Food, s.Fun, s.Health = 0, 0, 0
	if got := Segment(s, Startled); got != "(°o°)! ♡♡♡" {
		t.Fatalf("%q", got)
	}
	s.Away = true
	if got := Segment(s, Away); got != "(out exploring)" {
		t.Fatalf("%q", got)
	}
}
func TestSegmentSafeCharacters(t *testing.T) {
	for _, sp := range Species {
		for _, m := range append(Moods, Away) {
			s := New(t0)
			s.Species = sp
			seg := Segment(s, m)
			if seg == "" {
				t.Errorf("%s/%d empty", sp, m)
			}
			for _, r := range seg {
				if r < 0x20 || strings.ContainsRune("%$`\\\"'", r) {
					t.Errorf("%s/%d has %q", sp, m, r)
				}
			}
		}
	}
}
