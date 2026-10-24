package hooks

import _ "embed"

//go:embed promptpet.zsh
var zsh string

//go:embed promptpet.bash
var bash string

//go:embed promptpet.fish
var fish string

//go:embed completion.zsh
var completionZsh string

//go:embed completion.bash
var completionBash string

//go:embed completion.fish
var completionFish string

// Script returns the shell integration hook for the specified shell.
// The hook captures the previous command's exit code and passes it to
// `promptpet prompt`, allowing promptpet to broadcast command completion events.
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

func Completion(shell string) (string, bool) {
	switch shell {
	case "zsh":
		return completionZsh, true
	case "bash":
		return completionBash, true
	case "fish":
		return completionFish, true
	}
	return "", false
}
