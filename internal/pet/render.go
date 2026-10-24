package pet

import "strings"

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
