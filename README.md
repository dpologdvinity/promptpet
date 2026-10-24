# promptpet

A tiny pet that lives in your shell prompt: `(=^･ω･^=) ♥♥♡`.

`promptpet` is an ultra-fast, zero-dependency Go CLI tool that adds a responsive virtual companion to your shell. It eats commits you make in git repositories, reacts to command success and failure, and sleeps when you are idle—all rendered in under 10 ms without slowing down your terminal.

---

## Quickstart

### 1. Install

Using Homebrew:

```sh
brew install dpologdvinity/tap/promptpet
```

Using Go:

```sh
go install github.com/dpologdvinity/promptpet@latest
```

Or download precompiled release binaries for Linux (`amd64`, `arm64`) or macOS (`amd64`, `arm64`) from the GitHub Releases page and place `promptpet` on your `$PATH`.

### 2. Shell Setup

Add the initialization hook at the end of your shell configuration file so it wraps your prompt:

#### Zsh (`~/.zshrc`)
```sh
eval "$(promptpet init zsh)"
```

#### Bash (`~/.bashrc`)
```sh
eval "$(promptpet init bash)"
```

#### Fish (`~/.config/fish/config.fish`)
```fish
promptpet init fish | source
```

#### Shell Completions (Optional)
To enable tab completion:

```sh
eval "$(promptpet completion bash)" # bash
eval "$(promptpet completion zsh)"  # zsh
promptpet completion fish > ~/.config/fish/completions/promptpet.fish # fish
```

Open a new shell session. Your pet will appear in your prompt.

### 3. Check Status

Run `promptpet` without arguments to view your pet's full ASCII status card:

```text
 /\_/\
( o.o )
 > ^ <

Pip the cat - feeling content
food    ████████░░  80
fun     ████████░░  80
health  ██████████ 100
age 1 day · 5 meals · 0 adventures
```

---

## Architecture Overview

`promptpet` is designed with strict performance, stability, and isolation guarantees:

- **Strict Prompt Budget (<10 ms):** The prompt path performs no network requests and no git subprocess execution.
- **Detached Background Feeder:** Git commit inspection runs asynchronously in a detached daemon process (`promptpet _feed`), isolated from the interactive shell.
- **Subprocess-Free Repo Detection:** The prompt uses `os.Stat` on `.git/logs/HEAD` to detect git changes without invoking `git`.
- **Fault-Tolerant State:** Concurrency across shells is guarded with non-blocking file locks (`flock`), atomic file writes, and automatic recovery for corrupted state files.

For technical deep dive, data flow, and package boundaries, see [ARCHITECTURE.md](ARCHITECTURE.md).

---

## Pet Behavior & Rules

- **Meals:** Recent commits authored by you feed your pet (+20 food, +5 fun, +1 health).
- **Command Reactions:** Command failures startle your pet and reduce health; command successes restore health.
- **Decay:** Food and fun gradually decay over time. Time away makes your pet sleepy.
- **Adventures:** If food drops to 0, your pet departs on an adventure. Committing changes or running `promptpet pet` calls them back with a souvenir.
- **Safety:** Pets never die.

---

## Commands

```text
promptpet                  show your pet's status card
promptpet init <shell>     print the shell hook (zsh, bash, fish)
promptpet completion <sh>  print shell autocompletions (zsh, bash, fish)
promptpet pet              pet your pet
promptpet play             play with your pet
promptpet name <name>      rename your pet (1-20 chars: letters, digits, spaces, - _ ' .)
promptpet species <kind>   change species (cat, dragon, blob)
promptpet prompt           print the prompt segment (used by shell hooks)
promptpet --version        print current version
```

---

## Environment Variables

- `PROMPTPET_HOME`: Overrides the default state directory. Defaults to `$XDG_STATE_HOME/promptpet` (if absolute) or `$HOME/.local/state/promptpet`.
- `PROMPTPET_DISABLE`: Set to any non-empty value to temporarily disable prompt segment rendering.
- `PROMPTPET_SEGMENT`: Populated in `zsh` and `bash` hooks for manual placement in custom prompt themes.

---

## Privacy & Security

`promptpet` is completely local and offline:
- Zero telemetry or network calls.
- Background feeding inspects only local git author email, timestamps, and commit hashes for the active repository.
- State is stored in standard local state directories with `0700` directory and `0600` file permissions.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for local setup, testing guidelines, and pull request workflows.

---

## License

MIT. See [LICENSE](LICENSE).
