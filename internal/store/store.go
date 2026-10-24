package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dpologdvinity/promptpet/internal/pet"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

var ErrBusy = errors.New("another promptpet is busy; try again")
var ErrNewer = errors.New("state was written by a newer promptpet; please upgrade")

func Dir(getenv func(string) string) (string, error) {
	if d := getenv("PROMPTPET_HOME"); d != "" {
		return d, nil
	}
	if d := getenv("XDG_STATE_HOME"); filepath.IsAbs(d) {
		return filepath.Join(d, "promptpet"), nil
	}
	if d := getenv("HOME"); d != "" {
		return filepath.Join(d, ".local", "state", "promptpet"), nil
	}
	return "", errors.New("cannot find a home directory")
}
func Lock(dir, name string, budget time.Duration) (func(), error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(budget)
	for {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return func() { syscall.Flock(int(f.Fd()), syscall.LOCK_UN); f.Close() }, nil
		}
		if err != syscall.EWOULDBLOCK && err != syscall.EAGAIN {
			f.Close()
			return nil, err
		}
		if !time.Now().Before(deadline) {
			f.Close()
			return nil, ErrBusy
		}
		time.Sleep(500 * time.Microsecond)
	}
}
func LoadState(dir string, now time.Time) (*pet.State, string, error) {
	path := filepath.Join(dir, "state.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return pet.New(now), "", nil
	}
	if err != nil {
		return nil, "", err
	}
	var s pet.State
	if json.Unmarshal(data, &s) != nil {
		moved := fmt.Sprintf("%s.corrupt-%d", path, now.Unix())
		if err := os.Rename(path, moved); err != nil {
			return nil, "", err
		}
		return pet.New(now), moved, nil
	}
	if s.Version > 1 {
		return nil, "", ErrNewer
	}
	return &s, "", nil
}
func SaveState(dir string, s *pet.State) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return WriteAtomic(filepath.Join(dir, "state.json"), data)
}
func WriteAtomic(path string, data []byte) error {
	tmp := fmt.Sprintf("%s.tmp-%d", path, os.Getpid())
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}
