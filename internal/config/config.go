package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type NamingConfig struct {
	WorktreeDirSuffix string `json:"worktree_dir_suffix"`
	BranchSeparator   string `json:"branch_separator"`
}

type GlobalConfig struct {
	Editor         string       `json:"editor"`
	AutoOpenEditor bool         `json:"auto_open_editor"`
	Naming         NamingConfig `json:"naming"`
}

type ProjectConfig struct {
	SyncIgnored   bool     `json:"sync_ignored"`
	SyncExcludes  []string `json:"sync_excludes"`
	PostCopyHooks []string `json:"post_copy_hooks"`
}

func DefaultGlobal() GlobalConfig {
	return GlobalConfig{
		Editor:         "cursor",
		AutoOpenEditor: true,
		Naming: NamingConfig{
			WorktreeDirSuffix: "--worktrees",
			BranchSeparator:   "--",
		},
	}
}

func DefaultProject() ProjectConfig {
	return ProjectConfig{
		SyncIgnored:   true,
		SyncExcludes:  []string{},
		PostCopyHooks: []string{},
	}
}

func LoadGlobal() (GlobalConfig, error) {
	cfg := DefaultGlobal()

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg, nil
	}

	path := filepath.Join(home, ".config", "worktree-workflow", "config.json")
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

	// Re-apply defaults for zero values
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
	cfg := DefaultProject()

	path := filepath.Join(gitRoot, ".worktree-workflow.json")
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

	if cfg.SyncExcludes == nil {
		cfg.SyncExcludes = []string{}
	}
	if cfg.PostCopyHooks == nil {
		cfg.PostCopyHooks = []string{}
	}

	return cfg, nil
}
