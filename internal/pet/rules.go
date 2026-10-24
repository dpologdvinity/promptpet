package pet

import "time"

func Tick(s *State, now time.Time) {
	h := now.Sub(s.LastTick).Hours()
	if h < 0 {
		h = 0
	}
	if h > MaxTickHours {
		h = MaxTickHours
	}
	s.Food = clamp(s.Food - FoodPerHour*h)
	s.Fun = clamp(s.Fun - FunPerHour*h)
	s.Health = clamp(s.Health + HealthPerHour*h)
	s.LastTick = now
	if !s.Away && s.Food == 0 {
		s.Away = true
		s.AwaySince = now
	}
}
