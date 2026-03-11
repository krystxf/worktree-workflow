package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultGlobal(t *testing.T) {
	cfg := DefaultGlobal()
	if cfg.Editor != "cursor" {
		t.Errorf("expected editor 'cursor', got %q", cfg.Editor)
	}
	if cfg.AutoOpenEditor == nil || !*cfg.AutoOpenEditor {
		t.Error("expected auto_open_editor true")
	}
	if cfg.Naming.WorktreeDirSuffix != "--worktrees" {
		t.Errorf("expected suffix '--worktrees', got %q", cfg.Naming.WorktreeDirSuffix)
	}
	if cfg.Naming.BranchSeparator != "--" {
		t.Errorf("expected separator '--', got %q", cfg.Naming.BranchSeparator)
	}
}

func TestDefaultProject(t *testing.T) {
	cfg := DefaultProject()
	if cfg.SyncIgnored == nil || !*cfg.SyncIgnored {
		t.Error("expected sync_ignored true")
	}
	if len(cfg.SyncExcludes) != 0 {
		t.Errorf("expected empty sync_excludes, got %v", cfg.SyncExcludes)
	}
	if len(cfg.PostCopyHooks) != 0 {
		t.Errorf("expected empty post_copy_hooks, got %v", cfg.PostCopyHooks)
	}
}

func TestLoadGlobalMissingFile(t *testing.T) {
	cfg, err := loadGlobalFrom("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Editor != "cursor" {
		t.Errorf("expected default editor, got %q", cfg.Editor)
	}
	if !*cfg.AutoOpenEditor {
		t.Error("expected auto_open_editor true")
	}
}

func TestLoadGlobalPartialConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	// Config with only editor — auto_open_editor should default to true
	if err := os.WriteFile(path, []byte(`{"editor": "nvim"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadGlobalFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Editor != "nvim" {
		t.Errorf("expected editor 'nvim', got %q", cfg.Editor)
	}
	if !*cfg.AutoOpenEditor {
		t.Error("expected auto_open_editor to remain true when not specified in JSON")
	}
}

func TestLoadGlobalExplicitFalse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"editor": "nvim", "auto_open_editor": false}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadGlobalFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *cfg.AutoOpenEditor {
		t.Error("expected auto_open_editor false when explicitly set")
	}
}

func TestLoadProjectPartialConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".worktree-workflow.json")
	// Config with only hooks — sync_ignored should default to true
	if err := os.WriteFile(path, []byte(`{"post_copy_hooks": ["npm install"]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadProjectFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !*cfg.SyncIgnored {
		t.Error("expected sync_ignored to remain true when not specified in JSON")
	}
	if len(cfg.PostCopyHooks) != 1 || cfg.PostCopyHooks[0] != "npm install" {
		t.Errorf("unexpected hooks: %v", cfg.PostCopyHooks)
	}
}

func TestLoadProjectExplicitFalse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".worktree-workflow.json")
	if err := os.WriteFile(path, []byte(`{"sync_ignored": false}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadProjectFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *cfg.SyncIgnored {
		t.Error("expected sync_ignored false when explicitly set")
	}
}

func TestEditorArgsSingleWord(t *testing.T) {
	cfg := GlobalConfig{Editor: "cursor"}
	bin, args := cfg.EditorArgs("/path/to/dir")
	if bin != "cursor" {
		t.Errorf("bin = %q, want 'cursor'", bin)
	}
	if len(args) != 1 || args[0] != "/path/to/dir" {
		t.Errorf("args = %v, want [/path/to/dir]", args)
	}
}

func TestEditorArgsMultiWord(t *testing.T) {
	cfg := GlobalConfig{Editor: "code --wait"}
	bin, args := cfg.EditorArgs("/path/to/dir")
	if bin != "code" {
		t.Errorf("bin = %q, want 'code'", bin)
	}
	if len(args) != 2 || args[0] != "--wait" || args[1] != "/path/to/dir" {
		t.Errorf("args = %v, want [--wait /path/to/dir]", args)
	}
}

func TestEditorArgsMultipleFlags(t *testing.T) {
	cfg := GlobalConfig{Editor: "nvim -u NONE"}
	bin, args := cfg.EditorArgs("/path")
	if bin != "nvim" {
		t.Errorf("bin = %q, want 'nvim'", bin)
	}
	if len(args) != 3 || args[0] != "-u" || args[1] != "NONE" || args[2] != "/path" {
		t.Errorf("args = %v, want [-u NONE /path]", args)
	}
}

func TestEditorArgsEmpty(t *testing.T) {
	cfg := GlobalConfig{Editor: ""}
	bin, args := cfg.EditorArgs("/path")
	if bin != "cursor" {
		t.Errorf("bin = %q, want fallback 'cursor'", bin)
	}
	if len(args) != 1 || args[0] != "/path" {
		t.Errorf("args = %v, want [/path]", args)
	}
}
