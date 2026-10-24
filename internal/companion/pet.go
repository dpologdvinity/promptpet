package companion

type PetID string

const (
	PetUsagi  PetID = "usagi"
	PetKochi  PetID = "kochi"
	PetPuff   PetID = "puff"
	PetMochi  PetID = "mochi"
	PetSpook  PetID = "spook"
	PetPippin PetID = "pippin"
	PetGromit PetID = "gromit"
	PetBoba   PetID = "boba"
)

type PetDef struct {
	ID            PetID
	Name          string
	Species       string
	FavoriteSnack string
	PrimaryColor  uint32 // 0xRRGGBB
	AccentColor   uint32
	BlushColor    uint32
	Speech        map[string]string // Mood -> quote
}

func (p PetDef) SpeechFor(mood string) string {
	if p.Speech != nil {
		if quote, ok := p.Speech[mood]; ok && quote != "" {
			return quote
		}
	}
	return "..."
}

var petList = []PetDef{
	{
		ID:            PetUsagi,
		Name:          "Usagi",
		Species:       "Dumpling Bunny",
		FavoriteSnack: "Fresh Strawberry",
		PrimaryColor:  0xFFF8EE,
		AccentColor:   0xF5C2E7,
		BlushColor:    0xFFAEC9,
		Speech: map[string]string{
			"cheer":     "Hop hop! Great job!",
			"sympathy":  "Don't worry, take a soft bunny breath!",
			"energetic": "Zooming fast like a little bunny!",
			"resting":   "Curled up in a soft ball... nap time.",
			"water":     "Time for a sip of fresh water!",
			"stretch":   "Ears up, gentle hop stretch!",
			"lookaway":  "Look at the far meadow for 20 seconds!",
		},
	},
	{
		ID:            PetKochi,
		Name:          "Kochi",
		Species:       "Shiba Pup",
		FavoriteSnack: "Rice Dumpling",
		PrimaryColor:  0xFAB387,
		AccentColor:   0xA6E3A1,
		BlushColor:    0xF4845F,
		Speech: map[string]string{
			"cheer":     "Woof! You did it! Tail wagging!",
			"sympathy":  "Ruff... shake it off, pup's got your back!",
			"energetic": "Zoomies! We are on fire today!",
			"resting":   "Tucked in a warm cinnamon swirl roll.",
			"water":     "Slurp! Grab a bowl of cool water!",
			"stretch":   "Down-dog stretch and wiggle your tail!",
			"lookaway":  "Paws off the screen, scan the yard!",
		},
	},
	{
		ID:            PetPuff,
		Name:          "Puff",
		Species:       "Baby Cloud Dragon",
		FavoriteSnack: "Cinnamon Star",
		PrimaryColor:  0x94E2D5,
		AccentColor:   0xF9E2AF,
		BlushColor:    0xF5C2E7,
		Speech: map[string]string{
			"cheer":     "Tiny flame of victory! *poof*",
			"sympathy":  "Small embers still shine. You've got this!",
			"energetic": "Flapping tiny wings at full speed!",
			"resting":   "Drifting on a cozy cloud pillow... zzz",
			"water":     "Quench your spark with a fresh glass of water!",
			"stretch":   "Spread your wings high into the sky!",
			"lookaway":  "Rest your eyes on distant horizons!",
		},
	},
	{
		ID:            PetMochi,
		Name:          "Mochi",
		Species:       "Strawberry Cat",
		FavoriteSnack: "Taiyaki Waffle",
		PrimaryColor:  0xFFE5D9,
		AccentColor:   0xF38BA8,
		BlushColor:    0xF8AD9D,
		Speech: map[string]string{
			"cheer":     "Purr... that was purrfect!",
			"sympathy":  "Mrow... here's a warm paw for comfort.",
			"energetic": "Swatting bugs and typing like the wind!",
			"resting":   "Basking under a warm patch of sunlight.",
			"water":     "A fresh bowl of water keeps whiskers tidy!",
			"stretch":   "Arch your back, big kitty stretch!",
			"lookaway":  "Gaze out the window, watch the birds.",
		},
	},
	{
		ID:            PetSpook,
		Name:          "Spook",
		Species:       "Cloud Ghostie",
		FavoriteSnack: "Sugar Candy",
		PrimaryColor:  0xCBA6F7,
		AccentColor:   0xB4BEFE,
		BlushColor:    0xF5C2E7,
		Speech: map[string]string{
			"cheer":     "Boo-yah! That was spectacular!",
			"sympathy":  "Floating bugs can't haunt us forever!",
			"energetic": "Wiggling and floating through walls with joy!",
			"resting":   "Phasing into dreamland... floating peacefully.",
			"water":     "Keep your mortal vessel nicely hydrated!",
			"stretch":   "Float up tall, stretch those spooky limbs!",
			"lookaway":  "Stare into the tranquil void for a moment.",
		},
	},
	{
		ID:            PetPippin,
		Name:          "Pippin",
		Species:       "Cozy Penguin",
		FavoriteSnack: "Ice Popsicle",
		PrimaryColor:  0x1E1E2E,
		AccentColor:   0xF38BA8,
		BlushColor:    0xF5C2E7,
		Speech: map[string]string{
			"cheer":     "Waddle dance! Outstanding work!",
			"sympathy":  "Even slick ice has rough patches. Keep waddling!",
			"energetic": "Sliding on my tummy at supersonic speed!",
			"resting":   "Huddled warm in my knitted scarf.",
			"water":     "Chill refreshment time! Drink some crisp water!",
			"stretch":   "Flap your cozy flippers up and down!",
			"lookaway":  "Look across the icy peaks into the distance.",
		},
	},
	{
		ID:            PetGromit,
		Name:          "Gromit",
		Species:       "Rainy Froggy",
		FavoriteSnack: "Dewdrop Melon",
		PrimaryColor:  0xA6E3A1,
		AccentColor:   0xF9E2AF,
		BlushColor:    0xF4A261,
		Speech: map[string]string{
			"cheer":     "Ribbit! Splendid hop forward!",
			"sympathy":  "Croak... every storm brings fresh green sprouts.",
			"energetic": "Ribbit ribbit! Big leaping momentum!",
			"resting":   "Listening peacefully to raindrops on the leaf.",
			"water":     "Moisturize! Drink fresh pond water!",
			"stretch":   "Big froggy leg stretch—reach high!",
			"lookaway":  "Rest your eyes on gentle falling rain.",
		},
	},
	{
		ID:            PetBoba,
		Name:          "Boba",
		Species:       "Tea Panda",
		FavoriteSnack: "Tapioca Pearl",
		PrimaryColor:  0xDDB892,
		AccentColor:   0x313244,
		BlushColor:    0xF5C2E7,
		Speech: map[string]string{
			"cheer":     "Sweet boba victory! Great job!",
			"sympathy":  "Take a gentle pause and sip some sweet tea.",
			"energetic": "Rolling happily like a giant tapioca pearl!",
			"resting":   "Hugging my giant boba cup and dozing off.",
			"water":     "Time to hydrate! Grab some fresh water or tea.",
			"stretch":   "Reach high, unroll those panda shoulders!",
			"lookaway":  "Give your eyes a relaxing tea-break view.",
		},
	},
}

var petMap = func() map[PetID]PetDef {
	m := make(map[PetID]PetDef, len(petList))
	for _, pet := range petList {
		m[pet.ID] = pet
	}
	return m
}()

func GetAllPets() []PetDef {
	res := make([]PetDef, len(petList))
	copy(res, petList)
	return res
}

func GetPet(id PetID) (PetDef, bool) {
	pet, ok := petMap[id]
	return pet, ok
}
