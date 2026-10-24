package companion_test

import (
	"testing"

	"github.com/dpologdvinity/promptpet/internal/companion"
)

func TestAllEightPetsExist(t *testing.T) {
	pets := companion.GetAllPets()
	if len(pets) != 8 {
		t.Fatalf("expected exactly 8 pets, got %d", len(pets))
	}
	expected := []companion.PetID{
		companion.PetUsagi, companion.PetKochi, companion.PetPuff, companion.PetMochi,
		companion.PetSpook, companion.PetPippin, companion.PetGromit, companion.PetBoba,
	}
	for _, id := range expected {
		p, ok := companion.GetPet(id)
		if !ok {
			t.Errorf("missing pet %s", id)
			continue
		}
		if p.ID != id {
			t.Errorf("expected pet ID %s, got %s", id, p.ID)
		}
		if p.Name == "" {
			t.Errorf("pet %s has empty Name", id)
		}
		if p.Species == "" {
			t.Errorf("pet %s has empty Species", id)
		}
		if p.FavoriteSnack == "" {
			t.Errorf("pet %s has empty FavoriteSnack", id)
		}
		if p.PrimaryColor == 0 {
			t.Errorf("pet %s has zero PrimaryColor", id)
		}
		if p.BlushColor == 0 {
			t.Errorf("pet %s has zero BlushColor", id)
		}
		if len(p.Speech) == 0 {
			t.Errorf("pet %s has empty Speech map", id)
		}

		requiredMoods := []string{"cheer", "sympathy", "energetic", "resting", "water", "stretch", "lookaway"}
		for _, mood := range requiredMoods {
			quote := p.SpeechFor(mood)
			if quote == "" {
				t.Errorf("pet %s missing speech for mood %q", id, mood)
			}
		}
	}
}

func TestGetPetNotFound(t *testing.T) {
	p, ok := companion.GetPet("nonexistent")
	if ok {
		t.Fatalf("expected ok=false for unknown pet, got pet: %+v", p)
	}
}
