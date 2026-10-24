package hooks

import _ "embed"

//go:embed promptpet.zsh
var zsh string

//go:embed promptpet.bash
var bash string

//go:embed promptpet.fish
var fish string

func Script(shell string) (string, bool) {
	switch shell {
	case "zsh":
		return zsh, true
	case "bash":
		return bash, true
	case "fish":
		return fish, true
	}
	return "", false
}
