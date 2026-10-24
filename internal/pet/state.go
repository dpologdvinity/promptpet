package pet

import "time"

type State struct {
	Version     int       `json:"version"`
	Name        string    `json:"name"`
	Species     string    `json:"species"`
	Born        time.Time `json:"born"`
	Food        float64   `json:"food"`
	Fun         float64   `json:"fun"`
	Health      float64   `json:"health"`
	Away        bool      `json:"away"`
	AwaySince   time.Time `json:"away_since"`
	Adventures  int       `json:"adventures"`
	Meals       int       `json:"meals"`
	LastTick    time.Time `json:"last_tick"`
	LastPrompt  time.Time `json:"last_prompt"`
	EventOffset int64     `json:"event_offset"`
	Repo        RepoSig   `json:"repo"`
	Message     string    `json:"message"`
}
type RepoSig struct {
	GitDir  string `json:"git_dir"`
	Size    int64  `json:"size"`
	MtimeNs int64  `json:"mtime_ns"`
}

const (
	FoodPerHour      = 2.0
	FunPerHour       = 3.0
	HealthPerHour    = 4.0
	MealFood         = 20.0
	MealFun          = 5.0
	FailHealth       = 6.0
	OKHealth         = 1.0
	PetFun           = 10.0
	PetHealth        = 2.0
	PlayFun          = 25.0
	MaxTickHours     = 720.0
	SleepyGap        = 20 * time.Minute
	SickBelow        = 40.0
	HungryBelow      = 25.0
	BoredBelow       = 25.0
	MaxMealsPerCheck = 5
)

func New(now time.Time) *State {
	return &State{Version: 1, Name: "Pip", Species: "cat", Born: now, Food: 80, Fun: 80, Health: 100, LastTick: now}
}
func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
