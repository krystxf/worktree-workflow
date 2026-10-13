package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// CleanReason returns an explanation when a worktree must be preserved.
// Read the remote directly: cached tracking refs can be stale or deleted remotely.
func CleanReason(w Worktree) (string, error) {
	if w.Locked {
		return "locked", nil
	}
	if w.Branch == "" || w.Branch == "(detached HEAD)" {
		return "detached HEAD or bare repository", nil
	}
	run := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", w.Path}, args...)...).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(string(out)))
		}
		return strings.TrimSpace(string(out)), nil
	}
	status, err := run("status", "--porcelain", "--untracked-files=all", "--ignored", "--ignore-submodules=none")
	if err != nil {
		return "", err
	}
	if status != "" {
		return "modified, untracked, or ignored files", nil
	}
	upstream, err := run("for-each-ref", "--format=%(upstream:remotename)%09%(upstream:remoteref)", "refs/heads/"+w.Branch)
	if err != nil {
		return "", err
	}
	parts := strings.Split(upstream, "\t")
	if len(parts) != 2 || parts[0] == "" || parts[0] == "." || !strings.HasPrefix(parts[1], "refs/heads/") {
		return "no remote upstream", nil
	}
	head, err := run("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	remote, err := run("ls-remote", "--refs", "--", parts[0], parts[1])
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(remote, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == parts[1] && fields[0] == head {
			return "", nil
		}
	}
	return "remote upstream is missing or differs from HEAD", nil
}
