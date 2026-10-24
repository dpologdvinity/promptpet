package pet

import (
	"strings"
	"testing"
	"time"
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

func TestCardExact(t *testing.T) {
	s := New(t0.Add(-12*24*time.Hour - time.Hour))
	s.Food, s.Fun, s.Health, s.Meals, s.Adventures = 80.9, 60, 100, 34, 3
	want := " /\\_/\\\n( o.o )\n > ^ <\n\n" +
		"Pip the cat - feeling content\n" +
		"food    ████████░░  80\n" +
		"fun     ██████░░░░  60\n" +
		"health  ██████████ 100\n" +
		"age 12 days · 34 meals · 3 adventures\n"
	if got := Card(s, Content, t0); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}
func TestCardSingularAwayAndMessage(t *testing.T) {
	s := New(t0.Add(-25 * time.Hour))
	s.Species = "dragon"
	s.Meals, s.Adventures = 1, 1
	s.Food, s.Fun, s.Health = 4, 0, 45 // bars: 0 filled, 0 filled, 5 filled
	s.Message = "Pip came back from an adventure with a pinecone!"
	got := Card(s, Sleepy, t0)
	for _, line := range []string{"  <\\_/>", " ( - - )~~", "Pip the dragon - sleepy", "food    ░░░░░░░░░░   4", "health  █████░░░░░  45", "age 1 day · 1 meal · 1 adventure", "Pip came back from an adventure with a pinecone!"} {
		if !strings.Contains(got, line+"\n") {
			t.Errorf("missing %q in\n%s", line, got)
		}
	}
	s.Away = true
	if got := Card(s, Away, t0); !strings.Contains(got, "  (empty basket)\n") || !strings.Contains(got, "out exploring - commit something or run promptpet pet to call them back") {
		t.Fatal(got)
	}
}
