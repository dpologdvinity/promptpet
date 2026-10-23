# promptpet v0.1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `promptpet`, a Go CLI whose pet lives in the shell prompt, eats git commits detected by a detached background process, and renders its prompt segment in under 10 ms.

**Architecture:** A pure rules package (`internal/pet`) holds state, decay, mood and rendering with an injected clock and RNG. `internal/store` owns files under the state dir (atomic JSON, `flock`, an append-only meal log). `internal/repo` finds the git dir with `stat` only; `internal/feed` is the only code that runs git, and it runs detached. `internal/cli` wires commands with injected dependencies.

**Tech Stack:** Go (`go 1.26` in go.mod), stdlib only, GoReleaser v2, GitHub Actions.

**Spec:** `docs/SPEC.md`

## Global Constraints

- Module `github.com/dpologdvinity/promptpet`; `go.mod` line `go 1.26`; no third-party dependencies.
- License MIT, `Copyright (c) 2026 Kaitlyn Bassford`.
- State dir: `$PROMPTPET_HOME`, else `$XDG_STATE_HOME/promptpet` (absolute only), else `$HOME/.local/state/promptpet`; dir mode `0700`, files `0600`.
- Tests never touch the real HOME or real repos; every test uses `t.TempDir()` and sets `PROMPTPET_HOME`/`HOME`/`XDG_CONFIG_HOME`/`GIT_CONFIG_GLOBAL`/`GIT_CONFIG_NOSYSTEM=1` as needed.
- Time and randomness are always injected: tests use `time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)` as the base time `t0` and `rand.New(rand.NewPCG(1, 2))`.
- `promptpet prompt` and `promptpet _feed` always exit 0 and never write to stderr.
- Prompt path: under 10 ms, no network, no git subprocess; only the detached `_feed` spawn.
- Errors: `promptpet: <message>` on stderr; exit 1 runtime, 2 usage (SPEC §4.2).
- All user-facing copy is copied verbatim from SPEC §4.1, §8.
- Releases: linux/darwin × amd64/arm64, `CGO_ENABLED=0`; no Windows.
- Commits use Conventional Commits. Never commit `CLAUDE.md`, `.claude/`, `AGENTS.md`, `.mcp.json`, `.env`.

## Review Focus

- Two shells hitting the prompt at the same moment: the second must still print a segment within budget and must not corrupt `state.json` (test in Task 12: `TestPromptLockBusy`).
- The clock jumping backwards (NTP fix, VM resume): stats must not change and `LastTick` resets to now (test in Task 2: `TestTickClockBackwards`).
- A truncated or garbage `state.json` (power loss): the pet is recreated and the bad file is kept aside, never a crash or silent prompt failure (test in Task 7: `TestLoadStateCorrupt`).
- A half-written last line in `events.log` (feeder killed mid-write): it must not be counted until complete (test in Task 8: `TestReadMealsPartialLine`).
- The shell's cwd was deleted, or `.git` is a file (worktrees, submodules): the prompt still prints and finds the right git dir (tests in Task 9: `TestFindGitFile`, and Task 12: `TestPromptCwdGone`).

---

### Task 1: Scaffold, version and help

**Files:**
- Create: `go.mod`, `main.go`, `LICENSE`, `internal/cli/cli.go`, `.github/workflows/ci.yml`, `.github/workflows/release.yml`, `.goreleaser.yaml`, `.gitignore`
- Test: `internal/cli/cli_test.go`

**Interfaces:**
- Produces:
  ```go
  package cli
  type Deps struct {
      Version string
      Now     func() time.Time
      Rand    *rand.Rand              // math/rand/v2
      Getenv  func(string) string
      Getwd   func() (string, error)
      Spawn   func(worktree string) error
      Stdout  io.Writer
      Stderr  io.Writer
  }
  func Run(args []string, d Deps) int
  func RealDeps(version string) Deps // time.Now, PCG seeded from UnixNano, os.Getenv, os.Getwd, realSpawn, os.Stdout, os.Stderr
  ```
  Test helper in `cli_test.go` (used by Tasks 12–13): `func testDeps(t *testing.T) (Deps, *bytes.Buffer, *bytes.Buffer)` returning deps with `Now` fixed at `t0`, `Rand` PCG(1,2), `Getenv` reading a map that has `PROMPTPET_HOME=t.TempDir()`, `Getwd` returning `t.TempDir()`, and `Spawn` recording its argument in a slice the test can read.

- [ ] **Step 1: Write the failing test**

```go
func TestVersion(t *testing.T) {
    d, out, _ := testDeps(t)
    d.Version = "1.2.3"
    if code := Run([]string{"--version"}, d); code != 0 { t.Fatalf("exit %d", code) }
    if out.String() != "promptpet 1.2.3\n" { t.Fatalf("got %q", out.String()) }
}
func TestHelpAndUnknown(t *testing.T) {
    d, out, errb := testDeps(t)
    if Run([]string{"--help"}, d) != 0 || !strings.HasPrefix(out.String(), "promptpet - a tiny pet that lives in your shell prompt\n") { t.Fatal(out.String()) }
    if Run([]string{"bogus"}, d) != 2 || errb.String() != "promptpet: unknown command \"bogus\" (see promptpet --help)\n" { t.Fatal(errb.String()) }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cli/ -run 'TestVersion|TestHelpAndUnknown'`
Expected: FAIL (build error: `undefined: Run`)

- [ ] **Step 3: Implement**

`go mod init github.com/dpologdvinity/promptpet`, then set the `go` line to `go 1.26`. `main.go`: `var version = "dev"`; `func main() { os.Exit(cli.Run(os.Args[1:], cli.RealDeps(version))) }`. `Run` dispatches on `args[0]`: `--version`/`-V`, `--help`/`-h`/`help` (usage text verbatim from SPEC §4.1), else the unknown-command error. Leave other subcommands as unknown for now. `realSpawn` per SPEC §7.1 (`os.Executable`, `Setsid`, `Start`, `Release`).
`LICENSE`: standard MIT text with `Copyright (c) 2026 Kaitlyn Bassford`. `.gitignore`: `dist/` and `.env`.
`.github/workflows/ci.yml`: on push and pull_request; matrix `os: [ubuntu-latest, macos-latest]`; steps `actions/checkout@v7`, `actions/setup-go@v7` with `go-version-file: go.mod`, on ubuntu `sudo apt-get update && sudo apt-get install -y zsh fish`, then `test -z "$(gofmt -l .)"`, `go vet ./...`, `go test ./...`.
`.github/workflows/release.yml`: on push tags `v*`; `permissions: contents: write`; checkout with `fetch-depth: 0`, setup-go, `goreleaser/goreleaser-action@v7` with `version: "~> v2"`, `args: release --clean`, env `GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}`.
`.goreleaser.yaml`: `version: 2`, `project_name: promptpet`, one build with `env: [CGO_ENABLED=0]`, `goos: [linux, darwin]`, `goarch: [amd64, arm64]`, `ldflags: ["-s -w -X main.version={{.Version}}"]`; archives `formats: [tar.gz]` with `files: [LICENSE, README.md]`; `checksum.name_template: checksums.txt`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/cli/ && go vet ./... && test -z "$(gofmt -l .)"`
Expected: PASS, no output from gofmt

- [ ] **Step 5: Commit**

```bash
git add go.mod main.go LICENSE .gitignore internal/cli .github .goreleaser.yaml
git commit -m "chore: scaffold promptpet CLI with CI and release config"
```

---

### Task 2: Pet state and decay

**Files:**
- Create: `internal/pet/state.go`, `internal/pet/rules.go`
- Test: `internal/pet/rules_test.go`

**Interfaces:**
- Produces: `State`, `RepoSig` (fields and JSON tags exactly as SPEC §5.2); constants of SPEC §6.1 as exported `const`s with those names (`FoodPerHour`, `FunPerHour`, `HealthPerHour`, `MealFood`, `MealFun`, `FailHealth`, `OKHealth`, `PetFun`, `PetHealth`, `PlayFun`, `MaxTickHours`, `SleepyGap`, `SickBelow`, `HungryBelow`, `BoredBelow`, `MaxMealsPerCheck`); `func New(now time.Time) *State`; `func Tick(s *State, now time.Time)`; unexported `clamp(v float64) float64` to `[0,100]`.

- [ ] **Step 1: Write the failing tests**

```go
var t0 = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

func TestNew(t *testing.T) {
    s := New(t0)
    if s.Version != 1 || s.Name != "Pip" || s.Species != "cat" || s.Food != 80 || s.Fun != 80 || s.Health != 100 || !s.LastTick.Equal(t0) || !s.Born.Equal(t0) { t.Fatalf("%+v", s) }
}
func TestTickWorkedExample(t *testing.T) {
    s := New(t0); s.Health = 50
    Tick(s, t0.Add(10*time.Hour))
    if s.Food != 60 || s.Fun != 50 || s.Health != 90 || s.Away { t.Fatalf("%+v", s) }
}
func TestTickLeavesAfter40Hours(t *testing.T) {
    s := New(t0)
    Tick(s, t0.Add(40*time.Hour-time.Second))
    if s.Away { t.Fatal("left early") }
    Tick(s, t0.Add(40*time.Hour))
    if !s.Away || s.Food != 0 || !s.AwaySince.Equal(t0.Add(40*time.Hour)) { t.Fatalf("%+v", s) }
}
func TestTickClampsAndCaps(t *testing.T) {
    s := New(t0); s.Health = 10
    Tick(s, t0.Add(1000*24*time.Hour)) // capped at 720 h
    if s.Food != 0 || s.Fun != 0 || s.Health != 100 { t.Fatalf("%+v", s) }
}
func TestTickClockBackwards(t *testing.T) {
    s := New(t0)
    Tick(s, t0.Add(-5*time.Hour))
    if s.Food != 80 || s.Fun != 80 || !s.LastTick.Equal(t0.Add(-5*time.Hour)) { t.Fatalf("%+v", s) }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/pet/`
Expected: FAIL (build error: `undefined: New`)

- [ ] **Step 3: Implement `New` and `Tick` exactly as SPEC §5.2 and §6.2**

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/pet/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pet
git commit -m "feat(pet): add pet state and time-based decay"
```

---

### Task 3: Events, interactions and validation

**Files:**
- Modify: `internal/pet/rules.go`
- Test: `internal/pet/rules_test.go`

**Interfaces:**
- Consumes: `State`, `New`, constants (Task 2).
- Produces:
  ```go
  var Toys = []string{"ball of yarn", "paper airplane", "laser dot", "rubber duck", "cardboard box"}
  var Souvenirs = []string{"shiny pebble", "pinecone", "lost semicolon", "tiny map", "feather", "rubber duck"}
  var Species = []string{"cat", "dragon", "blob"}
  func ApplyExit(s *State, code int)
  func Return(s *State, rng *rand.Rand) string
  func Meal(s *State, rng *rand.Rand) string           // "" unless the pet returned
  func PetIt(s *State, rng *rand.Rand) []string
  func Play(s *State, rng *rand.Rand) []string
  func ValidName(n string) (string, bool)              // trimmed name, ok
  func ValidSpecies(k string) bool
  ```

- [ ] **Step 1: Write the failing tests**

```go
func rng() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

func TestApplyExit(t *testing.T) {
    s := New(t0); s.Health = 50
    ApplyExit(s, 0);   if s.Health != 51 { t.Fatal(s.Health) }
    ApplyExit(s, 130); ApplyExit(s, 148); if s.Health != 51 { t.Fatal(s.Health) }
    ApplyExit(s, 1);   if s.Health != 45 { t.Fatal(s.Health) }
    ApplyExit(s, 127); if s.Health != 39 { t.Fatal(s.Health) }
}
func TestMealCaps(t *testing.T) {
    s := New(t0); s.Food = 90; s.Fun = 98
    if msg := Meal(s, rng()); msg != "" { t.Fatal(msg) }
    if s.Food != 100 || s.Fun != 100 || s.Meals != 1 { t.Fatalf("%+v", s) }
}
func TestMealBringsPetBack(t *testing.T) {
    s := New(t0); s.Away = true; s.Food = 0; s.Fun = 10; s.Health = 30
    msg := Meal(s, rng())
    want := "Pip came back from an adventure with a " + Souvenirs[rng().IntN(6)] + "!"
    if msg != want || s.Message != want { t.Fatalf("%q", msg) }
    if s.Away || s.Food != 60 || s.Fun != 45 || s.Health != 60 || s.Adventures != 1 { t.Fatalf("%+v", s) }
}
func TestPetAndPlay(t *testing.T) {
    s := New(t0); s.Fun = 50; s.Health = 50
    if got := PetIt(s, rng()); len(got) != 1 || got[0] != "Pip purrs and headbutts your hand." { t.Fatal(got) }
    if s.Fun != 60 || s.Health != 52 { t.Fatalf("%+v", s) }
    s.Species = "dragon"
    if got := PetIt(s, rng()); got[0] != "Pip puffs a happy little smoke ring." { t.Fatal(got) }
    s.Species = "blob"
    if got := PetIt(s, rng()); got[0] != "Pip wobbles contentedly." { t.Fatal(got) }
    got := Play(s, rng())
    if got[0] != "Pip chases the "+Toys[rng().IntN(5)]+"!" || s.Fun != 100 { t.Fatal(got, s.Fun) }
    s.Away = true
    if got := Play(s, rng()); len(got) != 2 || !strings.HasPrefix(got[0], "Pip came back") { t.Fatal(got) }
}
func TestValidName(t *testing.T) {
    for _, ok := range []string{"Mochi", " Sir Pounce ", "Ünïcödé", "R2-D2", "O'Malley", "a.b_c"} {
        if _, v := ValidName(ok); !v { t.Errorf("%q rejected", ok) }
    }
    for _, bad := range []string{"", "   ", "123456789012345678901", "a%b", "a$b", "a`b", "a\\b", "a\x1b[31m", "a\nb", `a"b`} {
        if _, v := ValidName(bad); v { t.Errorf("%q accepted", bad) }
    }
    if n, _ := ValidName("  Pip  "); n != "Pip" { t.Fatal(n) }
    if !ValidSpecies("dragon") || ValidSpecies("dog") { t.Fatal("species") }
}
```

Note: `rng()` returns a fresh generator with the same seed, so the test computes the same index the code draws.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/pet/`
Expected: FAIL (build error: `undefined: ApplyExit`)

- [ ] **Step 3: Implement per SPEC §6.3 and §8.3**

`Return` draws exactly one `rng.IntN(6)`; `Play` draws exactly one `rng.IntN(5)` after any `Return`. Species lines from SPEC §8.3. `ValidName`: trim, 1–20 runes, each rune `unicode.IsLetter || unicode.IsDigit || strings.ContainsRune(" -_'.", r)`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/pet/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pet
git commit -m "feat(pet): add meals, exit reactions, petting, play and returning home"
```

---

### Task 4: Mood and hearts

**Files:**
- Create: `internal/pet/mood.go`
- Test: `internal/pet/mood_test.go`

**Interfaces:**
- Consumes: `State`, `New` (Task 2).
- Produces: `type Mood int` with constants `Content, Happy, Eating, Sick, Startled, Confused, Sleepy, Hungry, Bored, Away` (that order); `type MoodInput struct{ Now, PrevPrompt time.Time; Exit int; Ate bool }`; `func Hearts(s *State) int`; `func MoodOf(s *State, in MoodInput) Mood`.

- [ ] **Step 1: Write the failing tests**

```go
func TestHearts(t *testing.T) {
    cases := []struct{ f, fun, h float64; want int }{
        {0, 0, 50, 0},     // 16.67
        {0, 0, 51, 1},     // 17
        {50, 50, 49, 1},   // 49.67
        {50, 50, 50, 2},   // 50
        {80, 80, 88, 2},   // 82.67
        {80, 80, 89, 3},   // 83
    }
    for _, c := range cases {
        s := New(t0); s.Food, s.Fun, s.Health = c.f, c.fun, c.h
        if got := Hearts(s); got != c.want { t.Errorf("%v: got %d", c, got) }
    }
}
func TestMoodPriority(t *testing.T) {
    noon := t0 // 12:00 UTC
    base := func() *State { s := New(t0); s.Food, s.Fun, s.Health = 60, 60, 60; return s }
    in := MoodInput{Now: noon, Exit: -1}
    check := func(name string, s *State, in MoodInput, want Mood) {
        t.Helper(); if got := MoodOf(s, in); got != want { t.Errorf("%s: got %d want %d", name, got, want) }
    }
    s := base(); s.Away = true; check("away beats eating", s, MoodInput{Now: noon, Exit: 1, Ate: true}, Away)
    s = base(); s.Health = 10; check("eating beats sick", s, MoodInput{Now: noon, Exit: -1, Ate: true}, Eating)
    s = base(); s.Health = 39.9; check("sick beats exit", s, MoodInput{Now: noon, Exit: 1}, Sick)
    check("127 confused", base(), MoodInput{Now: noon, Exit: 127}, Confused)
    check("1 startled", base(), MoodInput{Now: noon, Exit: 1}, Startled)
    check("130 no reaction", base(), MoodInput{Now: noon, Exit: 130}, Content)
    check("gap sleepy", base(), MoodInput{Now: noon, PrevPrompt: noon.Add(-20 * time.Minute), Exit: 0}, Sleepy)
    check("short gap", base(), MoodInput{Now: noon, PrevPrompt: noon.Add(-19 * time.Minute), Exit: 0}, Content)
    check("01:00 sleepy", base(), MoodInput{Now: time.Date(2026, 3, 10, 1, 0, 0, 0, time.UTC), Exit: -1}, Sleepy)
    check("06:00 awake", base(), MoodInput{Now: time.Date(2026, 3, 10, 6, 0, 0, 0, time.UTC), Exit: -1}, Content)
    s = base(); s.Food = 24.9; s.Fun = 10; check("hungry beats bored", s, in, Hungry)
    s = base(); s.Fun = 24.9; check("bored", s, in, Bored)
    s = base(); s.Food, s.Fun, s.Health = 90, 90, 90; check("happy", s, in, Happy)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/pet/ -run 'TestHearts|TestMoodPriority'`
Expected: FAIL (build error: `undefined: Hearts`)

- [ ] **Step 3: Implement `Hearts` and `MoodOf` per SPEC §6.4 and §6.5**

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/pet/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pet
git commit -m "feat(pet): add mood priority and hearts"
```

---

### Task 5: Prompt segment

**Files:**
- Create: `internal/pet/render.go`
- Test: `internal/pet/render_test.go`

**Interfaces:**
- Consumes: `State`, `Mood`, `Hearts` (Tasks 2–4).
- Produces: `func Segment(s *State, m Mood) string`; `func Face(species string, m Mood) string`; `var Moods = []Mood{Content, Happy, Eating, Sick, Startled, Confused, Sleepy, Hungry, Bored}`.

- [ ] **Step 1: Write the failing tests**

```go
func TestSegmentExamples(t *testing.T) {
    s := New(t0); s.Food, s.Fun, s.Health = 60, 60, 60 // hearts 2
    if got := Segment(s, Content); got != "(=^･ω･^=) ♥♥♡" { t.Fatalf("%q", got) }
    s.Species = "dragon"
    if got := Segment(s, Eating); got != "<(°ᴗ°)> nom ♥♥♡" { t.Fatalf("%q", got) }
    s.Species = "blob"; s.Food, s.Fun, s.Health = 0, 0, 0
    if got := Segment(s, Startled); got != "(°o°)! ♡♡♡" { t.Fatalf("%q", got) }
    s.Away = true
    if got := Segment(s, Away); got != "(out exploring)" { t.Fatalf("%q", got) }
}
func TestSegmentSafeCharacters(t *testing.T) {
    for _, sp := range Species {
        for _, m := range append(Moods, Away) {
            s := New(t0); s.Species = sp
            seg := Segment(s, m)
            if seg == "" { t.Errorf("%s/%d empty", sp, m) }
            for _, r := range seg {
                if r < 0x20 || strings.ContainsRune("%$`\\\"'", r) { t.Errorf("%s/%d has %q", sp, m, r) }
            }
        }
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/pet/ -run TestSegment`
Expected: FAIL (build error: `undefined: Segment`)

- [ ] **Step 3: Implement with the face table copied verbatim from SPEC §8.1**

Store faces as `map[string][9]string` indexed by `Mood`. Unknown species falls back to `cat`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/pet/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pet
git commit -m "feat(pet): render the prompt segment for every species and mood"
```

---

### Task 6: Status card

**Files:**
- Modify: `internal/pet/render.go`
- Test: `internal/pet/render_test.go`

**Interfaces:**
- Consumes: `State`, `Mood` (Tasks 2–4).
- Produces: `func Card(s *State, m Mood, now time.Time) string` (ends with `\n`; includes the `Message` line when non-empty; does not clear it); `func MoodText(m Mood) string`.

- [ ] **Step 1: Write the failing tests**

```go
func TestCardExact(t *testing.T) {
    s := New(t0.Add(-12*24*time.Hour - time.Hour))
    s.Food, s.Fun, s.Health, s.Meals, s.Adventures = 80.9, 60, 100, 34, 3
    want := " /\\_/\\\n( o.o )\n > ^ <\n\n" +
        "Pip the cat - feeling content\n" +
        "food    ████████░░  80\n" +
        "fun     ██████░░░░  60\n" +
        "health  ██████████ 100\n" +
        "age 12 days · 34 meals · 3 adventures\n"
    if got := Card(s, Content, t0); got != want { t.Fatalf("got\n%s\nwant\n%s", got, want) }
}
func TestCardSingularAwayAndMessage(t *testing.T) {
    s := New(t0.Add(-25 * time.Hour)); s.Species = "dragon"; s.Meals, s.Adventures = 1, 1
    s.Food, s.Fun, s.Health = 4, 0, 45 // bars: 0 filled, 0 filled, 5 filled
    s.Message = "Pip came back from an adventure with a pinecone!"
    got := Card(s, Sleepy, t0)
    for _, line := range []string{"  <\\_/>", " ( - - )~~", "Pip the dragon - sleepy", "food    ░░░░░░░░░░   4", "health  █████░░░░░  45", "age 1 day · 1 meal · 1 adventure", "Pip came back from an adventure with a pinecone!"} {
        if !strings.Contains(got, line+"\n") { t.Errorf("missing %q in\n%s", line, got) }
    }
    s.Away = true
    if got := Card(s, Away, t0); !strings.Contains(got, "  (empty basket)\n") || !strings.Contains(got, "out exploring - commit something or run promptpet pet to call them back") { t.Fatal(got) }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/pet/ -run TestCard`
Expected: FAIL (build error: `undefined: Card`)

- [ ] **Step 3: Implement per SPEC §8.2** (art, eyes and mood text verbatim; bar format `fmt.Sprintf("%-7s %s %3d", label, bar, v)`)

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/pet/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pet
git commit -m "feat(pet): render the ASCII status card"
```

---

### Task 7: State dir, lock, load and save

**Files:**
- Create: `internal/store/store.go`
- Test: `internal/store/store_test.go`

**Interfaces:**
- Consumes: `pet.State`, `pet.New` (Task 2).
- Produces:
  ```go
  var ErrBusy = errors.New("another promptpet is busy; try again")
  var ErrNewer = errors.New("state was written by a newer promptpet; please upgrade")
  func Dir(getenv func(string) string) (string, error)                // SPEC §5.1; does not create
  func Lock(dir, name string, budget time.Duration) (unlock func(), err error) // creates dir 0700; LOCK_EX|LOCK_NB every 500µs; ErrBusy on timeout
  func LoadState(dir string, now time.Time) (s *pet.State, movedTo string, err error)
  func SaveState(dir string, s *pet.State) error
  func WriteAtomic(path string, data []byte) error                     // tmp-<pid> + rename, mode 0600
  ```

- [ ] **Step 1: Write the failing tests**

```go
func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func TestDir(t *testing.T) {
    if d, _ := Dir(env(map[string]string{"PROMPTPET_HOME": "/x/pp", "XDG_STATE_HOME": "/s", "HOME": "/h"})); d != "/x/pp" { t.Fatal(d) }
    if d, _ := Dir(env(map[string]string{"XDG_STATE_HOME": "/s", "HOME": "/h"})); d != "/s/promptpet" { t.Fatal(d) }
    if d, _ := Dir(env(map[string]string{"XDG_STATE_HOME": "rel", "HOME": "/h"})); d != "/h/.local/state/promptpet" { t.Fatal(d) }
    if _, err := Dir(env(map[string]string{})); err == nil { t.Fatal("want error without HOME") }
}
func TestSaveLoadRoundTrip(t *testing.T) {
    dir := filepath.Join(t.TempDir(), "pp")
    unlock, err := Lock(dir, "state.lock", time.Second); if err != nil { t.Fatal(err) }
    defer unlock()
    s := pet.New(t0); s.Name = "Mochi"; s.Food = 33.5
    if err := SaveState(dir, s); err != nil { t.Fatal(err) }
    got, moved, err := LoadState(dir, t0.Add(time.Hour))
    if err != nil || moved != "" || got.Name != "Mochi" || got.Food != 33.5 { t.Fatalf("%+v %q %v", got, moved, err) }
    if fi, _ := os.Stat(dir); fi.Mode().Perm() != 0o700 { t.Fatal(fi.Mode()) }
    if fi, _ := os.Stat(filepath.Join(dir, "state.json")); fi.Mode().Perm() != 0o600 { t.Fatal(fi.Mode()) }
}
func TestLoadStateMissing(t *testing.T) {
    s, moved, err := LoadState(t.TempDir(), t0)
    if err != nil || moved != "" || s.Name != "Pip" { t.Fatal(s, moved, err) }
}
func TestLoadStateCorrupt(t *testing.T) {
    dir := t.TempDir()
    os.WriteFile(filepath.Join(dir, "state.json"), []byte(`{"version":1,"name":"Mo`), 0o600)
    s, moved, err := LoadState(dir, t0)
    if err != nil || s.Name != "Pip" { t.Fatal(s, err) }
    if moved != filepath.Join(dir, fmt.Sprintf("state.json.corrupt-%d", t0.Unix())) { t.Fatal(moved) }
    if _, err := os.Stat(moved); err != nil { t.Fatal(err) }
}
func TestLoadStateNewer(t *testing.T) {
    dir := t.TempDir()
    os.WriteFile(filepath.Join(dir, "state.json"), []byte(`{"version":2}`), 0o600)
    if _, _, err := LoadState(dir, t0); !errors.Is(err, ErrNewer) { t.Fatal(err) }
}
func TestLockBusy(t *testing.T) {
    dir := t.TempDir()
    unlock, _ := Lock(dir, "state.lock", time.Second)
    defer unlock()
    start := time.Now()
    if _, err := Lock(dir, "state.lock", 5*time.Millisecond); !errors.Is(err, ErrBusy) { t.Fatal(err) }
    if time.Since(start) > 50*time.Millisecond { t.Fatal("lock wait exceeded budget") }
}
```

Note: `flock` locks are per open file description, so the second `Lock` in the same process opens the file again and is refused, which is what this test relies on.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/store/`
Expected: FAIL (build error: `undefined: Dir`)

- [ ] **Step 3: Implement per SPEC §5.1 and §5.3**

`Dir` returns an error `cannot find a home directory` if all three sources are empty. Unlock closes the file descriptor (releasing the lock).

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/store/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/store
git commit -m "feat(store): add state dir, flock, atomic save and corrupt-file recovery"
```

---

### Task 8: Meal log and feed file

**Files:**
- Create: `internal/store/events.go`
- Test: `internal/store/events_test.go`

**Interfaces:**
- Consumes: `WriteAtomic` (Task 7).
- Produces:
  ```go
  type MealLine struct { Unix int64; SHA string } // SHA already 7 chars
  func AppendMeals(dir string, meals []MealLine) error // "meal <unix> <sha>\n" each, O_APPEND|O_CREATE|O_WRONLY 0600
  func ReadMeals(dir string, offset int64) (count int, newOffset int64, err error) // SPEC §7.3
  type Feed struct { Version int `json:"version"`; Repos map[string]int64 `json:"repos"` }
  func LoadFeed(dir string) (*Feed, error) // missing or corrupt -> empty Feed{Version:1, Repos:{}}
  func SaveFeed(dir string, f *Feed) error // prunes to 200 keys per SPEC §7.2 step 8, then WriteAtomic
  ```

- [ ] **Step 1: Write the failing tests**

```go
func TestMealsAppendAndRead(t *testing.T) {
    dir := t.TempDir()
    if n, off, err := ReadMeals(dir, 0); n != 0 || off != 0 || err != nil { t.Fatal(n, off, err) }
    AppendMeals(dir, []MealLine{{100, "abcdef1"}, {200, "abcdef2"}})
    n, off, _ := ReadMeals(dir, 0)
    if n != 2 || off != int64(len("meal 100 abcdef1\nmeal 200 abcdef2\n")) { t.Fatal(n, off) }
    if n, off2, _ := ReadMeals(dir, off); n != 0 || off2 != off { t.Fatal(n, off2) }
}
func TestReadMealsPartialLine(t *testing.T) {
    dir := t.TempDir()
    os.WriteFile(filepath.Join(dir, "events.log"), []byte("meal 1 a\nnoise line\nmeal 2 b"), 0o600)
    n, off, _ := ReadMeals(dir, 0)
    if n != 1 || off != int64(len("meal 1 a\nnoise line\n")) { t.Fatal(n, off) }
}
func TestReadMealsTruncatedFile(t *testing.T) {
    dir := t.TempDir()
    os.WriteFile(filepath.Join(dir, "events.log"), []byte("meal 1 a\n"), 0o600)
    if n, off, _ := ReadMeals(dir, 500); n != 0 || off != 9 { t.Fatal(n, off) }
}
func TestFeedPrune(t *testing.T) {
    dir := t.TempDir()
    f, _ := LoadFeed(dir)
    for i := 0; i < 205; i++ { f.Repos[fmt.Sprintf("/r/%03d", i)] = int64(i % 203) } // i=203 gets 0 and i=204 gets 1, so the 5 smallest are 000, 203, 001, 204, 002
    SaveFeed(dir, f)
    g, _ := LoadFeed(dir)
    if len(g.Repos) != 200 { t.Fatal(len(g.Repos)) }
    for _, gone := range []string{"/r/000", "/r/203", "/r/001", "/r/204", "/r/002"} {
        if _, ok := g.Repos[gone]; ok { t.Errorf("%s kept", gone) }
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/store/ -run 'Meals|Feed'`
Expected: FAIL (build error: `undefined: AppendMeals`)

- [ ] **Step 3: Implement per SPEC §7.2–§7.3**

Prune order: sort keys by `(value asc, key asc)` and delete from the front until 200 remain.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/store/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/store
git commit -m "feat(store): add append-only meal log and feed bookkeeping"
```

---

### Task 9: Find the git dir without running git

**Files:**
- Create: `internal/repo/repo.go`
- Test: `internal/repo/repo_test.go`

**Interfaces:**
- Consumes: `pet.RepoSig` (Task 2).
- Produces: `func Find(cwd string) (worktree, gitDir string, ok bool)`; `func Sig(gitDir string) (pet.RepoSig, bool)`.

- [ ] **Step 1: Write the failing tests** (build fake repos with `os.MkdirAll`/`os.WriteFile`; no git needed)

```go
func TestFindDirAndSig(t *testing.T) {
    root := t.TempDir()
    os.MkdirAll(filepath.Join(root, ".git", "logs"), 0o755)
    sub := filepath.Join(root, "a", "b"); os.MkdirAll(sub, 0o755)
    wt, gd, ok := Find(sub)
    if !ok || wt != root || gd != filepath.Join(root, ".git") { t.Fatal(wt, gd, ok) }
    if _, ok := Sig(gd); ok { t.Fatal("no logs/HEAD yet") }
    os.WriteFile(filepath.Join(gd, "logs", "HEAD"), []byte("x\n"), 0o644)
    sig, ok := Sig(gd)
    if !ok || sig.GitDir != gd || sig.Size != 2 || sig.MtimeNs == 0 { t.Fatal(sig) }
}
func TestFindGitFile(t *testing.T) {
    root := t.TempDir()
    real := filepath.Join(root, "main", ".git", "worktrees", "wt"); os.MkdirAll(real, 0o755)
    wt := filepath.Join(root, "wt"); os.MkdirAll(wt, 0o755)
    os.WriteFile(filepath.Join(wt, ".git"), []byte("gitdir: ../main/.git/worktrees/wt\n"), 0o644)
    gotWt, gd, ok := Find(wt)
    if !ok || gotWt != wt || gd != real { t.Fatal(gotWt, gd, ok) }
}
func TestFindNone(t *testing.T) {
    if _, _, ok := Find(t.TempDir()); ok { t.Skip("a parent of the temp dir is a git repo") }
    if _, _, ok := Find(filepath.Join(t.TempDir(), "does", "not", "exist")); ok { t.Fatal("found repo in missing dir") }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/repo/`
Expected: FAIL (build error: `undefined: Find`)

- [ ] **Step 3: Implement per SPEC §7.1** (relative `gitdir:` paths are joined with the dir holding `.git` and cleaned with `filepath.Clean`; at most 64 levels)

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/repo/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/repo
git commit -m "feat(repo): locate the git dir and reflog signature with stat only"
```

---

### Task 10: Background commit feeder

**Files:**
- Create: `internal/feed/feed.go`
- Test: `internal/feed/feed_test.go`

**Interfaces:**
- Consumes: `store.Lock`, `store.LoadFeed`, `store.SaveFeed`, `store.AppendMeals`, `store.ReadMeals`, `store.MealLine` (Tasks 7–8).
- Produces: `func Run(ctx context.Context, dir, worktree string, now time.Time) error`.

- [ ] **Step 1: Write the failing tests**

Test helper `gitRepo(t) string`: create `t.TempDir()`, set `t.Setenv` for `HOME`, `XDG_CONFIG_HOME` (temp), `GIT_CONFIG_GLOBAL` (temp empty file), `GIT_CONFIG_NOSYSTEM=1`; run `git init -q -b main`, `git config user.email me@example.com`, `git config user.name Me`, `git config commit.gpgsign false`. Helper `commit(t, dir, unix int64, email string)` runs `git commit -q --allow-empty -m c` with env `GIT_AUTHOR_DATE=@<unix> +0000`, `GIT_COMMITTER_DATE=@<unix> +0000`, `GIT_AUTHOR_NAME=Me`, `GIT_AUTHOR_EMAIL=<email>`, `GIT_COMMITTER_NAME=Me`, `GIT_COMMITTER_EMAIL=<email>` (so commits work even after `user.email` is unset).

```go
func TestFeedNewRepoOnlyRecent(t *testing.T) {
    wt := gitRepo(t); state := t.TempDir()
    now := time.Unix(1_800_000_000, 0)
    commit(t, wt, now.Unix()-3600, "me@example.com") // older than 10 min: not fed
    commit(t, wt, now.Unix()-60, "ME@example.com")   // case-insensitive match: fed
    commit(t, wt, now.Unix()-30, "other@example.com") // someone else: not fed
    if err := Run(context.Background(), state, wt, now); err != nil { t.Fatal(err) }
    if n, _, _ := store.ReadMeals(state, 0); n != 1 { t.Fatal(n) }
    f, _ := store.LoadFeed(state)
    for _, v := range f.Repos { if v != now.Unix()-60 { t.Fatal(v) } }
}
func TestFeedNoDoubleCountAndCap(t *testing.T) {
    wt := gitRepo(t); state := t.TempDir()
    now := time.Unix(1_800_000_000, 0)
    Run(context.Background(), state, wt, now) // no commits yet: no error, nothing fed
    for i := 0; i < 7; i++ { commit(t, wt, now.Unix()-300+int64(i), "me@example.com") }
    Run(context.Background(), state, wt, now)
    Run(context.Background(), state, wt, now) // same commits again
    if n, _, _ := store.ReadMeals(state, 0); n != 5 { t.Fatal(n) }
    b, _ := os.ReadFile(filepath.Join(state, "events.log"))
    if !strings.HasPrefix(string(b), fmt.Sprintf("meal %d ", now.Unix()-298)) { t.Fatalf("oldest-first of newest five: %s", b) }
}
func TestFeedNoEmail(t *testing.T) {
    wt := gitRepo(t); state := t.TempDir()
    exec.Command("git", "-C", wt, "config", "--unset", "user.email").Run()
    commit(t, wt, 1_800_000_000-10, "me@example.com")
    if err := Run(context.Background(), state, wt, time.Unix(1_800_000_000, 0)); err != nil { t.Fatal(err) }
    if n, _, _ := store.ReadMeals(state, 0); n != 0 { t.Fatal(n) }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/feed/`
Expected: FAIL (build error: `undefined: Run`)

- [ ] **Step 3: Implement per SPEC §7.2**

Git via `exec.CommandContext(ctx, "git", "-C", worktree, ...)` with `cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0", "GIT_TERMINAL_PROMPT=0", "LC_ALL=C")`. "No commits yet" and "no email" return `nil`. Lock `feed.lock` with `store.Lock(dir, "feed.lock", 5*time.Second)`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/feed/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/feed
git commit -m "feat(feed): detect new commits by author time in a background run"
```

---

### Task 11: Shell hooks

**Files:**
- Create: `internal/hooks/hooks.go`, `internal/hooks/promptpet.zsh`, `internal/hooks/promptpet.bash`, `internal/hooks/promptpet.fish`
- Test: `internal/hooks/hooks_test.go`

**Interfaces:**
- Produces: `func Script(shell string) (string, bool)` (`zsh`, `bash`, `fish`; files via `//go:embed`).

- [ ] **Step 1: Write the failing tests**

Helper `fakeBin(t, script string) string` writes an executable `promptpet` shell script into a temp dir and returns the dir. The fake prints `SEG:$3` (its third argument is the exit code).

```go
func TestBashHook(t *testing.T) {
    hook, _ := Script("bash"); f := filepath.Join(t.TempDir(), "h.bash"); os.WriteFile(f, []byte(hook), 0o644)
    bin := fakeBin(t, "#!/bin/sh\necho \"SEG:$3\"\n")
    out, errOut := runShell(t, "bash", []string{"--norc", "-c", `PS1='$ '; source "$1"; source "$1"; (exit 7); _promptpet_precmd; echo "rc=$? seg=$PROMPTPET_SEGMENT pc=$PROMPT_COMMAND"; echo "$PS1"`, "x", f}, bin+":/usr/bin:/bin")
    want := "rc=7 seg=SEG:7 pc=_promptpet_precmd\n${PROMPTPET_SEGMENT:+$PROMPTPET_SEGMENT }$ \n"
    if out != want || errOut != "" { t.Fatalf("out=%q err=%q", out, errOut) }
}
func TestBashHookMissingBinary(t *testing.T) {
    hook, _ := Script("bash"); f := filepath.Join(t.TempDir(), "h.bash"); os.WriteFile(f, []byte(hook), 0o644)
    out, errOut := runShell(t, "bash", []string{"--norc", "-c", `source "$1"; (exit 3); _promptpet_precmd; echo "rc=$? seg=[$PROMPTPET_SEGMENT]"`, "x", f}, t.TempDir())
    if out != "rc=3 seg=[]\n" || errOut != "" { t.Fatalf("out=%q err=%q", out, errOut) }
}
```

Also `TestZshHook` (skip if `exec.LookPath("zsh")` fails): same as bash with `zsh -f -c`, `PROMPT='%# '`, expecting `rc=7 seg=SEG:7` and `PROMPT` = `${PROMPTPET_SEGMENT:+$PROMPTPET_SEGMENT }%# `.
And `TestFishHook` (skip without fish): `fish --no-config -c 'function fish_prompt; echo -n "orig:$status"; end; source $argv[1]; source $argv[1]; false; fish_prompt' <hook>` expects stdout `SEG:1 orig:1` and empty stderr.
`runShell(t, shell, args, path)` runs with env `PATH=<path>`, `HOME=<t.TempDir()>` and returns stdout and stderr as strings. `Script("tcsh")` returns `"", false`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/hooks/`
Expected: FAIL (build error: `undefined: Script`)

- [ ] **Step 3: Write the three hook files verbatim from SPEC §10 and implement `Script`**

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -v ./internal/hooks/`
Expected: PASS (zsh/fish tests PASS or SKIP when the shell is not installed)

- [ ] **Step 5: Commit**

```bash
git add internal/hooks
git commit -m "feat(hooks): add zsh, bash and fish prompt hooks that preserve exit status"
```

---

### Task 12: `promptpet prompt` hot path

**Files:**
- Create: `internal/cli/prompt.go`
- Modify: `internal/cli/cli.go` (dispatch `prompt`)
- Test: `internal/cli/prompt_test.go`

**Interfaces:**
- Consumes: `pet.Tick`, `pet.Meal`, `pet.ApplyExit`, `pet.MoodOf`, `pet.Segment` (Tasks 2–5); `store.Dir`, `store.Lock`, `store.LoadState`, `store.SaveState`, `store.ReadMeals`, `store.AppendMeals` (Tasks 7–8); `repo.Find`, `repo.Sig` (Task 9); `Deps`, `testDeps` (Task 1).
- Produces: `func runPrompt(args []string, d Deps) int` (always returns 0).

- [ ] **Step 1: Write the failing tests**

```go
func TestPromptFirstRun(t *testing.T) {
    d, out, errb := testDeps(t)
    if Run([]string{"prompt", "--exit", "0"}, d) != 0 { t.Fatal() }
    if out.String() != "(=^‿^=) ♥♥♥\n" || errb.Len() != 0 { t.Fatalf("%q %q", out, errb) } // 80/80/100 -> 86.7 -> 3 hearts, Happy at noon
}
func TestPromptMealAndExit(t *testing.T) {
    d, out, _ := testDeps(t)
    dir := d.Getenv("PROMPTPET_HOME")
    Run([]string{"prompt"}, d); out.Reset()
    store.AppendMeals(dir, []store.MealLine{{1, "aaaaaaa"}})
    Run([]string{"prompt", "--exit", "1"}, d)
    if !strings.HasPrefix(out.String(), "(=^･ω･^=) nom ") { t.Fatalf("%q", out) }
    out.Reset(); Run([]string{"prompt", "--exit", "127"}, d)
    if !strings.HasPrefix(out.String(), "(=･ω･=)? ") { t.Fatalf("%q", out) } // meal consumed once
    s, _, _ := store.LoadState(dir, t0)
    if s.Meals != 1 || s.Health != 88 { t.Fatalf("%+v", s) } // 100 -6 -6
}
func TestPromptSpawnsOnRepoChange(t *testing.T) {
    d, _, _ := testDeps(t)
    wt := t.TempDir(); os.MkdirAll(filepath.Join(wt, ".git", "logs"), 0o755)
    os.WriteFile(filepath.Join(wt, ".git", "logs", "HEAD"), []byte("a\n"), 0o644)
    d.Getwd = func() (string, error) { return wt, nil }
    Run([]string{"prompt"}, d); Run([]string{"prompt"}, d)
    if got := spawned(d); len(got) != 1 || got[0] != wt { t.Fatal(got) } // unchanged sig: no second spawn
    os.WriteFile(filepath.Join(wt, ".git", "logs", "HEAD"), []byte("a\nb\n"), 0o644)
    Run([]string{"prompt"}, d)
    if got := spawned(d); len(got) != 2 { t.Fatal(got) }
}
func TestPromptCwdGone(t *testing.T) {
    d, out, _ := testDeps(t)
    d.Getwd = func() (string, error) { return "", os.ErrNotExist }
    if Run([]string{"prompt"}, d) != 0 || out.Len() == 0 { t.Fatal("no segment") }
}
func TestPromptLockBusy(t *testing.T) {
    d, out, _ := testDeps(t)
    unlock, _ := store.Lock(d.Getenv("PROMPTPET_HOME"), "state.lock", time.Second); defer unlock()
    start := time.Now()
    Run([]string{"prompt"}, d)
    if out.Len() == 0 || time.Since(start) > 50*time.Millisecond { t.Fatal("must print within budget") }
    if _, err := os.Stat(filepath.Join(d.Getenv("PROMPTPET_HOME"), "state.json")); err == nil { t.Fatal("saved without lock") }
}
func TestPromptDisabledAndCorrupt(t *testing.T) {
    d, out, errb := testDeps(t)
    os.WriteFile(filepath.Join(d.Getenv("PROMPTPET_HOME"), "state.json"), []byte("{garbage"), 0o600)
    Run([]string{"prompt", "--exit", "x"}, d)
    if out.Len() == 0 || errb.Len() != 0 { t.Fatalf("%q %q", out, errb) }
    out.Reset(); d.Getenv = withEnv(d.Getenv, "PROMPTPET_DISABLE", "1")
    if Run([]string{"prompt"}, d) != 0 || out.Len() != 0 { t.Fatal(out) }
}
func TestPromptBudget(t *testing.T) {
    if testing.Short() { t.Skip() }
    d, _, _ := testDeps(t)
    d.Stdout = io.Discard
    var times []time.Duration
    for i := 0; i < 20; i++ { s := time.Now(); Run([]string{"prompt", "--exit", "0"}, d); times = append(times, time.Since(s)) }
    slices.Sort(times)
    if times[10] >= 10*time.Millisecond { t.Fatalf("median %v >= 10ms", times[10]) }
}
func BenchmarkPrompt(b *testing.B) { /* same deps with b.TempDir(); loop Run(prompt) */ }
```

`spawned(d)` and `withEnv(getenv, k, v)` are small helpers added to `cli_test.go` next to `testDeps` (`testDeps` stores spawn calls in a slice reachable from the returned deps, for example via a package-level map keyed by the state dir).

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/cli/ -run Prompt`
Expected: FAIL (`unknown command "prompt"`)

- [ ] **Step 3: Implement `runPrompt` following SPEC §9 steps 1–10 in order**

Parse `--exit` by hand (`strconv.Atoi` on the value after `--exit`; on error use -1). Swallow every error. When `Getwd` fails, skip the repo check. Spawn only when the lock was taken (otherwise the new signature is not saved and every prompt would spawn).

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/cli/ && go test -bench Prompt -run '^$' ./internal/cli/`
Expected: PASS; benchmark reports well under 10 ms/op

- [ ] **Step 5: Commit**

```bash
git add internal/cli
git commit -m "feat(cli): add the fast prompt command with background feeding"
```

---

### Task 13: Card, interactions, init and `_feed`

**Files:**
- Modify: `internal/cli/cli.go`
- Test: `internal/cli/cli_test.go`

**Interfaces:**
- Consumes: everything above; `hooks.Script` (Task 11); `feed.Run` (Task 10).
- Produces: dispatch for `""` (no args: card), `init`, `pet`, `play`, `name`, `species`, `_feed`.

- [ ] **Step 1: Write the failing tests**

```go
func TestCardCommand(t *testing.T) {
    d, out, _ := testDeps(t)
    if Run(nil, d) != 0 || !strings.Contains(out.String(), "Pip the cat - feeling great\n") { t.Fatal(out) }
}
func TestInteractions(t *testing.T) {
    d, out, errb := testDeps(t)
    if Run([]string{"name", "  Mochi "}, d) != 0 || out.String() != "Your pet is now called Mochi.\n" { t.Fatal(out) }
    out.Reset(); Run([]string{"species", "blob"}, d)
    if out.String() != "Mochi is now a blob.\n" { t.Fatal(out) }
    out.Reset(); Run([]string{"pet"}, d)
    if out.String() != "Mochi wobbles contentedly.\n" { t.Fatal(out) }
    if Run([]string{"species", "dog"}, d) != 2 || errb.String() != "promptpet: unknown species \"dog\" (choose cat, dragon or blob)\n" { t.Fatal(errb) }
    errb.Reset()
    if Run([]string{"name", "a$b"}, d) != 2 || errb.String() != "promptpet: names can use letters, digits, spaces and - _ ' . (1-20 characters)\n" { t.Fatal(errb) }
    errb.Reset()
    if Run([]string{"name"}, d) != 2 || errb.String() != "promptpet: usage: promptpet name <name>\n" { t.Fatal(errb) }
}
func TestCardShowsMessageOnce(t *testing.T) {
    d, out, _ := testDeps(t)
    dir := d.Getenv("PROMPTPET_HOME")
    s := pet.New(t0); s.Away = true; s.Food = 0
    unlock, _ := store.Lock(dir, "state.lock", time.Second); store.SaveState(dir, s); unlock()
    Run([]string{"pet"}, d)
    out.Reset(); Run(nil, d)
    if !strings.Contains(out.String(), "came back from an adventure") { t.Fatal(out) }
    out.Reset(); Run(nil, d)
    if strings.Contains(out.String(), "came back from an adventure") { t.Fatal("message shown twice") }
}
func TestInit(t *testing.T) {
    d, out, errb := testDeps(t)
    want, _ := hooks.Script("fish")
    if Run([]string{"init", "fish"}, d) != 0 || out.String() != want { t.Fatal(out) }
    if Run([]string{"init", "tcsh"}, d) != 2 || errb.String() != "promptpet: init needs a shell: zsh, bash or fish\n" { t.Fatal(errb) }
}
func TestCorruptStateWarns(t *testing.T) {
    d, _, errb := testDeps(t)
    os.WriteFile(filepath.Join(d.Getenv("PROMPTPET_HOME"), "state.json"), []byte("{"), 0o600)
    if Run(nil, d) != 0 || !strings.HasPrefix(errb.String(), "promptpet: your pet's state file was unreadable; moved it to ") { t.Fatal(errb) }
}
func TestFeedCommandSilent(t *testing.T) {
    d, out, errb := testDeps(t)
    if Run([]string{"_feed", t.TempDir()}, d) != 0 || out.Len()+errb.Len() != 0 { t.Fatal(out, errb) } // not a repo: silent
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/cli/`
Expected: FAIL (`unknown command "name"` and others)

- [ ] **Step 3: Implement**

A shared helper `withState(d Deps, fn func(s *pet.State, ate bool) error) int` does SPEC §9 steps 2–5 with a 2 s lock budget, prints the corrupt-file warning, maps `ErrBusy`/`ErrNewer` to exit 1 with `promptpet: ` + the error text, calls `fn`, then saves. The card clears `s.Message` after printing. `_feed` calls `feed.Run(ctx with 5 s timeout, dir, args[1], d.Now())`, ignores the error, prints nothing and returns 0.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cli
git commit -m "feat(cli): add status card, pet, play, name, species, init and _feed"
```

---

### Task 14: README and release check

**Files:**
- Create: `README.md`

**Interfaces:**
- Consumes: the finished CLI.

- [ ] **Step 1: Write README.md** with these sections: one-line pitch and a segment example; Install (`go install github.com/dpologdvinity/promptpet@latest`, release archives); Setup for zsh, bash and fish (the three lines from SPEC §4.1, and "put it at the end of your rc file so it wraps your final prompt"); What your pet reacts to (commits, failures, idle time, Ctrl-C ignored, exploring and coming back, all from SPEC §6); Commands (copy SPEC §4.1 usage); Privacy (reads only the current repo's last 50 commit times, author emails and short hashes, in a background process; stores only times, short hashes and the repo path under the state dir; no network); Environment (`PROMPTPET_HOME`, `PROMPTPET_DISABLE`, `$PROMPTPET_SEGMENT` for custom placement); License (MIT).

- [ ] **Step 2: Full verification**

Run: `test -z "$(gofmt -l .)" && go vet ./... && go test ./... && go build -o /tmp/promptpet . && PROMPTPET_HOME=$(mktemp -d) /tmp/promptpet`
Expected: all pass; the last command prints a cat card titled `Pip the cat - feeling great`.
If `goreleaser` is installed: `goreleaser check` prints no errors.

- [ ] **Step 3: Manual check (SPEC §13)**

In a scratch dir: `PROMPTPET_HOME=$(mktemp -d) zsh -f`, then `eval "$(/tmp/promptpet init zsh)"` (with `/tmp` on `PATH` as `promptpet`), `git init -q x && cd x && git commit -q --allow-empty -m hi`, press Enter twice. Expected: a prompt with ` nom ` appears within two prompts.

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: add README with setup, behavior and privacy notes"
```
