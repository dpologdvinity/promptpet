package repo

import (
	"github.com/dpologdvinity/promptpet/internal/pet"
	"os"
	"path/filepath"
	"strings"
)

func Find(cwd string) (string, string, bool) {
	fi, err := os.Stat(cwd)
	if err != nil || !fi.IsDir() {
		return "", "", false
	}
	d := filepath.Clean(cwd)
	for i := 0; i < 64; i++ {
		path := filepath.Join(d, ".git")
		fi, err := os.Lstat(path)
		if err == nil {
			if fi.IsDir() {
				return d, path, true
			}
			if fi.Mode().IsRegular() {
				data, err := os.ReadFile(path)
				if err == nil {
					line := strings.SplitN(string(data), "\n", 2)[0]
					if strings.HasPrefix(line, "gitdir: ") {
						gd := strings.TrimSpace(strings.TrimPrefix(line, "gitdir: "))
						if !filepath.IsAbs(gd) {
							gd = filepath.Join(d, gd)
						}
						return d, filepath.Clean(gd), true
					}
				}
			}
		}
		parent := filepath.Dir(d)
		if parent == d {
			break
		}
		d = parent
	}
	return "", "", false
}
func Sig(gitDir string) (pet.RepoSig, bool) {
	fi, err := os.Stat(filepath.Join(gitDir, "logs", "HEAD"))
	if err != nil {
		return pet.RepoSig{}, false
	}
	return pet.RepoSig{GitDir: gitDir, Size: fi.Size(), MtimeNs: fi.ModTime().UnixNano()}, true
}
