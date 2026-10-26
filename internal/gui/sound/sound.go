package sound

import (
	"encoding/binary"
	"math"
	"os"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const sampleRate = 44100

var (
	once         sync.Once
	audioCtx     *audio.Context
	mu           sync.Mutex
	muted        bool
	blipBytes    []byte
	jumpBytes    []byte
	successBytes []byte
	missBytes    []byte
)

func generateSquareSweep(startFreq, endFreq, durationSec, volume float64) []byte {
	numSamples := int(float64(sampleRate) * durationSec)
	buf := make([]byte, numSamples*4)

	phase := 0.0
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(numSamples)
		freq := startFreq + (endFreq-startFreq)*t
		phase += freq / float64(sampleRate)
		if phase >= 1.0 {
			phase -= math.Floor(phase)
		}

		// Fade out envelope in the last 25% of the sound to avoid clicks
		env := 1.0
		if t > 0.75 {
			env = (1.0 - t) / 0.25
		}

		sampleVal := int16(0)
		if phase < 0.5 {
			sampleVal = int16(volume * env * 32767.0)
		} else {
			sampleVal = int16(-volume * env * 32767.0)
		}

		binary.LittleEndian.PutUint16(buf[i*4:], uint16(sampleVal))
		binary.LittleEndian.PutUint16(buf[i*4+2:], uint16(sampleVal))
	}
	return buf
}

func initAudio() {
	if os.Getenv("PROMPTPET_SOUND") == "0" {
		muted = true
		return
	}
	defer func() {
		_ = recover() // safe fallback in case audio device initialization fails
	}()
	audioCtx = audio.NewContext(sampleRate)
	blipBytes = generateSquareSweep(880, 880, 0.04, 0.08)
	jumpBytes = generateSquareSweep(320, 680, 0.10, 0.09)
	successBytes = append(
		generateSquareSweep(587, 587, 0.07, 0.09),
		generateSquareSweep(880, 880, 0.12, 0.09)...,
	)
	missBytes = generateSquareSweep(200, 110, 0.12, 0.08)
}

func play(src []byte) {
	once.Do(initAudio)
	mu.Lock()
	defer mu.Unlock()
	if muted || audioCtx == nil || len(src) == 0 {
		return
	}
	p := audioCtx.NewPlayerFromBytes(src)
	p.Play()
}

// PlayBlip plays a soft UI confirmation blip.
func PlayBlip() {
	play(blipBytes)
}

// PlayJump plays an upward frequency chirp.
func PlayJump() {
	play(jumpBytes)
}

// PlaySuccess plays a celebratory two-tone retro jingle.
func PlaySuccess() {
	play(successBytes)
}

// PlayMiss plays a gentle downward low buzz.
func PlayMiss() {
	play(missBytes)
}

// ToggleMute toggles sound playback on or off.
func ToggleMute() bool {
	once.Do(initAudio)
	mu.Lock()
	defer mu.Unlock()
	muted = !muted
	return muted
}

// IsMuted returns true if sound playback is muted.
func IsMuted() bool {
	mu.Lock()
	defer mu.Unlock()
	return muted
}

// SetMuted sets mute state directly.
func SetMuted(m bool) {
	mu.Lock()
	defer mu.Unlock()
	muted = m
}
