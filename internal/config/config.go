package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type NamingConfig struct {
	WorktreeDirSuffix string `json:"worktree_dir_suffix"`
	BranchSeparator   string `json:"branch_separator"`
}

type GlobalConfig struct {
	Editor         string       `json:"editor"`
	AutoOpenEditor *bool        `json:"auto_open_editor"`
	Naming         NamingConfig `json:"naming"`
}

type ProjectConfig struct {
	SyncIgnored   *bool    `json:"sync_ignored"`
	SyncExcludes  []string `json:"sync_excludes"`
	PostCopyHooks []string `json:"post_copy_hooks"`
}

func boolPtr(b bool) *bool { return &b }

func DefaultGlobal() GlobalConfig {
	return GlobalConfig{
		Editor:         "cursor",
		AutoOpenEditor: boolPtr(true),
		Naming: NamingConfig{
			WorktreeDirSuffix: "--worktrees",
			BranchSeparator:   "--",
		},
	}
}

func DefaultProject() ProjectConfig {
	return ProjectConfig{
		SyncIgnored:   boolPtr(true),
		SyncExcludes:  []string{},
		PostCopyHooks: []string{},
	}
}

// EditorArgs returns the binary and arguments for opening a path in the configured editor.
// Supports multi-word editor commands like "code --wait".
func (c GlobalConfig) EditorArgs(path string) (string, []string) {
	parts := strings.Fields(c.Editor)
	if len(parts) == 0 {
		return "cursor", []string{path}
	}
	return parts[0], append(parts[1:], path)
}

func LoadGlobal() (GlobalConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return DefaultGlobal(), nil
	}
	return loadGlobalFrom(filepath.Join(home, ".config", "worktree-workflow", "config.json"))
}

func loadGlobalFrom(path string) (GlobalConfig, error) {
	cfg := DefaultGlobal()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	if cfg.AutoOpenEditor == nil {
		cfg.AutoOpenEditor = boolPtr(true)
	}
	if cfg.Naming.WorktreeDirSuffix == "" {
		cfg.Naming.WorktreeDirSuffix = "--worktrees"
	}
	if cfg.Naming.BranchSeparator == "" {
		cfg.Naming.BranchSeparator = "--"
	}
	if cfg.Editor == "" {
		cfg.Editor = "cursor"
	}

	return cfg, nil
}

func LoadProject(gitRoot string) (ProjectConfig, error) {
	return loadProjectFrom(filepath.Join(gitRoot, ".worktree-workflow.json"))
}

func loadProjectFrom(path string) (ProjectConfig, error) {
	cfg := DefaultProject()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	if cfg.SyncIgnored == nil {
		cfg.SyncIgnored = boolPtr(true)
	}
	if cfg.SyncExcludes == nil {
		cfg.SyncExcludes = []string{}
	}
	if cfg.PostCopyHooks == nil {
		cfg.PostCopyHooks = []string{}
	}

	return cfg, nil
}
