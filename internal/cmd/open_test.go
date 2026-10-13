package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krystof/worktree-workflow/internal/config"
)

func TestOpenEditorRejectsInvalidPaths(t *testing.T) {
	root := t.TempDir()
	if out, err := exec.Command("git", "init", root).CombinedOutput(); err != nil {
		t.Fatalf("init: %s", out)
	}
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{filepath.Join(root, "missing"), file, nested, t.TempDir()} {
		t.Run(filepath.Base(path), func(t *testing.T) {
			// An invalid executable would yield "failed to open editor" if validation
			// were bypassed. All callers share this guard, including TUI selections.
			err := openEditor(config.GlobalConfig{Editor: "wtw-nonexistent-editor"}, path)
			if err == nil || !strings.Contains(err.Error(), "cannot open worktree") {
				t.Fatalf("expected path validation error, got %v", err)
			}
		})
	}
}
