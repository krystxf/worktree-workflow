package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// CleanReason returns an explanation when a worktree must be preserved.
// Read the remote directly: cached tracking refs can be stale or deleted remotely.
func CleanReason(w Worktree) (string, error) {
	if w.Prunable {
		return "prunable worktree (directory missing)", nil
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	remote, err := remoteRefs(ctx, w.Path, parts[0], parts[1])
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

// remoteRefs bounds remote access while preserving user-selected SSH transports.
func remoteRefs(ctx context.Context, path, remote, ref string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", path, "ls-remote", "--refs", "--", remote, ref)
	cmd.WaitDelay = time.Second
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	// Apply OpenSSH options only to the default transport. Custom commands and
	// variants may use different flags; the deadline still bounds their execution.
	if os.Getenv("GIT_SSH_COMMAND") == "" && os.Getenv("GIT_SSH") == "" && os.Getenv("GIT_SSH_VARIANT") == "" {
		configured := false
		for _, key := range []string{"core.sshCommand", "ssh.variant"} {
			probe := exec.CommandContext(ctx, "git", "-C", path, "config", "--get", key)
			probe.WaitDelay = time.Second
			if err := probe.Run(); err == nil {
				configured = true
			} else if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 {
				return "", fmt.Errorf("read %s: %w", key, err)
			}
		}
		if !configured {
			cmd.Env = append(cmd.Env, "GIT_SSH_COMMAND=ssh -oBatchMode=yes")
		}
	}
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return "", fmt.Errorf("git ls-remote: %w", ctx.Err())
	}
	if err != nil {
		return "", fmt.Errorf("git ls-remote: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}
