package pet

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
	"unicode"
)

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

var Toys = []string{"ball of yarn", "paper airplane", "laser dot", "rubber duck", "cardboard box"}
var Souvenirs = []string{"shiny pebble", "pinecone", "lost semicolon", "tiny map", "feather", "rubber duck"}
var Species = []string{"cat", "dragon", "blob"}

func ApplyExit(s *State, code int) {
	switch code {
	case 0:
		s.Health = clamp(s.Health + OKHealth)
	case 130, 148:
	default:
		s.Health = clamp(s.Health - FailHealth)
	}
}
func Return(s *State, rng *rand.Rand) string {
	s.Away = false
	s.AwaySince = time.Time{}
	if s.Food < 40 {
		s.Food = 40
	}
	if s.Fun < 40 {
		s.Fun = 40
	}
	if s.Health < 60 {
		s.Health = 60
	}
	s.Adventures++
	s.Message = fmt.Sprintf("%s came back from an adventure with a %s!", s.Name, Souvenirs[rng.IntN(len(Souvenirs))])
	return s.Message
}
func Meal(s *State, rng *rand.Rand) string {
	msg := ""
	if s.Away {
		msg = Return(s, rng)
	}
	s.Food = clamp(s.Food + MealFood)
	s.Fun = clamp(s.Fun + MealFun)
	s.Meals++
	return msg
}
func PetIt(s *State, rng *rand.Rand) []string {
	lines := []string{}
	if s.Away {
		lines = append(lines, Return(s, rng))
	}
	s.Fun = clamp(s.Fun + PetFun)
	s.Health = clamp(s.Health + PetHealth)
	phrase := map[string]string{"cat": "purrs and headbutts your hand.", "dragon": "puffs a happy little smoke ring.", "blob": "wobbles contentedly."}[s.Species]
	return append(lines, s.Name+" "+phrase)
}
func Play(s *State, rng *rand.Rand) []string {
	lines := []string{}
	if s.Away {
		lines = append(lines, Return(s, rng))
	}
	s.Fun = clamp(s.Fun + PlayFun)
	return append(lines, fmt.Sprintf("%s chases the %s!", s.Name, Toys[rng.IntN(len(Toys))]))
}
func ValidName(n string) (string, bool) {
	n = strings.TrimSpace(n)
	length := 0
	for _, r := range n {
		length++
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune(" -_'.", r)) {
			return n, false
		}
	}
	return n, length >= 1 && length <= 20
}
func ValidSpecies(k string) bool {
	for _, s := range Species {
		if k == s {
			return true
		}
	}
	return false
}
