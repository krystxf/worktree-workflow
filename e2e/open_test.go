package e2e_test

import (
	"os"
	"path/filepath"
	"testing"
)

// A recording editor proves failures never launch an editor, even one that would
// happily accept a missing path as a new file.
func recordingEditor(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "wtw-test-editor")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$WTW_EDITOR_LOG\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(dir, "editor.log")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("WTW_EDITOR_LOG", log)
	writeGlobalConfig(t, `{"editor":"wtw-test-editor --wait","auto_open_editor":true}`)
	return log
}

func TestOpenInvalidWorktree(t *testing.T) {
	for _, kind := range []string{"unregistered", "missing", "file", "ordinary-directory"} {
		t.Run(kind, func(t *testing.T) {
			tmp := t.TempDir()
			repo := filepath.Join(tmp, "repo")
			initRepo(t, repo, "feature")
			log := recordingEditor(t)
			path := filepath.Join(tmp, "linked")
			if kind != "unregistered" {
				gitCmd(t, repo, "worktree", "add", path, "feature")
				if err := os.RemoveAll(path); err != nil {
					t.Fatal(err)
				}
				switch kind {
				case "file":
					if err := os.WriteFile(path, []byte("keep"), 0o644); err != nil {
						t.Fatal(err)
					}
				case "ordinary-directory":
					if err := os.Mkdir(path, 0o755); err != nil {
						t.Fatal(err)
					}
				}
			}
			out, err := wtw(t, repo, "open", "feature")
			if err == nil {
				t.Fatalf("expected error: %s", out)
			}
			assertFileNotExists(t, log)
			if kind == "file" {
				assertFileContents(t, path, "keep")
			}
		})
	}
}

func TestOpenRegisteredPath(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	initRepo(t, repo, "feature")
	log := recordingEditor(t)
	path := filepath.Join(tmp, "custom worktree path")
	gitCmd(t, repo, "worktree", "add", path, "feature")
	mustWtw(t, repo, "open", "feature")
	assertFileContents(t, log, "--wait\n"+realPath(t, path))
	// The main checkout is also an existing worktree, regardless of naming config.
	mustWtw(t, path, "open", currentBranch(t, repo))
	assertFileContents(t, log, "--wait\n"+realPath(t, repo))
}

func TestCreateEditorValidation(t *testing.T) {
	for _, hook := range []string{"", "exit 1", `rm -rf "$PWD"`, `rm -rf .git`} {
		t.Run(hook, func(t *testing.T) {
			tmp := t.TempDir()
			repo := filepath.Join(tmp, "repo")
			initRepo(t, repo, "feature")
			log := recordingEditor(t)
			hooks := []string{}
			if hook != "" {
				hooks = append(hooks, hook)
			}
			writeProjectConfig(t, repo, map[string]any{"sync_ignored": false, "post_copy_hooks": hooks})
			out, err := wtw(t, repo, "create", "feature")
			if hook == "" {
				if err != nil {
					t.Fatalf("create failed: %s", out)
				}
				path := filepath.Join(tmp, "repo--worktrees", "repo--feature")
				assertFileContents(t, log, "--wait\n"+realPath(t, path))
			} else {
				if err == nil {
					t.Fatalf("expected create error: %s", out)
				}
				assertFileNotExists(t, log)
			}
		})
	}
}
