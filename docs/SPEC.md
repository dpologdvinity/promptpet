# promptpet — Specification (v0.1)

`promptpet` is a single-binary CLI that puts a small ASCII pet in your shell
prompt. The pet eats when you make git commits, gets sick when many commands
fail, gets sleepy when you have been away, gets bored when you ignore it, and
reacts to the exit code of the last command. Neglect is gentle: the pet never
dies. When it gets too hungry it goes off exploring, and it comes back the
next time you commit or call it.

This document is the source of truth for v0.1 behavior. `docs/PLAN.md` says how
to build it. If the plan and the spec disagree, the spec wins.

## 1. Goals and non-goals

### Goals (v0.1)
- `promptpet init zsh|bash|fish` prints a shell hook that shows the pet segment in the prompt.
- `promptpet prompt` prints a short segment such as `(=^･ω･^=) ♥♥♡` in under 10 ms, with no network and no git calls.
- `promptpet` with no arguments prints a big ASCII status card.
- Interactions: `pet`, `play`, `name`, `species` (cat, dragon, blob).
- Commits feed the pet. Commit detection runs in a detached background process, so it never slows the prompt.
- All stat changes over time are exact, linear functions of real elapsed time (§6).

### Non-goals (v0.1)
- No network access of any kind.
- No color in the prompt segment or the card (§11 Later).
- No Windows build. The hooks target zsh, bash and fish, and file locking uses `syscall.Flock`.
- No multiple pets, no config file, no TUI, no animations.
- No permanent death and no "your pet is sad because of you" copy. All copy is friendly (§8).

## 2. Name availability (checked 2026-10-23)

| Place | Result |
|---|---|
| GitHub `dpologdvinity/promptpet` | free (404) |
| Homebrew formula / cask `promptpet` | free (404 / 404) |
| npm `promptpet` | free (404) |
| GitHub repos named like it | 8 total; the largest is `hamutama/PromptPET` (91 stars, a prompt-engineering toolkit, unrelated). Not a conflict. |

## 3. Tech stack (fixed)

| Item | Choice |
|---|---|
| Language | Go. `go.mod` says `go 1.26` (builds with the installed 1.27.3). |
| Module path | `github.com/dpologdvinity/promptpet` |
| Dependencies | none. Stdlib only: `flag`, `encoding/json`, `math/rand/v2`, `os/exec` (background feeder only), `syscall` (`Flock`, `Setsid`), `embed`. |
| Tests | stdlib `testing` only |
| Release | GoReleaser v2, linux/darwin × amd64/arm64, `CGO_ENABLED=0` |
| CI | GitHub Actions: `actions/checkout@v7`, `actions/setup-go@v7` (latest tags v7.0.1 and v7.0.0 on 2026-10-23), `goreleaser/goreleaser-action@v7` with `version: "~> v2"` (latest v2.19.0) |
| License | MIT, `Copyright (c) 2026 Kaitlyn Bassford` |

No third-party module is needed: flag parsing is a handful of subcommands, the
state is one JSON file, and randomness only picks copy strings. Add no
dependencies in v0.1.

## 4. CLI interface

```
promptpet                     show the status card
promptpet init <shell>        print the hook for zsh, bash or fish
promptpet prompt [--exit N]   print the prompt segment (called by the hook)
promptpet pet                 pet your pet
promptpet play                play with your pet
promptpet name <name>         rename your pet
promptpet species <kind>      change species: cat, dragon, blob
promptpet _feed <dir>         internal: background commit check (§7)
promptpet --version | -V      print "promptpet <version>"
promptpet --help | -h | help  print usage
```

### 4.1 Usage text (exact, printed to stdout by `--help`, `-h`, `help`)

```
promptpet - a tiny pet that lives in your shell prompt

usage:
  promptpet                  show your pet
  promptpet init <shell>     print the shell hook (zsh, bash, fish)
  promptpet pet              pet your pet
  promptpet play             play with your pet
  promptpet name <name>      rename your pet
  promptpet species <kind>   change species (cat, dragon, blob)
  promptpet prompt           print the prompt segment (used by the hook)
  promptpet --version        print the version

Your pet eats when you make git commits.
Setup (zsh):  add  eval "$(promptpet init zsh)"  at the end of ~/.zshrc
Setup (bash): add  eval "$(promptpet init bash)"  at the end of ~/.bashrc
Setup (fish): add  promptpet init fish | source  to ~/.config/fish/config.fish
```

### 4.2 Exit codes

| Code | When |
|---|---|
| 0 | Success. `prompt` and `_feed` **always** exit 0, even on internal errors. |
| 1 | Runtime error: state dir not writable, state file from a newer version, lock busy. |
| 2 | Usage error: unknown subcommand, missing or extra argument, bad shell, bad name, bad species. |

Errors go to stderr as `promptpet: <message>`. `prompt` and `_feed` never write to stderr.

Usage error messages (exact):
- unknown subcommand: `promptpet: unknown command "<x>" (see promptpet --help)`
- `init` with no or bad shell: `promptpet: init needs a shell: zsh, bash or fish`
- `name` without exactly one argument: `promptpet: usage: promptpet name <name>` (the name may contain spaces if quoted; it is one argv element)
- bad name: `promptpet: names can use letters, digits, spaces and - _ ' . (1-20 characters)`
- bad species: `promptpet: unknown species "<x>" (choose cat, dragon or blob)`
- `--exit` not an integer: ignored silently (treated as no exit code), because `prompt` never fails.

## 5. State

### 5.1 Location

State dir, in priority order:
1. `$PROMPTPET_HOME` if set and non-empty (tests always set this).
2. `$XDG_STATE_HOME/promptpet` if `XDG_STATE_HOME` is set and is an absolute path.
3. `$HOME/.local/state/promptpet`.

The dir is created with mode `0700` on first write. Files are written with mode `0600`.

| File | Writer | Content |
|---|---|---|
| `state.json` | foreground commands and `prompt` | the pet (§5.2) |
| `state.lock` | — | empty file used for `flock` around read-modify-write of `state.json` |
| `events.log` | `_feed` only (append) | one meal per line (§7.3) |
| `feed.json` | `_feed` only | per-repo last seen commit time (§7.2) |
| `feed.lock` | — | empty file used for `flock` by `_feed` |

### 5.2 `state.json`

```go
type State struct {
    Version     int       `json:"version"`      // always 1
    Name        string    `json:"name"`
    Species     string    `json:"species"`      // "cat" | "dragon" | "blob"
    Born        time.Time `json:"born"`
    Food        float64   `json:"food"`         // 0..100
    Fun         float64   `json:"fun"`          // 0..100
    Health      float64   `json:"health"`       // 0..100
    Away        bool      `json:"away"`
    AwaySince   time.Time `json:"away_since"`
    Adventures  int       `json:"adventures"`   // number of returns from exploring
    Meals       int       `json:"meals"`
    LastTick    time.Time `json:"last_tick"`    // decay applied up to here
    LastPrompt  time.Time `json:"last_prompt"`  // last `prompt` run; zero if never
    EventOffset int64     `json:"event_offset"` // bytes of events.log consumed
    Repo        RepoSig   `json:"repo"`         // last repo seen by `prompt`
    Message     string    `json:"message"`      // one-shot line for the next card
}

type RepoSig struct {
    GitDir  string `json:"git_dir"`
    Size    int64  `json:"size"`     // size of <gitdir>/logs/HEAD
    MtimeNs int64  `json:"mtime_ns"` // mtime of <gitdir>/logs/HEAD in Unix ns
}
```

New pet (no `state.json`): `Version 1, Name "Pip", Species "cat", Born now, Food 80, Fun 80, Health 100, LastTick now`, everything else zero.

### 5.3 Load, save, lock

- Read-modify-write of `state.json` always happens while holding an exclusive `syscall.Flock` on `state.lock`.
- Lock budget: `prompt` tries `LOCK_EX|LOCK_NB` every 500 µs for at most **5 ms**. Other commands try for at most **2 s**.
  - `prompt` on lock timeout: compute and print the segment from the loaded state, and do not save (the next prompt applies the same events again, so nothing is lost twice).
  - Other commands on lock timeout: exit 1 with `promptpet: another promptpet is busy; try again`.
- Save is atomic: write `state.json.tmp-<pid>`, `fsync` is not needed, then `os.Rename` over `state.json`.
- Missing file: new pet.
- Unparseable JSON: rename the file to `state.json.corrupt-<unix seconds>`, start a new pet, and (foreground commands only) print `promptpet: your pet's state file was unreadable; moved it to <path> and started fresh` to stderr. Exit code stays 0.
- `version` greater than 1: foreground commands exit 1 with `promptpet: state was written by a newer promptpet; please upgrade`. `prompt` prints nothing.

## 6. Pet rules (package `internal/pet`, pure, no I/O)

All functions take `now time.Time` (and `*rand.Rand` where random) as
arguments. Production passes `time.Now()` and a `rand.New(rand.NewPCG(seed1, seed2))`
seeded from `time.Now().UnixNano()`. Tests pass fixed values.

### 6.1 Constants

| Name | Value | Meaning |
|---|---|---|
| `FoodPerHour` | 2.0 | Food lost per hour (full to empty in 50 h) |
| `FunPerHour` | 3.0 | Fun lost per hour |
| `HealthPerHour` | 4.0 | Health regained per hour |
| `MealFood` | 20.0 | Food gained per commit |
| `MealFun` | 5.0 | Fun gained per commit |
| `FailHealth` | 6.0 | Health lost per failed command |
| `OKHealth` | 1.0 | Health gained per successful command |
| `PetFun`, `PetHealth` | 10.0, 2.0 | `promptpet pet` |
| `PlayFun` | 25.0 | `promptpet play` |
| `MaxTickHours` | 720.0 | cap on one tick (30 days) |
| `SleepyGap` | 20 min | gap between prompts that makes the pet sleepy |
| `SickBelow`, `HungryBelow`, `BoredBelow` | 40, 25, 25 | mood thresholds |
| return floors | Food 40, Fun 40, Health 60 | stats after coming back |
| `MaxMealsPerCheck` | 5 | per `_feed` run |

All stats are clamped to `[0, 100]` after every change.

### 6.2 Tick (decay over real elapsed time)

`func Tick(s *State, now time.Time)`:

```
h := now.Sub(s.LastTick).Hours()
if h < 0 { h = 0 }                // clock went backwards: no change
if h > MaxTickHours { h = MaxTickHours }
s.Food   = clamp(s.Food   - FoodPerHour*h)
s.Fun    = clamp(s.Fun    - FunPerHour*h)
s.Health = clamp(s.Health + HealthPerHour*h)
s.LastTick = now
if !s.Away && s.Food == 0 { s.Away = true; s.AwaySince = now }
```

Worked example (test values): Food 80, Fun 80, Health 50, 10 h later gives Food 60, Fun 50, Health 90.
From Food 80 the pet leaves after exactly 40 h without a meal.

### 6.3 Events

- `ApplyExit(s, code)`: `0` gives Health +1. `130` (Ctrl-C) and `148` (Ctrl-Z) change nothing. Any other code gives Health −6. Health only changes; exits never make the pet leave or return.
- `Meal(s, rng) (msg string)`: if Away, call `Return` first (and return its message). Then Food +20, Fun +5, Meals +1.
- `PetIt(s, rng) (lines []string)`: if Away, `Return` first and add its message as the first line. Then Fun +10, Health +2. Add the species line (§8.3).
- `Play(s, rng) (lines []string)`: if Away, `Return` first. Then Fun +25. Add `<Name> chases the <toy>!` with `toy = Toys[rng.IntN(5)]`.
- `Return(s, rng) string`: `Away=false`, `AwaySince` zero, `Food=max(Food,40)`, `Fun=max(Fun,40)`, `Health=max(Health,60)`, `Adventures+1`. Message `<Name> came back from an adventure with a <souvenir>!` with `souvenir = Souvenirs[rng.IntN(6)]`. Also stored in `s.Message` so the next card shows it.

`Toys = ["ball of yarn", "paper airplane", "laser dot", "rubber duck", "cardboard box"]`
`Souvenirs = ["shiny pebble", "pinecone", "lost semicolon", "tiny map", "feather", "rubber duck"]`

### 6.4 Hearts

`Hearts(s) int`: `score := (Food + Fun + Health) / 3`. `score < 17` gives 0,
`< 50` gives 1, `< 83` gives 2, otherwise 3. Rendered as `♥` repeated
`hearts` times followed by `♡` repeated `3-hearts` times.

### 6.5 Mood

```go
type Mood int
const (Content Mood = iota; Happy; Eating; Sick; Startled; Confused; Sleepy; Hungry; Bored; Away)

type MoodInput struct {
    Now        time.Time
    PrevPrompt time.Time // zero if none
    Exit       int       // -1 when there is no exit code (card)
    Ate        bool      // a meal was applied in this run
}
func MoodOf(s *State, in MoodInput) Mood
```

First match wins:
1. `s.Away` → Away
2. `in.Ate` → Eating
3. `Health < 40` → Sick
4. `Exit == 127` → Confused
5. `Exit > 0` and `Exit` not 130 or 148 → Startled
6. (`PrevPrompt` non-zero and `Now − PrevPrompt ≥ 20 min`) or `Now.Hour()` in 1..5 (01:00–05:59 in `Now`'s location) → Sleepy
7. `Food < 25` → Hungry
8. `Fun < 25` → Bored
9. `Hearts == 3` → Happy
10. Content

## 7. Commit feeding

### 7.1 Prompt side (no git, no exec except one detached spawn)

`repo.Find(cwd string) (worktree, gitDir string, ok bool)`: walk up from
`cwd` (at most 64 levels, stop at the filesystem root). At each dir `d`,
`os.Lstat(d/.git)`:
- a directory: `worktree=d`, `gitDir=d/.git`.
- a regular file whose first line is `gitdir: <p>`: `gitDir = p` (joined with `d` if relative). `worktree=d`.
- anything else: continue upward.

`repo.Sig(gitDir) (RepoSig, bool)`: `os.Stat(gitDir/logs/HEAD)`. Missing file
gives `false` (no commits yet, or reflog off; nothing to feed).

In `prompt` (§9), if `Find` and `Sig` succeed and the signature differs from
`state.Repo`, set `state.Repo` to it and spawn the feeder:
`exec.Command(os.Executable(), "_feed", worktree)` with stdin/stdout/stderr
nil, `SysProcAttr{Setsid: true}`, `Start()`, then `Process.Release()`. Never
wait. A spawn error is ignored. The spawner is injected so tests use a fake.

### 7.2 Feeder (`promptpet _feed <worktree>`, runs detached)

Overall deadline 5 s (`context.WithTimeout`). Git runs with the user's
environment plus `GIT_OPTIONAL_LOCKS=0`, `GIT_TERMINAL_PROMPT=0`, `LC_ALL=C`.
Any error ends the run silently with exit 0.

1. Take a blocking `flock` on `feed.lock` (inside the deadline).
2. `git -C <wt> rev-parse --path-format=absolute --git-common-dir` → `key` (trimmed). Worktrees of one repo share a key.
3. `git -C <wt> config --get user.email` → `email` (trimmed). Empty → stop.
4. `git -C <wt> log -n 50 --no-merges --format=%H%x09%at%x09%ae HEAD` → lines `sha<TAB>authorUnix<TAB>authorEmail`.
5. `last := feed.Repos[key]`; if the key is absent, `last = now.Unix() - 600` (a brand-new repo feeds commits from the last 10 minutes only, never history).
6. `matches` = lines where `strings.EqualFold(email, authorEmail)` and `authorUnix > last`. Author time is used so rebases, amends and cherry-picks of old work do not feed twice.
7. Take the newest `min(len(matches), 5)` matches and append them to `events.log` oldest first, one line each: `meal <authorUnix> <first 7 chars of sha>\n`, opened with `O_APPEND|O_CREATE|O_WRONLY`, mode `0600`.
8. `feed.Repos[key] = max(last, max authorUnix over all matches)`. If `Repos` has more than 200 keys, delete the keys with the smallest values until 200 remain (tie: lexically smallest key first). Save `feed.json` atomically (`{"version":1,"repos":{...}}`).

### 7.3 Consuming meals

`store.ReadMeals(dir string, offset int64) (count int, newOffset int64, err error)`:
open `events.log` (missing → `0, offset, nil`). If its size is smaller than
`offset`, return `0, size` (file was replaced; skip). Read from `offset` to the
end, at most 64 KiB. Count complete lines (ending in `\n`) that start with
`meal `. Unknown complete lines are skipped. `newOffset` is the offset just
after the last complete line; a partial last line is left for next time.

## 8. Output

### 8.1 Prompt segment

`Segment(s, mood) string` = `<face> <hearts>`, for example `(=^･ω･^=) ♥♥♡`.
When the mood is Away the segment is exactly `(out exploring)` (no hearts).

Faces (exact):

| Mood | cat | dragon | blob |
|---|---|---|---|
| Content | `(=^･ω･^=)` | `<(°ᴗ°)>` | `(•‿•)` |
| Happy | `(=^‿^=)` | `<(^ᴗ^)>` | `(^‿^)` |
| Eating | `(=^･ω･^=) nom` | `<(°ᴗ°)> nom` | `(•o•) nom` |
| Sick | `(=~ω~=) sniffle` | `<(~_~)> sniffle` | `(~_~) sniffle` |
| Startled | `(=°ω°=)!` | `<(°o°)>!` | `(°o°)!` |
| Confused | `(=･ω･=)?` | `<(°_°)>?` | `(•_•)?` |
| Sleepy | `(=-ω-=) zz` | `<(-ᴗ-)> zz` | `(-‿-) zz` |
| Hungry | `(=･ω･=) snack?` | `<(°ᴗ°)> snack?` | `(•‿•) snack?` |
| Bored | `(=¬ω¬=)` | `<(¬_¬)>` | `(¬_¬)` |

Safety: no segment may contain `%`, `$`, `` ` ``, `\`, `"`, `'`, or any rune
below U+0020. A test checks every species × mood.

### 8.2 Status card (`promptpet`)

Exact layout (cat, Content, example values):

```
 /\_/\
( o.o )
 > ^ <

Pip the cat - feeling content
food    ████████░░  80
fun     ██████░░░░  60
health  ██████████ 100
age 12 days · 34 meals · 3 adventures
```

- Lines 1–3: species art with every `E` replaced by the mood's eye. Line 4 is empty.
- Art (exact, `E` = eye):
  - cat: ` /\_/\` / `( E.E )` / ` > ^ <`
  - dragon: `  <\_/>` / ` ( E E )~~` / `  /|_|\`
  - blob: `  .---.` / ` ( E E )` / `  '---'`
  - Away (any species, no eyes): `  \     /` / `   \___/` / `  (empty basket)`
- Eyes: Content `o`, Happy `^`, Eating `^`, Sick `~`, Startled `O`, Confused `?`, Sleepy `-`, Hungry `o`, Bored `¬`.
- Title: `<Name> the <species> - <mood text>`.
- Bars: `fmt.Sprintf("%-7s %s %3d", label, bar, v)` with `v = int(math.Floor(stat))`, `bar` = 10 cells, `filled = (v+5)/10` cells of `█` then `░`.
- Footer: `age <d> day(s) · <m> meal(s) · <a> adventure(s)`, `d = floor(now − Born in days)`, singular when the number is 1.
- If `Message` is non-empty, print it on one more line, then clear it and save.

Mood text (exact): Content `feeling content`, Happy `feeling great`, Eating
`munching on your latest commit`, Sick `feeling a bit under the weather`,
Startled `a little startled`, Confused `confused by that command`, Sleepy
`sleepy`, Hungry `peckish - a commit would be lovely`, Bored
`a bit bored - try promptpet play`, Away
`out exploring - commit something or run promptpet pet to call them back`.

### 8.3 Interaction output (stdout, exit 0)

- `pet`: optional return line, then cat `<Name> purrs and headbutts your hand.`, dragon `<Name> puffs a happy little smoke ring.`, blob `<Name> wobbles contentedly.`
- `play`: optional return line, then `<Name> chases the <toy>!`
- `name <n>`: `Your pet is now called <n>.` Name rule: `strings.TrimSpace`, 1–20 runes, each rune a letter, digit, space, or one of `-_'.` (per `unicode.IsLetter`/`IsDigit`).
- `species <k>`: `<Name> is now a <k>.` Stats are kept.

## 9. `promptpet prompt` (the hot path)

Budget: **under 10 ms** wall time per call on a warm cache, with no network
and no git subprocess. Measured by `BenchmarkPrompt` and enforced by
`TestPromptBudget` (median of 20 runs < 10 ms; skipped with `-short`).

Steps, in order:
1. If `PROMPTPET_DISABLE` is non-empty: print nothing, exit 0.
2. Resolve the state dir. Try the lock (5 ms budget, §5.3).
3. Load state (new pet if missing; corrupt handled per §5.3; newer version → print nothing, exit 0).
4. `prev := s.LastPrompt`. `Tick(s, now)`.
5. `n, off := ReadMeals(dir, s.EventOffset)`; call `Meal` n times; `s.EventOffset = off`. `ate := n > 0`.
6. If `--exit` was given, `ApplyExit(s, code)`.
7. `mood := MoodOf(s, MoodInput{now, prev, exit or -1, ate})`.
8. `s.LastPrompt = now`. Repo check and maybe spawn (§7.1), only when the lock was taken (otherwise the new signature is not saved and every prompt would spawn).
9. Save (only if the lock was taken).
10. Print `Segment(s, mood)` followed by `\n`.

Any error at any step: print nothing (or the segment if one was already computed) and exit 0.

Foreground commands (`promptpet`, `pet`, `play`, `name`, `species`) also run
steps 2–5 before their action and save after it, but never set `LastPrompt`
and never spawn the feeder. The card uses `MoodInput{now, s.LastPrompt, -1, ate}`.

## 10. Shell hooks

Embedded files `internal/hooks/promptpet.{zsh,bash,fish}`, printed verbatim by
`promptpet init <shell>`. Rules: never print errors (`2>/dev/null`), always
preserve `$?`/`$status`, never re-add itself if the rc file is sourced twice.
Users who want the segment elsewhere use `$PROMPTPET_SEGMENT` (zsh/bash).

zsh (exact):
```zsh
# promptpet hook for zsh: eval "$(promptpet init zsh)"
_promptpet_precmd() {
  local __pp_exit=$?
  PROMPTPET_SEGMENT="$(command promptpet prompt --exit "$__pp_exit" 2>/dev/null)"
  return $__pp_exit
}
autoload -Uz add-zsh-hook
add-zsh-hook precmd _promptpet_precmd
setopt PROMPT_SUBST
case "$PROMPT" in
  *PROMPTPET_SEGMENT*) ;;
  *) PROMPT='${PROMPTPET_SEGMENT:+$PROMPTPET_SEGMENT }'"$PROMPT" ;;
esac
```

bash (exact):
```bash
# promptpet hook for bash: eval "$(promptpet init bash)"
_promptpet_precmd() {
  local __pp_exit=$?
  PROMPTPET_SEGMENT="$(command promptpet prompt --exit "$__pp_exit" 2>/dev/null)"
  return $__pp_exit
}
case ";${PROMPT_COMMAND:-};" in
  *";_promptpet_precmd;"*) ;;
  *) PROMPT_COMMAND="_promptpet_precmd${PROMPT_COMMAND:+;$PROMPT_COMMAND}" ;;
esac
case "$PS1" in
  *PROMPTPET_SEGMENT*) ;;
  *) PS1='${PROMPTPET_SEGMENT:+$PROMPTPET_SEGMENT }'"$PS1" ;;
esac
```

fish (exact):
```fish
# promptpet hook for fish: promptpet init fish | source
if not functions -q _promptpet_orig_prompt
    functions -q fish_prompt; and functions -c fish_prompt _promptpet_orig_prompt
    function _promptpet_status
        return $argv[1]
    end
    function fish_prompt
        set -l __pp_exit $status
        set -l __pp_seg (command promptpet prompt --exit $__pp_exit 2>/dev/null)
        test -n "$__pp_seg"; and printf '%s ' $__pp_seg
        if functions -q _promptpet_orig_prompt
            _promptpet_status $__pp_exit  # restore $status for the original prompt
            _promptpet_orig_prompt
        end
    end
end
```

Prompt expansion safety: zsh and bash do not re-expand the result of a
parameter expansion, and §8.1 forbids `%`, `$`, backslash and quotes in the
segment, so the segment cannot inject prompt escapes.

## 11. Package layout

```
promptpet/
├── main.go                   // os.Exit(cli.Run(os.Args[1:], cli.RealDeps(version)))
├── internal/
│   ├── pet/state.go          // State, RepoSig, New, constants, clamp
│   ├── pet/rules.go          // Tick, ApplyExit, Meal, PetIt, Play, Return
│   ├── pet/mood.go           // Mood, MoodInput, MoodOf, Hearts
│   ├── pet/render.go         // Segment, Card, faces, art, copy
│   ├── store/store.go        // Dir, Lock, LoadState, SaveState
│   ├── store/events.go       // ReadMeals, AppendMeals, LoadFeed, SaveFeed
│   ├── repo/repo.go          // Find, Sig
│   ├── feed/feed.go          // Run (the _feed command)
│   ├── hooks/hooks.go        // Script(shell) + embedded files
│   └── cli/{cli.go,prompt.go}
├── .goreleaser.yaml
├── .github/workflows/{ci.yml,release.yml}
├── README.md
└── LICENSE
```

`internal/pet` imports only the stdlib and no `os`. Only `internal/feed` and
the real spawner in `internal/cli` import `os/exec`.

## 12. Testing requirements

- Tests never touch the real HOME or real repos: every test that needs state sets `PROMPTPET_HOME` (or passes a dir) under `t.TempDir()`. Git tests set `HOME`, `XDG_CONFIG_HOME` and `GIT_CONFIG_GLOBAL` to temp paths and `GIT_CONFIG_NOSYSTEM=1`.
- Time and randomness are injected: tests use fixed `time.Date(..., time.UTC)` values and `rand.New(rand.NewPCG(1, 2))`.
- Unit tests for every rule in §6 with the worked examples.
- Every species × mood segment is checked for the forbidden characters in §8.1.
- Hook tests run the real shells (bash always; zsh and fish when installed, skipped otherwise) with a fake `promptpet` on `PATH` and assert `$?` is preserved and nothing reaches stderr, including when `promptpet` is missing from `PATH`.
- `_feed` tests build a real temp repo with fixed `GIT_AUTHOR_DATE`.
- Prompt budget: `TestPromptBudget` and `BenchmarkPrompt` (§9).
- CI runs `gofmt -l .` (must print nothing), `go vet ./...`, `go test ./...` on ubuntu-latest (with `zsh` and `fish` installed by apt) and macos-latest.

## 13. Definition of done (v0.1)

- Every command in §4 behaves as specified; `go test ./...`, `go vet ./...`, `gofmt -l .` are clean.
- `TestPromptBudget` passes locally.
- Manual check: `eval "$(go run . init zsh)"` in a real zsh, make a commit in a scratch repo, and see `nom` on the next or second prompt.
- README with install, setup per shell, what the pet reacts to, `PROMPTPET_DISABLE`, `PROMPTPET_HOME`, and the privacy note (only the current repo's recent commit times and short hashes are read; nothing leaves the machine).
- `.goreleaser.yaml` passes `goreleaser check` (if installed) and builds linux/darwin × amd64/arm64.

## 14. Later

- Optional ANSI color (respecting `NO_COLOR`), and an `RPROMPT` placement option.
- More species, accessories earned from adventures, an `adventures` log.
- `promptpet reset`, compaction of `events.log`, Windows/PowerShell support.
- Homebrew tap via GoReleaser `brews`.
