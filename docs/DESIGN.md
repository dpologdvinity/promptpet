# Design Document: Virtual Device Mini-Window Companion

**Document:** `docs/DESIGN.md`  
**Date:** 2026-10-24  
**Status:** Approved by User · Ready for Implementation Planning  

---

## 1. Intent & Purpose

`promptpet` originated as a shell prompt segment. While fast, plain ASCII text art in the prompt lacks visual richness, animation, and true interactive play.

This design transforms `promptpet` into a **vibrant, interactive virtual device companion** living in a dedicated mini-window pinned to the corner of the user's screen. Styled like a retro-futuristic virtual pet console (inspired by Tamagotchi and classic handheld gadgets), it features smooth, cute vector/illustrated creatures that hang out quietly during work and offer instant 15-second keyboard mini-games whenever the user needs a mental break.

The mini-window companion **coexists with the shell CLI**:
- The shell CLI retains its fast prompt hooks (<5ms) and command reporting.
- The shell sends lightweight background signals (command executions, git commits, exit codes) to the companion window.
- The companion window brings the pet alive on screen without cluttering the terminal.

---

## 2. Goals, Non-Goals & Anti-Goals

### Goals
- **Charming Visual Companion:** Smooth illustrated artwork with expressive animations (not coarse pixel grids and not ASCII text).
- **8 Distinct Cute Pets:** Diverse roster of adorable creatures with unique idle animations, favorite snacks, and moods. **All 8 pets unlocked and playable from day 1.**
- **8 Fast Keyboard Mini-Games:** Bite-sized (<20 seconds) games controlled purely via keyboard (Arrow keys, Spacebar, 1/2/3).
- **Unobtrusive Screen Corner Presence:** Floating borderless or compact mini-window pinned to the corner of the workspace (WSLg / Linux / Windows / macOS).
- **Dual Coexistence Architecture:** Shell prompt remains fast and minimal; floating mini-window handles visual rendering and mini-games.
- **Rich Work Rhythm Reactions:** Companion reacts dynamically to work pace, exit codes, git milestones, time of day, and stretch reminders.
- **Zero-Stress Philosophy:** Pets never die, get sick, starve, or guilt-trip the user.

### Non-Goals
- Not a full-screen video game or complex RPG with inventories and leveling grinds.
- Not an intrusive desktop pet that walks across windows, steals the mouse, or disrupts active typing.
- No network/online multiplayer or microtransactions.

### Anti-Goals
- **No CPU Hogging:** Must remain lightweight (<2% CPU while idling, zero GPU strain).
- **No Demanding Care Chores:** No notifications demanding feeding while the user is deep in focus.
- **No Terminal Latency:** Inter-process event signaling must never block or delay the shell prompt.

---

## 3. Form Factor & User Interface

### The Virtual Console Device
The application runs as a compact (approx. 240×340px) always-on-top or pinnable floating mini-window launched via `promptpet gui` (or running as a desktop daemon).
- **Outer Shell:** Egg-shaped or rounded pebble console with customizable pastel casing colors (Lavender, Coral, Mint, Midnight).
- **LCD Display Screen:** Center glass screen displaying the active pet, status emotes, work speech bubbles, and active mini-games.
- **Control Layout:**
  - 3 Face Buttons (A: Left/Action, B: Select/Jump, C: Right/Cancel).
  - Full Keyboard Bindings: `Left`/`Right` arrows, `Spacebar`, `1`/`2`/`3`, `Tab` (switch pet), `Esc` (exit game/menu).

---

## 4. The 8 Pet Companions Roster (All Unlocked from Start)

Every pet has smooth illustrated art, custom idle behaviors, and reactions:

| Pet Name | Species | Visual Style | Idle Behavior | Favorite Snack |
|---|---|---|---|---|
| **1. Usagi** | Dumpling Bunny | Cream white, floppy ears, pastel pink blush | Nibbles strawberry, naps in a ball | Fresh Strawberry |
| **2. Kochi** | Shiba Pup | Golden honey fur, white cheeks, green bandana | Wags curled tail, rolls over | Rice Dumpling |
| **3. Puff** | Baby Cloud Dragon | Mint teal, tiny gold horns, soft wings | Puffs harmless tiny heart smoke | Cinnamon Star |
| **4. Mochi** | Strawberry Cat | Calico pastel patches, round paws, glassy eyes | Paws at floating sparkles, sips tea | Taiyaki Waffle |
| **5. Spook** | Cloud Ghostie | Lavender translucent glow, floating wiggles | Drifts up and down, does gentle boos | Sugar Candy |
| **6. Pippin** | Cozy Penguin | Navy & cream, tiny knitted scarf | Waddles, slides on tummy | Ice Popsicle |
| **7. Gromit** | Rainy Froggy | Soft sage green, yellow belly, leaf umbrella | Sits under leaf, blinks peacefully | Dewdrop Melon |
| **8. Boba** | Tea Panda | Charcoal & milk-tea beige, round chubby ears | Hugs boba straw, rolls like a ball | Tapioca Pearl |

Pets can be freely switched anytime from the device menu or by pressing `Tab` inside the window.

---

## 5. The 8 Keyboard Mini-Games

Each mini-game takes 10–25 seconds, ends with a clear score/win celebration, and rewards the pet with happiness and treats:

### 1. Snack Catch
- **Objective:** Falling treats descend from the top of the mini-screen.
- **Controls:** `Left` / `Right` arrows (or `A` / `D`).
- **Mechanics:** Scurry left and right to catch 5 snacks before missing 3.

### 2. Puddle Leap
- **Objective:** Side-scrolling runner where pet leaps over obstacles.
- **Controls:** `Spacebar` (or `Up` arrow).
- **Mechanics:** Tap to jump over puddles, rocks, and wandering beetles. Reach the flag to win!

### 3. Look Away (Classic Tamagotchi Game)
- **Objective:** Guess which direction your pet will turn its head.
- **Controls:** `Left` / `Right` arrows.
- **Mechanics:** 5 quick rounds. Guess correctly 3+ times to win.

### 4. Rhythm Dance
- **Objective:** Match falling musical arrows to the rhythm bar.
- **Controls:** `Left`, `Up`, `Down`, `Right` arrows.
- **Mechanics:** Hit 8 beats on time as music sparkles shower down.

### 5. Memory Simon
- **Objective:** Repeat a sequence of flashing treats.
- **Controls:** `1`, `2`, `3` (or `Left`, `Up`, `Right`).
- **Mechanics:** Pet shows a 3-to-5 step sequence; input the pattern from memory.

### 6. Whack-a-Bug
- **Objective:** Keep pests away from the pet's flower garden.
- **Controls:** `1`, `2`, `3` (corresponding to left, middle, right holes).
- **Mechanics:** Tap bugs as they pop up; avoid tapping friendly ladybugs.

### 7. Balloon Pop
- **Objective:** Aim and pop floating colorful balloons.
- **Controls:** `Up` / `Down` arrows to adjust aim angle, `Spacebar` to shoot dart.
- **Mechanics:** Pop 5 balloons with 6 darts.

### 8. Word Snack
- **Objective:** Quick typing reflex game.
- **Controls:** Standard alphanumeric keys.
- **Mechanics:** A cute 3-4 letter word floats across (e.g. `TEA`, `CAT`, `JOY`). Type it before it floats away.

---

## 6. Work Rhythm & Reactive Engine

The companion reacts dynamically to the user's computer and coding rhythm via 5 triggers:

1. **Command Flow & Burst Tracking:**
   - **Flow State (Active Bursts):** High command frequency triggers energized pet animations, sparkle particles, and playful tail wags.
   - **Thinking / Idle State:** Long pauses between commands cause the pet to doze off, read a tiny book, or daydream quietly.
2. **Success & Error Sympathy:**
   - **Exit Code 0 (Success/Build Win):** Pet jumps with mini confetti or high-fives the screen.
   - **Exit Code != 0 (Error/Failure):** Pet offers a comforting hot mug of tea or holds a tiny band-aid ("You got this!").
3. **Day & Night Atmosphere:**
   - **Morning (6 AM - 12 PM):** Pet drinks a steaming cup of coffee/tea.
   - **Afternoon (12 PM - 6 PM):** Bright daytime vibes, nibbling favorite snack.
   - **Late Night (10 PM - 5 AM):** Pet wears a tiny nightcap, curls up in a blanket, snoozes quietly with "zzz" floats.
4. **Gentle Wellness Reminders (Hydration, Stretch, Eye Rest):**
   - **Drink Water:** Every ~45-60 mins, pet holds up a glass of water with cute water drops: *"hydration check!"*
   - **Stretch:** Pet stands on tiptoes, stretches paws to the sky, and rolls over to prompt a 30-second body stretch.
   - **Look Away (20-20-20 Rule):** Pet peers through tiny binoculars or gazes out a window: *"rest eyes, look into the distance!"*
   - All reminders are gentle visual prompts with soft animations — zero noisy alarms, zero nagging, zero guilt.
5. **Git Milestone Fanfare:**
   - Creating a new branch, merging, or hitting the 5th/10th commit of the day triggers a special victory celebration dance.

---

## 7. Technical Architecture & Stack

### Core Technology: Go + Ebitengine (`github.com/hajimehoshi/ebiten/v2`)
- **Native Pure Go Desktop Binary:** Zero heavy webview runtimes or browser engines.
- **Cross-Platform:** Runs on Linux (X11 / Wayland / WSLg), Windows native, and macOS.
- **Window Management:** Borderless or floating tool window with custom rounded frame artwork.
- **Keyboard & Audio:** Direct keyboard polling for games, optional retro blip SFX (toggleable).

### Inter-Process Communication (Shell CLI ↔ Mini-Window)
- **Non-blocking Event Stream:**
  - The existing shell CLI hooks (`promptpet prompt`, `precmd`) append a tiny single-line JSON event to `$XDG_STATE_HOME/promptpet/events.jsonl` (or write to a local domain socket if available).
  - Shell execution overhead is strictly bounded (<1ms file append / non-blocking socket write).
  - The GUI companion window reads newly arrived events on its 15-30 FPS tick.
- **Event Schema:**
  ```json
  {
    "timestamp": 1729759200,
    "event": "command_finish",
    "exit_code": 0,
    "duration_ms": 320,
    "git": {
      "branch": "master",
      "action": "commit",
      "commit_count": 5
    }
  }
  ```

### Performance & Low Resource Consumption
- **Frame Rate Throttling:**
  - **In-Game Mode:** 60 FPS for crisp keyboard responsiveness.
  - **Idle Desktop Companion Mode:** Throttled to 15–20 FPS with sleep pauses, yielding <1.5% CPU load.
- **State Persistence:**
  - Saved to `$XDG_STATE_HOME/promptpet/state.json`: active pet selection, high scores for each mini-game, friendship heart counters, casing color choice.

---

## 8. CLI Commands & Workflow

- `promptpet init <bash|zsh|fish>`: Emits hook functions that update shell prompt and notify the companion window.
- `promptpet gui`: Launches the floating virtual companion mini-window.
- `promptpet prompt`: Fast prompt segment renderer (keeps shell responsive).
- `promptpet status`: Shows pet stats, friendship level, and mini-game high scores in terminal.
- `promptpet pet <name>`: Switch active pet via CLI command if window is closed.
