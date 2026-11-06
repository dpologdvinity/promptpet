// Package repo locates git repositories and reads their reflogs directly from
// disk (never via a git subprocess) so the prompt hook stays within budget.
package repo

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dpologdvinity/promptpet/internal/pet"
)

// Find walks up from cwd to the nearest ".git" (directory, or a "gitdir:"
// file as used by linked worktrees and submodules) and returns the worktree
// root and the git dir.
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

// Sig fingerprints the HEAD reflog (size + mtime). The prompt stores it and
// only inspects the repo again when it changes, i.e. after a git operation.
func Sig(gitDir string) (pet.RepoSig, bool) {
	fi, err := os.Stat(filepath.Join(gitDir, "logs", "HEAD"))
	if err != nil {
		return pet.RepoSig{}, false
	}
	return pet.RepoSig{GitDir: gitDir, Size: fi.Size(), MtimeNs: fi.ModTime().UnixNano()}, true
}

// Info summarises the repository state the prompt hook reacts to.
type Info struct {
	Branch string // current branch, or the short SHA when HEAD is detached
	// LastAction classifies the newest HEAD reflog entry as "commit",
	// "branch" or "merge" ("" for anything else); see classify.
	LastAction string
	// CommitCount is the number of new commits recorded in the HEAD reflog
	// on now's local calendar day (amends excluded; see isNewCommit).
	CommitCount int
}

// GetInfo reads HEAD and the HEAD reflog directly and reports the current
// branch, the newest reflog action and the number of commits made today
// (the calendar day of now, in now's location).
func GetInfo(gitDir string, now time.Time) Info {
	var info Info
	headData, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err == nil {
		headStr := strings.TrimSpace(string(headData))
		if strings.HasPrefix(headStr, "ref: refs/heads/") {
			info.Branch = strings.TrimPrefix(headStr, "ref: refs/heads/")
		} else if len(headStr) >= 7 {
			info.Branch = headStr[:7]
		}
	}

	f, err := os.Open(filepath.Join(gitDir, "logs", "HEAD"))
	if err != nil {
		return info
	}
	defer f.Close()
	y, m, d := now.Date()
	var lastTs int64
	var lastMsg string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ts, msg, ok := parseReflogLine(scanner.Text())
		if !ok {
			continue
		}
		lastTs, lastMsg = ts, msg
		if isNewCommit(msg) {
			if cy, cm, cd := time.Unix(ts, 0).In(now.Location()).Date(); cy == y && cm == m && cd == d {
				info.CommitCount++
			}
		}
	}
	info.LastAction = classify(gitDir, lastTs, lastMsg)
	return info
}

// parseReflogLine splits "<old> <new> <name> <email> <unix-ts> <tz>\t<msg>"
// into its timestamp and message. Fields are taken from the end because the
// committer name may contain spaces.
func parseReflogLine(line string) (int64, string, bool) {
	head, msg, ok := strings.Cut(line, "\t")
	fields := strings.Fields(head)
	if !ok || len(fields) < 2 {
		return 0, "", false
	}
	ts, err := strconv.ParseInt(fields[len(fields)-2], 10, 64)
	return ts, msg, err == nil
}

// reflogAction splits a reflog message into its git-generated prefix (e.g.
// "commit (amend)", "merge feat", "checkout") and the rest. Ref names cannot
// contain ':', so the first ": " always ends the prefix.
func reflogAction(msg string) (action, rest string) {
	action, rest, _ = strings.Cut(msg, ": ")
	return action, rest
}

// isNewCommit reports whether a reflog message records a newly created
// commit: "commit", "commit (initial)" or "commit (merge)". Amends rewrite an
// existing commit and are not counted.
func isNewCommit(msg string) bool {
	switch action, _ := reflogAction(msg); action {
	case "commit", "commit (initial)", "commit (merge)":
		return true
	}
	return false
}

// classify maps the newest HEAD reflog entry to a git milestone:
//   - "merge": "commit (merge)" (a conflicted merge concluded), "merge <x>"
//     (including fast-forward merges), or "pull…" ending in "Merge made by…"
//   - "commit": "commit" or "commit (initial)"
//   - "branch": "checkout: moving from X to Y" where Y's own reflog starts
//     with "branch: Created from" within 2s of the checkout (checkout -b,
//     switch -c, or the first checkout of a remote branch)
//
// Everything else (amend, reset, rebase, plain checkout, ff pull) is "".
func classify(gitDir string, ts int64, msg string) string {
	action, rest := reflogAction(msg)
	switch {
	case action == "commit (merge)", strings.HasPrefix(action, "merge "),
		strings.HasPrefix(action, "pull") && strings.HasPrefix(rest, "Merge made by"):
		return "merge"
	case action == "commit", action == "commit (initial)":
		return "commit"
	case action == "checkout":
		if i := strings.LastIndex(rest, " to "); i >= 0 && justCreated(gitDir, rest[i+len(" to "):], ts) {
			return "branch"
		}
	}
	return ""
}

// justCreated reports whether local branch name's reflog begins with a
// "branch: Created from" entry within 2 seconds of ts.
func justCreated(gitDir, name string, ts int64) bool {
	f, err := os.Open(filepath.Join(commonDir(gitDir), "logs", "refs", "heads", filepath.FromSlash(name)))
	if err != nil {
		return false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	if !sc.Scan() {
		return false
	}
	created, msg, ok := parseReflogLine(sc.Text())
	return ok && strings.HasPrefix(msg, "branch: Created from") && created >= ts-2 && created <= ts+2
}

// commonDir returns the directory holding shared refs and branch reflogs:
// for a linked worktree's git dir it follows the "commondir" file, otherwise
// it is gitDir itself.
func commonDir(gitDir string) string {
	data, err := os.ReadFile(filepath.Join(gitDir, "commondir"))
	if err != nil {
		return gitDir
	}
	p := strings.TrimSpace(string(data))
	if !filepath.IsAbs(p) {
		p = filepath.Join(gitDir, p)
	}
	return filepath.Clean(p)
}
