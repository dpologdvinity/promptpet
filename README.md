# promptpet

A tiny pet that lives in your shell prompt: `(=^･ω･^=) ♥♥♡`.

## Install

Install with Go:

```sh
go install github.com/dpologdvinity/promptpet@latest
```

Or download a Linux or macOS archive for your architecture from the GitHub releases page and put `promptpet` on your `PATH`.

## Setup

Put the hook at the end of your shell rc file so it wraps your final prompt:

```sh
# ~/.zshrc
eval "$(promptpet init zsh)"
```

```sh
# ~/.bashrc
eval "$(promptpet init bash)"
```

```fish
# ~/.config/fish/config.fish
promptpet init fish | source
```

Open a new shell. Run `promptpet` to see the full status card.

## What your pet reacts to

Recent commits you authored feed your pet. Failed commands startle it and lower its health; successful commands help it recover. Ctrl-C and Ctrl-Z do not affect it. Time away makes it sleepy, while food and fun gradually decline. If food reaches zero it goes exploring. A commit, `promptpet pet`, or `promptpet play` brings it back. The pet never dies.

## Commands

```text
promptpet                  show your pet
promptpet init <shell>     print the shell hook (zsh, bash, fish)
promptpet pet              pet your pet
promptpet play             play with your pet
promptpet name <name>      rename your pet
promptpet species <kind>   change species (cat, dragon, blob)
promptpet prompt           print the prompt segment (used by the hook)
promptpet --version        print the version
```

## Privacy

A background process reads only the current repository's last 50 commit times, author emails, and hashes. The state directory stores commit times, short hashes, and the repository path. Nothing leaves your machine; promptpet makes no network requests.

## Environment

- `PROMPTPET_HOME` sets the state directory. Otherwise promptpet uses `$XDG_STATE_HOME/promptpet` when `XDG_STATE_HOME` is absolute, or `$HOME/.local/state/promptpet`.
- `PROMPTPET_DISABLE` set to any nonempty value hides the prompt segment.
- `$PROMPTPET_SEGMENT` (zsh and bash) holds the segment for custom prompt placement.

## License

MIT. See [LICENSE](LICENSE).
