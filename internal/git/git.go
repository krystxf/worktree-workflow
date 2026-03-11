package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

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

func RepoName(root string) string {
	return filepath.Base(root)
}

func WorktreeDir(root, suffix, separator, branch string) string {
	repo := RepoName(root)
	parent := filepath.Dir(root)
	return filepath.Join(parent, repo+suffix, repo+separator+branch)
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
		return string(out), fmt.Errorf("git worktree remove failed: %s", strings.TrimSpace(string(out)))
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
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %w", err)
	}

	var worktrees []Worktree
	scanner := bufio.NewScanner(bytes.NewReader(out))
	var current Worktree

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "worktree "):
			current.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
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

	return worktrees, nil
}

func IgnoredFiles(root string, excludes []string) ([]byte, error) {
	cmd := exec.Command("git", "-C", root, "ls-files", "--others", "--ignored", "--exclude-standard", "-z")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files failed: %w", err)
	}

	if len(excludes) == 0 {
		return out, nil
	}

	// Filter out excluded paths (null-separated)
	var filtered [][]byte
	for _, entry := range bytes.Split(out, []byte{0}) {
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

	// Rejoin with null separators
	result := bytes.Join(filtered, []byte{0})
	if len(result) > 0 {
		result = append(result, 0)
	}
	return result, nil
}
