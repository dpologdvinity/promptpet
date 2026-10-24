package pet

import (
	"fmt"
	"math"
	"strings"
	"time"
)

var Moods = []Mood{Content, Happy, Eating, Sick, Startled, Confused, Sleepy, Hungry, Bored}
var faces = map[string][9]string{
	"cat":    {"(=^･ω･^=)", "(=^‿^=)", "(=^･ω･^=) nom", "(=~ω~=) sniffle", "(=°ω°=)!", "(=･ω･=)?", "(=-ω-=) zz", "(=･ω･=) snack?", "(=¬ω¬=)"},
	"dragon": {"<(°ᴗ°)>", "<(^ᴗ^)>", "<(°ᴗ°)> nom", "<(~_~)> sniffle", "<(°o°)>!", "<(°_°)>?", "<(-ᴗ-)> zz", "<(°ᴗ°)> snack?", "<(¬_¬)>"},
	"blob":   {"(•‿•)", "(^‿^)", "(•o•) nom", "(~_~) sniffle", "(°o°)!", "(•_•)?", "(-‿-) zz", "(•‿•) snack?", "(¬_¬)"},
}

func Face(species string, m Mood) string {
	if m == Away {
		return "(out exploring)"
	}
	row, ok := faces[species]
	if !ok {
		row = faces["cat"]
	}
	if m < Content || m > Bored {
		m = Content
	}
	return row[m]
}
func Segment(s *State, m Mood) string {
	if m == Away {
		return Face(s.Species, m)
	}
	n := Hearts(s)
	return Face(s.Species, m) + " " + strings.Repeat("♥", n) + strings.Repeat("♡", 3-n)
}

var moodTexts = [10]string{"feeling content", "feeling great", "munching on your latest commit", "feeling a bit under the weather", "a little startled", "confused by that command", "sleepy", "peckish - a commit would be lovely", "a bit bored - try promptpet play", "out exploring - commit something or run promptpet pet to call them back"}

func MoodText(m Mood) string {
	if m < Content || m > Away {
		return moodTexts[Content]
	}
	return moodTexts[m]
}

var arts = map[string]string{
	"cat":    " /\\_/\\\n( E.E )\n > ^ <",
	"dragon": "  <\\_/>\n ( E E )~~\n  /|_|\\",
	"blob":   "  .---.\n ( E E )\n  '---'",
}
var eyes = [9]string{"o", "^", "^", "~", "O", "?", "-", "o", "¬"}

func Card(s *State, m Mood, now time.Time) string {
	art := "  \\     /\n   \\___/\n  (empty basket)"
	if m != Away {
		art = arts[s.Species]
		if art == "" {
			art = arts["cat"]
		}
		if m < Content || m > Bored {
			m = Content
		}
		art = strings.ReplaceAll(art, "E", eyes[m])
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n%s the %s - %s\n", art, s.Name, s.Species, MoodText(m))
	for _, stat := range []struct {
		label string
		value float64
	}{{"food", s.Food}, {"fun", s.Fun}, {"health", s.Health}} {
		v := int(math.Floor(stat.value))
		filled := (v + 5) / 10
		if filled < 0 {
			filled = 0
		}
		if filled > 10 {
			filled = 10
		}
		fmt.Fprintf(&b, "%-7s %s %3d\n", stat.label, strings.Repeat("█", filled)+strings.Repeat("░", 10-filled), v)
	}
	days := int(now.Sub(s.Born).Hours() / 24)
	if days < 0 {
		days = 0
	}
	fmt.Fprintf(&b, "age %d %s · %d %s · %d %s\n", days, plural(days, "day"), s.Meals, plural(s.Meals, "meal"), s.Adventures, plural(s.Adventures, "adventure"))
	if s.Message != "" {
		fmt.Fprintln(&b, s.Message)
	}
	return b.String()
}
func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
