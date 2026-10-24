package pet

import "time"

type Mood int

const (
	Content Mood = iota
	Happy
	Eating
	Sick
	Startled
	Confused
	Sleepy
	Hungry
	Bored
	Away
)

type MoodInput struct {
	Now, PrevPrompt time.Time
	Exit            int
	Ate             bool
}

func Hearts(s *State) int {
	score := (s.Food + s.Fun + s.Health) / 3
	switch {
	case score < 17:
		return 0
	case score < 50:
		return 1
	case score < 83:
		return 2
	default:
		return 3
	}
}
func MoodOf(s *State, in MoodInput) Mood {
	switch {
	case s.Away:
		return Away
	case in.Ate:
		return Eating
	case s.Health < SickBelow:
		return Sick
	case in.Exit == 127:
		return Confused
	case in.Exit > 0 && in.Exit != 130 && in.Exit != 148:
		return Startled
	case (!in.PrevPrompt.IsZero() && in.Now.Sub(in.PrevPrompt) >= SleepyGap) || (in.Now.Hour() >= 1 && in.Now.Hour() <= 5):
		return Sleepy
	case s.Food < HungryBelow:
		return Hungry
	case s.Fun < BoredBelow:
		return Bored
	case Hearts(s) == 3:
		return Happy
	default:
		return Content
	}
}
