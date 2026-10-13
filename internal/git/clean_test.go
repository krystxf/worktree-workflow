package git

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRemoteRefsTransport(t *testing.T) {
	for _, mode := range []string{"default", "command", "config", "ssh"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			if out, err := exec.Command("git", "init", dir).CombinedOutput(); err != nil {
				t.Fatalf("init: %s", out)
			}
			for _, key := range []string{"GIT_SSH_COMMAND", "GIT_SSH", "GIT_SSH_VARIANT"} {
				t.Setenv(key, "")
				if err := os.Unsetenv(key); err != nil {
					t.Fatal(err)
				}
			}
			wrapper := filepath.Join(dir, "ssh")
			script := "#!/bin/sh\nprintf '%s|%s|%s' \"$GIT_TERMINAL_PROMPT\" \"$GIT_SSH_COMMAND\" \"$*\" >&2\nexit 1\n"
			if err := os.WriteFile(wrapper, []byte(script), 0o755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			expected := "ssh -oBatchMode=yes"
			switch mode {
			case "command":
				expected = "ssh -oCustomOption=yes"
				t.Setenv("GIT_SSH_COMMAND", expected)
			case "config":
				expected = "CustomOption=yes"
				if out, err := exec.Command("git", "-C", dir, "config", "core.sshCommand", "ssh -oCustomOption=yes").CombinedOutput(); err != nil {
					t.Fatalf("config: %s", out)
				}
			case "ssh":
				expected = "0||"
				t.Setenv("GIT_SSH", wrapper)
			}
			_, err := remoteRefs(context.Background(), dir, "ssh://example.invalid/repo", "refs/heads/main")
			if err == nil || !strings.Contains(err.Error(), "0|") || !strings.Contains(err.Error(), expected) {
				t.Fatalf("unexpected transport: %v", err)
			}
		})
	}
}

func TestRemoteRefsTimeout(t *testing.T) {
	dir := t.TempDir()
	wrapper := filepath.Join(dir, "slow-ssh")
	if err := os.WriteFile(wrapper, []byte("#!/bin/sh\nexec sleep 10\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_SSH_COMMAND", wrapper)
	t.Setenv("GIT_SSH_VARIANT", "ssh")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := remoteRefs(ctx, dir, "ssh://example.invalid/repo", "refs/heads/main")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline: %v", err)
	}
	if time.Since(start) > 3*time.Second {
		t.Fatal("remote lookup did not return promptly")
	}
}
