# promptpet

A tiny pet that lives in your shell prompt: `(=^･ω･^=) ♥♥♡` — and on your desktop as a retro handheld companion!

`promptpet` is a fast Go CLI tool and handheld desktop companion. It adds a responsive virtual companion to your shell that eats commits you make in git repositories, reacts to command success and failure, and sleeps when you are idle, in about 2 ms per prompt, so your terminal never slows down. It also includes an optional retro handheld virtual companion mini-window (`promptpet gui`) with 8 cute pets, 8 keyboard mini-games, and gentle wellness check-ins.

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
go install github.com/dpologdvinity/promptpet/cmd/promptpet-gui@latest  # only for the desktop window
```

Or download precompiled release archives or `.deb`/`.rpm` packages for Linux (`amd64`, `arm64`) or macOS (`amd64`, `arm64`) from the GitHub Releases page and put `promptpet` and `promptpet-gui` together in a directory on your `$PATH`.

The desktop window is a separate binary, `promptpet-gui`, so the `promptpet` binary that runs on every prompt never loads the graphics library. Homebrew and the packages install both; `promptpet gui` looks for `promptpet-gui` next to `promptpet`, then on your `$PATH`.

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
promptpet gui            # run in this terminal
promptpet gui --detach   # run in the background as a desktop companion
```

This opens a pocket-sized (240×340) retro handheld LCD console, floating always-on-top and borderless in the bottom-right corner of your screen. It shows your companion with idle animations, reactions to shell commands, and 8 playable mini-games!

---

## Virtual Device Companion (`promptpet gui`)

The companion window brings your shell pet to life in a dedicated retro virtual device:

- **Lightweight & Smooth:** Handcrafted vector art rendered via Ebitengine. Updates at 20 TPS when idle and 60 TPS during mini-games. Idle CPU depends mostly on your graphics driver: under WSLg, Mesa may fall back to software rendering, which used about three cores on our test laptop; `GALLIUM_DRIVER=d3d12 promptpet gui` uses the GPU instead and brought it to about half a core.
- **Real-Time Shell Awareness:** Reacts to shell successes, command errors, and work velocity via asynchronous event IPC, with visual props: sparkles when you are in the flow, confetti on success, a cup of tea and a band-aid after an error, dozing "zzz" when idle, a morning coffee and a night-time nightcap.
- **Casing Colours:** Lavender, Coral, Mint or Midnight; your choice is saved.
- **Git Milestone Fanfare:** Creating a branch (`git checkout -b` / `git switch -c`), merging (`git merge`, a merge `git pull`, or concluding a conflicted merge), and every 5th commit of the day trigger a celebration. Each milestone fires once, on the first prompt after it happens; amends don't count as new commits.
- **Mini-Window Controls:**
  - `Tab`: Cycle through the 8 companion pets.
  - `1` – `8`: Launch one of the 8 keyboard mini-games directly.
  - `K`: Change the casing colour.
  - `M`: Toggle retro 8-bit sound blips on/off.
  - `Esc`: Exit the current mini-game back to companion mode.
  - `Q`: Quit the companion window.
  - **Face buttons (click):** in companion mode `A`/`C` switch to the previous/next pet and `B` changes the casing. In games, `A` = `Left`/`1`, `B` = `Space`/`2`, `C` = `Right`/`3` (aiming in Balloon Pop and the Up/Down lanes in Rhythm Dance still need the keyboard).
  - **Drag** the casing with the mouse to move the window. (Native Wayland sessions don't let apps place or move windows; there the compositor decides. WSLg and X11 work.)
- **One Window Owns the State:** While the window is open it holds `companion.lock`, so a second `promptpet gui` refuses to start and `promptpet pet <name>` asks you to switch with `Tab` instead (otherwise the window would overwrite the change).

---

## Shell & GUI Coexistence

`promptpet` is architected so that the shell prompt and GUI window coexist seamlessly with zero interference:

- **Fast Prompt (about 2 ms):** The shell prompt path only appends a small event record (`command_finish`, exit code, command duration, timestamp, and any git milestone) to `events.jsonl`. The shell hooks time each command without spawning processes (`$EPOCHREALTIME` in zsh and bash 5+, `$CMD_DURATION` in fish); older bash omits the duration. It never waits on GUI locks or blocks terminal execution. The release binary takes a median 2.1 ms per prompt (hyperfine, 500 runs, WSL2 laptop), and CI fails if a build takes more than 10 ms.
- **Asynchronous Event IPC:** The GUI process polls `events.jsonl` once per second (`1 Hz`), maintaining an incremental byte offset cursor. It seeks and streams new terminal events without file contention. The event log automatically rotates at 2 MB to prevent unbounded disk growth.
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
| `5` | **Memory Simon** | Repeat a flashing 3-to-5 step colored sequence without making a mistake | `1`, `2`, `3` |
| `6` | **Whack-a-Bug** | Whack pesky bugs popping from 3 burrows while avoiding ladybugs | `1`, `2`, `3` |
| `7` | **Balloon Pop** | Aim your dart launcher and pop 5 rising balloons with only 6 darts | `Up` / `Down` arrows (or `W` / `S`) to aim, `Spacebar` to shoot |
| `8` | **Word Snack** | Type the drifting letters of delicious snacks before they drift away | Keyboard letters `A` – `Z` |

View high scores anytime from your terminal: `promptpet games`.

---

## Gentle Wellness Reminders

The companion monitors healthy coding habits and delivers subtle in-character speech bubbles, with a matching prop (a glass of water, stretching arrows, tiny binoculars) for 30 seconds:

- **Drink Water (every 45 min):** Prompts you to hydrate (*"Quench your spark with a fresh glass of water!"*).
- **Stretch (every 50 min):** Encourages shoulder and posture reset (*"Ears up, gentle hop stretch!"*).
- **20-20-20 Eye Rest (every 20 min):** Suggests looking away from screens at distance (*"Look at the far meadow for 20 seconds!"*).
- **Anti-Spam Staggering:** Reminders are prioritized (Water > Stretch > Look Away) and automatically staggered so you are never overwhelmed with multiple reminders simultaneously.

---

## Architecture Overview

`promptpet` is designed with strict performance, stability, and isolation guarantees:

- **Strict Prompt Budget (<10 ms):** The prompt path performs no network requests and no git subprocess execution. CI times the built binary end to end, because an in-process benchmark cannot see process startup.
- **Separate Window Binary:** The desktop window is `promptpet-gui`. Go runs every linked package's init code on each start, and the graphics library's took 12 ms, so linking it into `promptpet` once made every prompt 6.6 times slower (median 13.9 ms in v0.5.0, 2.1 ms now). A test fails if `promptpet` links it again.
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
promptpet status           show pet status card, companion hearts, casing and high scores
promptpet pet [name|list]  pet your pet, switch, or list companion pets
promptpet init <shell>     print the shell hook (zsh, bash, fish)
promptpet completion <sh>  print shell autocompletions (zsh, bash, fish)
promptpet gui [--detach]   launch virtual companion handheld mini-window (--detach: in background)
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
- `PROMPTPET_SOUND`: Set to `0` to disable retro audio sound effects.
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

This project is built with AI coding agents using a spec → plan → execute workflow. Agent instructions live in [AGENTS.md](AGENTS.md). The design specs and implementation plans live in [docs/agents/](docs/agents/).

---

## License

MIT. See [LICENSE](LICENSE).
