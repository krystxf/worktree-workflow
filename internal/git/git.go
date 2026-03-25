package git

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrWorktreeModified is returned when a worktree cannot be removed due to modified or untracked files.
var ErrWorktreeModified = errors.New("worktree contains modified or untracked files")

type Worktree struct {
	Path   string
	Branch string
}

func Root() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

func MainRoot() (string, error) {
	out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository")
	}
	worktrees := parseWorktreeList(out)
	if len(worktrees) == 0 {
		return "", fmt.Errorf("not a git repository")
	}
	return worktrees[0].Path, nil
}

func RepoName(root string) string {
	return filepath.Base(root)
}

func WorktreeDir(root, suffix, separator, branch string) string {
	repo := RepoName(root)
	parent := filepath.Dir(root)
	return filepath.Join(parent, repo+suffix, repo+separator+SanitizeBranch(branch))
}

// SanitizeBranch replaces characters that are problematic in filesystem paths
// so that branch names like "feature/login" become "feature-login" in directory names.
func SanitizeBranch(branch string) string {
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		" ", "-",
		"..", "-",
		"~", "-",
		"^", "-",
		"*", "-",
		"?", "-",
		"[", "-",
		"]", "-",
		"@{", "-",
	)
	s := replacer.Replace(branch)
	// Trim leading/trailing dots and dashes
	s = strings.Trim(s, ".-")
	return s
}

func BranchExists(branch string) bool {
	err := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+branch).Run()
	return err == nil
}

func BranchCreate(branch string) (string, error) {
	cmd := exec.Command("git", "branch", branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git branch create failed: %s", strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func WorktreeAdd(path, branch string) (string, error) {
	cmd := exec.Command("git", "worktree", "add", path, branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git worktree add failed: %s", strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func WorktreeRemove(path string, force bool) (string, error) {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := strings.TrimSpace(string(out))
		if strings.Contains(errMsg, "modified or untracked") {
			return string(out), ErrWorktreeModified
		}
		return string(out), fmt.Errorf("git worktree remove failed: %s", errMsg)
	}
	return string(out), nil
}

func WorktreePrune() (string, error) {
	cmd := exec.Command("git", "worktree", "prune")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git worktree prune failed: %s", strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func WorktreeList() ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return parseWorktreeList(out), nil
}

func parseWorktreeList(data []byte) []Worktree {
	var worktrees []Worktree
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var current Worktree

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "worktree "):
			current.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "detached":
			current.Branch = "(detached HEAD)"
		case line == "":
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			current = Worktree{}
		}
	}
	// Handle last entry if no trailing newline
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees
}

func IgnoredFiles(root string, excludes []string) ([]byte, error) {
	cmd := exec.Command("git", "-C", root, "ls-files", "--others", "--ignored", "--exclude-standard", "-z")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files failed: %w", err)
	}
	return filterExcluded(out, excludes), nil
}

func filterExcluded(data []byte, excludes []string) []byte {
	if len(excludes) == 0 {
		return data
	}

	var filtered [][]byte
	for _, entry := range bytes.Split(data, []byte{0}) {
		if len(entry) == 0 {
			continue
		}
		excluded := false
		for _, exc := range excludes {
			if strings.Contains(string(entry), exc) {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, entry)
		}
	}

	result := bytes.Join(filtered, []byte{0})
	if len(result) > 0 {
		result = append(result, 0)
	}
	return result
}
