package companion_test

import (
	"strings"
	"testing"
	"time"

	"github.com/dpologdvinity/promptpet/internal/companion"
	"github.com/dpologdvinity/promptpet/internal/event"
)

func TestWellnessReminders(t *testing.T) {
	s := &companion.State{
		LastWaterTime: time.Now().Add(-65 * time.Minute).Unix(),
	}
	reminder := s.CheckWellness(time.Now())
	if reminder != companion.WellnessWater {
		t.Fatalf("expected WellnessWater, got %v", reminder)
	}
}

func TestRhythmEvaluation(t *testing.T) {
	s := &companion.State{}
	events := []event.Event{
		{Timestamp: time.Now().Unix(), ExitCode: 0, Type: "command_finish"},
	}
	mood, _ := s.EvaluateRhythm(events, time.Now())
	if mood != companion.MoodCheer {
		t.Fatalf("expected MoodCheer on exit 0, got %v", mood)
	}
}

func TestWellnessIntervalsAndResilience(t *testing.T) {
	now := time.Now()

	// LookAway reminder (after 20m)
	sEye := &companion.State{
		LastLookAwayTime: now.Add(-25 * time.Minute).Unix(),
	}
	if rem := sEye.CheckWellness(now); rem != companion.WellnessLookAway {
		t.Errorf("expected WellnessLookAway, got %v", rem)
	}

	// Stretch reminder (after 50m)
	sStretch := &companion.State{
		LastStretchTime: now.Add(-55 * time.Minute).Unix(),
	}
	if rem := sStretch.CheckWellness(now); rem != companion.WellnessStretch {
		t.Errorf("expected WellnessStretch, got %v", rem)
	}

	// Uninitialized state (timestamps 0) should not immediately trigger reminders
	sFresh := &companion.State{}
	if rem := sFresh.CheckWellness(now); rem != companion.WellnessNone {
		t.Errorf("fresh state should return WellnessNone, got %v", rem)
	}

	// Flooding prevention after long idle (all overdue)
	// Water should fire first and others should be staggered, not fired simultaneously
	sIdle := &companion.State{
		LastWaterTime:    now.Add(-120 * time.Minute).Unix(),
		LastStretchTime:  now.Add(-120 * time.Minute).Unix(),
		LastLookAwayTime: now.Add(-120 * time.Minute).Unix(),
	}
	firstRem := sIdle.CheckWellness(now)
	if firstRem != companion.WellnessWater {
		t.Errorf("expected water first on idle return, got %v", firstRem)
	}
	// Immediately subsequent call should be none, avoiding spamming
	secondRem := sIdle.CheckWellness(now)
	if secondRem != companion.WellnessNone {
		t.Errorf("expected no immediate duplicate reminder, got %v", secondRem)
	}
}

func TestRhythmEvaluationVariants(t *testing.T) {
	now := time.Now()
	s := &companion.State{}

	// Exit code != 0 -> Sympathy
	errEvents := []event.Event{
		{Timestamp: now.Unix(), ExitCode: 1, Type: "command_finish"},
	}
	mood, speech := s.EvaluateRhythm(errEvents, now)
	if mood != companion.MoodSympathy {
		t.Errorf("expected MoodSympathy, got %v", mood)
	}
	if !strings.Contains(speech, "tea") {
		t.Errorf("expected tea sympathy speech, got %q", speech)
	}

	// Command burst (3+ commands in past 2 mins) -> Energetic
	burstEvents := []event.Event{
		{Timestamp: now.Add(-40 * time.Second).Unix(), ExitCode: 0, Type: "command_finish"},
		{Timestamp: now.Add(-20 * time.Second).Unix(), ExitCode: 0, Type: "command_finish"},
		{Timestamp: now.Unix(), ExitCode: 0, Type: "command_finish"},
	}
	moodBurst, _ := s.EvaluateRhythm(burstEvents, now)
	if moodBurst != companion.MoodEnergetic {
		t.Errorf("expected MoodEnergetic for burst, got %v", moodBurst)
	}

	// Long silence (> 5m) -> Resting with time of day
	oldEvents := []event.Event{
		{Timestamp: now.Add(-10 * time.Minute).Unix(), ExitCode: 0, Type: "command_finish"},
	}
	morning := time.Date(2026, 10, 24, 8, 30, 0, 0, time.Local)
	moodRest, speechMorning := s.EvaluateRhythm(oldEvents, morning)
	if moodRest != companion.MoodResting {
		t.Errorf("expected MoodResting for long silence, got %v", moodRest)
	}
	if !strings.Contains(speechMorning, "coffee") {
		t.Errorf("expected morning coffee speech, got %q", speechMorning)
	}

	lateNight := time.Date(2026, 10, 24, 23, 30, 0, 0, time.Local)
	_, speechNight := s.EvaluateRhythm([]event.Event{}, lateNight)
	if !strings.Contains(speechNight, "zzz") && !strings.Contains(speechNight, "quiet") {
		t.Errorf("expected late night speech, got %q", speechNight)
	}
}

func TestRhythmEvaluationGitMilestone(t *testing.T) {
	now := time.Now()
	s := &companion.State{}

	// 1. 5th commit milestone
	evCommit5 := []event.Event{
		{Timestamp: now.Unix(), ExitCode: 0, GitAction: "commit", CommitCount: 5},
	}
	mood, speech := s.EvaluateRhythm(evCommit5, now)
	if mood != companion.MoodCheer {
		t.Errorf("expected MoodCheer, got %v", mood)
	}
	if speech != "5th commit! Amazing progress!" {
		t.Errorf("expected '5th commit! Amazing progress!', got %q", speech)
	}

	// 2. 10th commit milestone
	evCommit10 := []event.Event{
		{Timestamp: now.Unix(), ExitCode: 0, GitAction: "commit", CommitCount: 10},
	}
	mood, speech = s.EvaluateRhythm(evCommit10, now)
	if mood != companion.MoodCheer {
		t.Errorf("expected MoodCheer, got %v", mood)
	}
	if speech != "10th commit! Amazing progress!" {
		t.Errorf("expected '10th commit! Amazing progress!', got %q", speech)
	}

	// 3. Regular commit pushed
	evCommitRegular := []event.Event{
		{Timestamp: now.Unix(), ExitCode: 0, GitAction: "commit", CommitCount: 2},
	}
	mood, speech = s.EvaluateRhythm(evCommitRegular, now)
	if mood != companion.MoodCheer {
		t.Errorf("expected MoodCheer, got %v", mood)
	}
	if speech != "Commit pushed! High five!" {
		t.Errorf("expected 'Commit pushed! High five!', got %q", speech)
	}

	// 4. Failed commit gets sympathy over milestone
	evCommitFailed := []event.Event{
		{Timestamp: now.Unix(), ExitCode: 1, GitAction: "commit", CommitCount: 5},
	}
	mood, speech = s.EvaluateRhythm(evCommitFailed, now)
	if mood != companion.MoodSympathy {
		t.Errorf("expected MoodSympathy for failed commit, got %v", mood)
	}
	if !strings.Contains(speech, "tea") {
		t.Errorf("expected sympathy tea speech, got %q", speech)
	}
}
