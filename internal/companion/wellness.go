package companion

import (
	"time"

	"github.com/dpologdvinity/promptpet/internal/event"
)

type WellnessType string

const (
	WellnessNone     WellnessType = ""
	WellnessWater    WellnessType = "water"
	WellnessStretch  WellnessType = "stretch"
	WellnessLookAway WellnessType = "lookaway"
)

type RhythmMood string

const (
	MoodEnergetic RhythmMood = "energetic"
	MoodResting   RhythmMood = "resting"
	MoodCheer     RhythmMood = "cheer"
	MoodSympathy  RhythmMood = "sympathy"
)

const (
	WaterInterval    = 45 * time.Minute
	StretchInterval  = 50 * time.Minute
	LookAwayInterval = 20 * time.Minute
)

// CheckWellness evaluates elapsed time since last wellness reminders.
// It avoids flooding multiple reminders at once and staggers pending reminders.
func (s *State) CheckWellness(now time.Time) WellnessType {
	if s == nil {
		return WellnessNone
	}

	nowUnix := now.Unix()
	// Uninitialized state: initialize to now to prevent immediate spamming.
	if s.LastWaterTime == 0 && s.LastStretchTime == 0 && s.LastLookAwayTime == 0 {
		s.LastWaterTime = nowUnix
		s.LastStretchTime = nowUnix
		s.LastLookAwayTime = nowUnix
		return WellnessNone
	}
	if s.LastWaterTime == 0 {
		s.LastWaterTime = nowUnix
	}
	if s.LastStretchTime == 0 {
		s.LastStretchTime = nowUnix
	}
	if s.LastLookAwayTime == 0 {
		s.LastLookAwayTime = nowUnix
	}

	waterDue := now.Sub(time.Unix(s.LastWaterTime, 0)) >= WaterInterval
	stretchDue := now.Sub(time.Unix(s.LastStretchTime, 0)) >= StretchInterval
	lookAwayDue := now.Sub(time.Unix(s.LastLookAwayTime, 0)) >= LookAwayInterval

	// Priority: Water > Stretch > LookAway
	// Anti-flooding / resilience: stagger remaining overdue reminders.
	if waterDue {
		s.LastWaterTime = nowUnix
		if stretchDue {
			// Delay stretch by at least 15m from now
			s.LastStretchTime = now.Add(-35 * time.Minute).Unix()
		}
		if lookAwayDue {
			// Delay look away by at least 10m from now
			s.LastLookAwayTime = now.Add(-10 * time.Minute).Unix()
		}
		return WellnessWater
	}

	if stretchDue {
		s.LastStretchTime = nowUnix
		if lookAwayDue {
			// Delay look away by at least 10m from now
			s.LastLookAwayTime = now.Add(-10 * time.Minute).Unix()
		}
		return WellnessStretch
	}

	if lookAwayDue {
		s.LastLookAwayTime = nowUnix
		return WellnessLookAway
	}

	return WellnessNone
}

// EvaluateRhythm analyzes recent events to evaluate work rhythm and mascot dialogue.
func (s *State) EvaluateRhythm(events []event.Event, now time.Time) (RhythmMood, string) {
	if len(events) == 0 {
		return MoodResting, restingSpeechForTime(now)
	}

	nowUnix := now.Unix()
	var latestEvent *event.Event
	recentBurstCount := 0

	for i := range events {
		ev := &events[i]
		evTs := ev.Timestamp
		if evTs == 0 {
			evTs = nowUnix
		}

		if latestEvent == nil || evTs > latestEvent.Timestamp {
			latestEvent = ev
		}

		// Count command bursts in the last 2 minutes
		if nowUnix-evTs <= 120 && evTs <= nowUnix+5 {
			recentBurstCount++
		}
	}

	latestTs := latestEvent.Timestamp
	if latestTs == 0 {
		latestTs = nowUnix
	}

	// Long silence (> 5m with no events) -> resting
	if now.Sub(time.Unix(latestTs, 0)) > 5*time.Minute {
		return MoodResting, restingSpeechForTime(now)
	}

	// If the latest event errored, prioritize sympathy
	if latestEvent.ExitCode != 0 {
		return MoodSympathy, "You got this! Here's some tea"
	}

	// Command bursts (>= 3 in 2 mins) -> energetic
	if recentBurstCount >= 3 {
		return MoodEnergetic, "In the zone! Keep up the great pace!"
	}

	// Exit code 0 success -> cheer
	return MoodCheer, "Yay! Success!"
}

func restingSpeechForTime(t time.Time) string {
	hour := t.Hour()
	switch {
	case hour >= 6 && hour < 12:
		return "Morning coffee time... ready when you are!"
	case hour >= 12 && hour < 18:
		return "Afternoon playtime... enjoying the day!"
	case hour >= 18 && hour < 22:
		return "Cozy evening resting... take a deep breath."
	default:
		return "Late night quiet... zzz"
	}
}
