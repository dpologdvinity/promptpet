# promptpet

A tiny pet that lives in your shell prompt: `(=^･ω･^=) ♥♥♡` — and on your desktop as a retro handheld companion!

`promptpet` is an ultra-fast, zero-dependency Go CLI tool and handheld desktop companion. It adds a responsive virtual companion to your shell that eats commits you make in git repositories, reacts to command success and failure, and sleeps when you are idle—all rendered in under 5 ms without slowing down your terminal. It also includes an optional retro handheld virtual companion mini-window (`promptpet gui`) with 8 cute pets, 8 keyboard mini-games, and gentle wellness check-ins.

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

### 4. Launch Desktop Companion (GUI)

Want your pet on your desktop alongside your terminal? Run:

```sh
promptpet gui
```

This opens a pocket-sized (240×340) retro handheld LCD console displaying your companion with animated blinking, breathing, reactions to shell commands, and 8 playable mini-games!

---

## Virtual Device Companion (`promptpet gui`)

The companion window brings your shell pet to life in a dedicated retro virtual device:

- **Ultra-Lightweight & Smooth:** Handcrafted vector art rendered via Ebitengine. Runs at 20 TPS / ~1% CPU when idle; switches to 60 TPS during mini-gameplay for fluid responsiveness.
- **Real-Time Shell Awareness:** Reacts to shell successes, command errors, and work velocity via asynchronous event IPC.
- **Mini-Window Controls:**
  - `Tab`: Cycle through the 8 companion pets.
  - `1` – `8`: Launch one of the 8 keyboard mini-games directly.
  - `Esc`: Exit the current mini-game back to companion mode.

---

## Shell & GUI Coexistence

`promptpet` is architected so that the shell prompt and GUI window coexist seamlessly with zero interference:

- **Guaranteed Fast Prompt Latency (<5 ms):** The shell prompt path only appends a small event record (`command_finish`, exit code, timestamp) to `events.jsonl`. It never waits on GUI locks or blocks terminal execution. Benchmarks confirm prompt execution averages ~1.2 ms/op.
- **Asynchronous Event IPC:** The GUI process polls `events.jsonl` once per second (`1 Hz`), maintaining an incremental byte offset cursor. It seeks and streams new terminal events without file contention.
- **Independent State Isolation:** Prompt vitals and feeding live in `state.json`, while companion pet preferences, wellness timestamps, and mini-game high scores reside in `companion.json`.
- **Headless Detection:** If launched without an active X11, Wayland, or desktop display server, `promptpet gui` exits cleanly with a helpful error message instead of crashing.

---

## 8 Companion Pets Roster

Switch your companion anytime using `promptpet pet <name>` or pressing `Tab` in the GUI:

| ID | Name | Species | Favorite Snack | Personality & Quote Vibe |
|---|---|---|---|---|
| `usagi` | Usagi | Dumpling Bunny | Fresh Strawberry | Cheerful & encouraging (*"Hop hop! Great job!"*) |
| `kochi` | Kochi | Shiba Pup | Rice Dumpling | Enthusiastic loyal puppy (*"Woof! You did it! Tail wagging!"*) |
| `puff` | Puff | Baby Cloud Dragon | Cinnamon Star | Fiery & aspirational (*"Tiny flame of victory! \*poof\*"*) |
| `mochi` | Mochi | Strawberry Cat | Taiyaki Waffle | Gentle & comforting (*"Purr... that was purrfect!"*) |
| `spook` | Spook | Cloud Ghostie | Sugar Candy | Playful & ethereal (*"Boo-yah! That was spectacular!"*) |
| `pippin` | Pippin | Cozy Penguin | Ice Popsicle | Resilient winter buddy (*"Waddle dance! Outstanding work!"*) |
| `gromit` | Gromit | Rainy Froggy | Dewdrop Melon | Peaceful nature lover (*"Ribbit! Splendid hop forward!"*) |
| `boba` | Boba | Tea Panda | Tapioca Pearl | Chill & relaxing (*"Sweet boba victory! Great job!"*) |

---

## 8 Keyboard Mini-Games

Take a short mental break with 8 retro mini-games playable in the companion window (press keys `1`–`8` to launch, `Esc` to return):

| ID | Mini-Game | Description | Controls |
|---|---|---|---|
| `1` | **Snack Catch** | Catch your pet's favorite falling treats in the basket | `Left` / `Right` arrows (or `A` / `D`) |
| `2` | **Puddle Leap** | Time your jumps over approaching mud puddles | `Spacebar` (or `Up` arrow / `W`) |
| `3` | **Look Away** | Quickly match target directions to rest and stretch your eyes | `Left` / `Right` arrows (or `A` / `D`) |
| `4` | **Rhythm Dance** | Tap dancing arrow notes precisely as they reach the target line | `Arrow keys` (or `W` `A` `S` `D`) |
| `5` | **Memory Simon** | Repeat the flashing colored sequence without making a mistake | `1`, `2`, `3` |
| `6` | **Whack-a-Bug** | Whack pesky bugs popping from 3 burrows while avoiding ladybugs | `1`, `2`, `3` |
| `7` | **Balloon Pop** | Aim your dart launcher angle and pop rising balloons | `Up` / `Down` arrows (or `W` / `S`) to aim, `Spacebar` to shoot |
| `8` | **Word Snack** | Type the drifting letters of delicious snacks before they drift away | Keyboard letters `A` – `Z` |

View high scores anytime from your terminal: `promptpet games`.

---

## Gentle Wellness Reminders

The companion monitors healthy coding habits and delivers subtle in-character speech bubbles:

- **Drink Water (every 45 min):** Prompts you to hydrate (*"Quench your spark with a fresh glass of water!"*).
- **Stretch (every 50 min):** Encourages shoulder and posture reset (*"Ears up, gentle hop stretch!"*).
- **20-20-20 Eye Rest (every 20 min):** Suggests looking away from screens at distance (*"Look at the far meadow for 20 seconds!"*).
- **Anti-Spam Staggering:** Reminders are prioritized (Water > Stretch > Look Away) and automatically staggered so you are never overwhelmed with multiple reminders simultaneously.

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
promptpet status           show pet status card and companion hearts
promptpet pet [name|list]  pet your pet, switch, or list companion pets
promptpet init <shell>     print the shell hook (zsh, bash, fish)
promptpet completion <sh>  print shell autocompletions (zsh, bash, fish)
promptpet gui              launch virtual companion handheld mini-window
promptpet games            list mini-games, controls, and high scores
promptpet play             play with your pet (terminal interaction)
promptpet name <name>      rename your prompt pet (1-20 chars: letters, digits, spaces, - _ ' .)
promptpet species <kind>   change prompt pet species (cat, dragon, blob)
promptpet prompt           print the prompt segment (used by shell hooks)
promptpet --version        print current version
```

---

## Environment Variables

- `PROMPTPET_HOME`: Overrides the default state directory. Defaults to `$XDG_STATE_HOME/promptpet` (if absolute) or `$HOME/.local/state/promptpet`.
- `PROMPTPET_COMPANION_STATE_FILE`: Overrides the companion state file path (defaults to `$PROMPTPET_HOME/companion.json`).
- `PROMPTPET_EVENT_FILE`: Overrides the event stream file path (defaults to `$PROMPTPET_HOME/events.jsonl`).
- `PROMPTPET_DISABLE`: Set to any non-empty value to temporarily disable prompt segment rendering.
- `PROMPTPET_HEADLESS`: Set to `1` to simulate a headless environment without display.
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
