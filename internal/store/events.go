package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type MealLine struct {
	Unix int64
	SHA  string
}

func AppendMeals(dir string, meals []MealLine) error {
	if len(meals) == 0 {
		return nil
	}
	f, err := os.OpenFile(filepath.Join(dir, "events.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	var b strings.Builder
	for _, m := range meals {
		fmt.Fprintf(&b, "meal %d %s\n", m.Unix, m.SHA)
	}
	_, err = io.WriteString(f, b.String())
	return err
}
func ReadMeals(dir string, offset int64) (int, int64, error) {
	f, err := os.Open(filepath.Join(dir, "events.log"))
	if errors.Is(err, os.ErrNotExist) {
		return 0, offset, nil
	}
	if err != nil {
		return 0, offset, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return 0, offset, err
	}
	if st.Size() < offset {
		return 0, st.Size(), nil
	}
	if _, err = f.Seek(offset, io.SeekStart); err != nil {
		return 0, offset, err
	}
	data, err := io.ReadAll(io.LimitReader(f, 64*1024))
	if err != nil {
		return 0, offset, err
	}
	end := bytes.LastIndexByte(data, '\n')
	if end < 0 {
		return 0, offset, nil
	}
	n := 0
	for _, line := range bytes.Split(data[:end+1], []byte{'\n'}) {
		if bytes.HasPrefix(line, []byte("meal ")) {
			n++
		}
	}
	return n, offset + int64(end+1), nil
}

type Feed struct {
	Version int              `json:"version"`
	Repos   map[string]int64 `json:"repos"`
}

func LoadFeed(dir string) (*Feed, error) {
	data, err := os.ReadFile(filepath.Join(dir, "feed.json"))
	if errors.Is(err, os.ErrNotExist) {
		return &Feed{Version: 1, Repos: map[string]int64{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var f Feed
	if json.Unmarshal(data, &f) != nil || f.Version != 1 || f.Repos == nil {
		return &Feed{Version: 1, Repos: map[string]int64{}}, nil
	}
	return &f, nil
}
func SaveFeed(dir string, f *Feed) error {
	if len(f.Repos) > 200 {
		keys := make([]string, 0, len(f.Repos))
		for k := range f.Repos {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			a, b := f.Repos[keys[i]], f.Repos[keys[j]]
			if a == b {
				return keys[i] < keys[j]
			}
			return a < b
		})
		for _, k := range keys[:len(keys)-200] {
			delete(f.Repos, k)
		}
	}
	data, err := json.Marshal(f)
	if err != nil {
		return err
	}
	return WriteAtomic(filepath.Join(dir, "feed.json"), data)
}
