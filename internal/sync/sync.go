package sync

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/krystof/worktree-workflow/internal/git"
)

func SyncIgnored(root, dest string, excludes []string) (string, error) {
	files, err := git.IgnoredFiles(root, excludes)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "No gitignored files to sync.", nil
	}

	cmd := exec.Command("rsync", "-av", "--hard-links", "--from0", "--files-from=-", root+"/", dest+"/")
	cmd.Stdin = strings.NewReader(string(files))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("rsync failed: %s", strings.TrimSpace(string(out)))
	}

	return string(out), nil
}
