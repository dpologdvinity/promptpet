package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMealsAppendAndRead(t *testing.T) {
	dir := t.TempDir()
	if n, off, err := ReadMeals(dir, 0); n != 0 || off != 0 || err != nil {
		t.Fatal(n, off, err)
	}
	AppendMeals(dir, []MealLine{{100, "abcdef1"}, {200, "abcdef2"}})
	n, off, _ := ReadMeals(dir, 0)
	if n != 2 || off != int64(len("meal 100 abcdef1\nmeal 200 abcdef2\n")) {
		t.Fatal(n, off)
	}
	if n, off2, _ := ReadMeals(dir, off); n != 0 || off2 != off {
		t.Fatal(n, off2)
	}
}
func TestReadMealsPartialLine(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "events.log"), []byte("meal 1 a\nnoise line\nmeal 2 b"), 0o600)
	n, off, _ := ReadMeals(dir, 0)
	if n != 1 || off != int64(len("meal 1 a\nnoise line\n")) {
		t.Fatal(n, off)
	}
}
func TestReadMealsTruncatedFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "events.log"), []byte("meal 1 a\n"), 0o600)
	if n, off, _ := ReadMeals(dir, 500); n != 0 || off != 9 {
		t.Fatal(n, off)
	}
}
func TestFeedPrune(t *testing.T) {
	dir := t.TempDir()
	f, _ := LoadFeed(dir)
	for i := 0; i < 205; i++ {
		f.Repos[fmt.Sprintf("/r/%03d", i)] = int64(i % 203)
	} // i=203 gets 0 and i=204 gets 1, so the 5 smallest are 000, 203, 001, 204, 002
	SaveFeed(dir, f)
	g, _ := LoadFeed(dir)
	if len(g.Repos) != 200 {
		t.Fatal(len(g.Repos))
	}
	for _, gone := range []string{"/r/000", "/r/203", "/r/001", "/r/204", "/r/002"} {
		if _, ok := g.Repos[gone]; ok {
			t.Errorf("%s kept", gone)
		}
	}
}
