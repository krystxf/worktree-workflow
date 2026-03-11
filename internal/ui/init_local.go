package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/krystof/worktree-workflow/internal/config"
)

type localInitPhase int

const (
	localSyncIgnored localInitPhase = iota
	localSyncExcludes
	localPostCopyHooks
	localDone
)

type LocalInitResult struct {
	SyncIgnored   bool
	SyncExcludes  []string
	PostCopyHooks []string
	Canceled      bool
}

type LocalInitModel struct {
	syncIgnored bool
	excludes    textinput.Model
	hooks       textinput.Model
	phase       localInitPhase
	result      LocalInitResult
	quitting    bool
	defaults    config.ProjectConfig
}

func NewLocalInitModel(current config.ProjectConfig) LocalInitModel {
	excludes := textinput.New()
	excludes.Placeholder = "node_modules, .venv, dist"
	excludes.CharLimit = 200
	excludes.Width = 60

	if len(current.SyncExcludes) > 0 {
		excludes.SetValue(strings.Join(current.SyncExcludes, ", "))
	}

	hooks := textinput.New()
	hooks.Placeholder = "npm install, make build"
	hooks.CharLimit = 500
	hooks.Width = 60

	if len(current.PostCopyHooks) > 0 {
		hooks.SetValue(strings.Join(current.PostCopyHooks, ", "))
	}

	return LocalInitModel{
		syncIgnored: *current.SyncIgnored,
		excludes:    excludes,
		hooks:       hooks,
		phase:       localSyncIgnored,
		defaults:    current,
	}
}

func (m LocalInitModel) Init() tea.Cmd {
	return nil
}

func (m LocalInitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.result.Canceled = true
			m.quitting = true
			return m, tea.Quit
		case "enter":
			switch m.phase {
			case localSyncIgnored:
				m.result.SyncIgnored = m.syncIgnored
				if m.syncIgnored {
					m.phase = localSyncExcludes
					m.excludes.Focus()
					return m, textinput.Blink
				}
				// Skip excludes if sync is off
				m.phase = localPostCopyHooks
				m.hooks.Focus()
				return m, textinput.Blink

			case localSyncExcludes:
				m.result.SyncExcludes = parseCommaSeparated(m.excludes.Value())
				m.phase = localPostCopyHooks
				m.excludes.Blur()
				m.hooks.Focus()
				return m, textinput.Blink

			case localPostCopyHooks:
				m.result.PostCopyHooks = parseCommaSeparated(m.hooks.Value())
				m.phase = localDone
				m.quitting = true
				return m, tea.Quit
			}

		case "y", "Y":
			if m.phase == localSyncIgnored {
				m.syncIgnored = true
				return m, nil
			}
		case "n", "N":
			if m.phase == localSyncIgnored {
				m.syncIgnored = false
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	switch m.phase {
	case localSyncExcludes:
		m.excludes, cmd = m.excludes.Update(msg)
	case localPostCopyHooks:
		m.hooks, cmd = m.hooks.Update(msg)
	}
	return m, cmd
}

func (m LocalInitModel) View() string {
	if m.phase == localDone {
		return ""
	}

	var b strings.Builder
	fmt.Fprint(&b, "\n")

	// Sync ignored
	if m.phase == localSyncIgnored {
		fmt.Fprintf(&b, "  %s\n", labelStyle.Render("Sync gitignored files to worktrees?"))
		fmt.Fprintf(&b, "  %s\n\n", hintStyle.Render("copies .env, build artifacts etc. via rsync"))
		if m.syncIgnored {
			fmt.Fprintf(&b, "  %s / n\n", labelStyle.Render("Y"))
		} else {
			fmt.Fprintf(&b, "  y / %s\n", labelStyle.Render("N"))
		}
	} else {
		label := "no"
		if m.result.SyncIgnored {
			label = "yes"
		}
		fmt.Fprintf(&b, "  %s %s\n", checkStyle.Render("✓"), fmt.Sprintf("Sync gitignored files: %s", label))
	}

	// Excludes
	if m.phase == localSyncExcludes {
		fmt.Fprintf(&b, "\n  %s\n", labelStyle.Render("Exclude from sync"))
		fmt.Fprintf(&b, "  %s\n\n", hintStyle.Render("comma-separated paths to skip, e.g. node_modules, .venv"))
		fmt.Fprintf(&b, "  %s\n", m.excludes.View())
	} else if m.phase > localSyncExcludes && m.result.SyncIgnored {
		excludes := "none"
		if len(m.result.SyncExcludes) > 0 {
			excludes = strings.Join(m.result.SyncExcludes, ", ")
		}
		fmt.Fprintf(&b, "  %s %s\n", checkStyle.Render("✓"), fmt.Sprintf("Sync excludes: %s", excludes))
	}

	// Hooks
	if m.phase == localPostCopyHooks {
		fmt.Fprintf(&b, "\n  %s\n", labelStyle.Render("Post-copy hooks"))
		fmt.Fprintf(&b, "  %s\n\n", hintStyle.Render("comma-separated commands to run after worktree creation"))
		fmt.Fprintf(&b, "  %s\n", m.hooks.View())
	}

	fmt.Fprintf(&b, "\n  %s\n\n", hintStyle.Render("enter to confirm · esc to cancel"))
	return b.String()
}

func (m LocalInitModel) Result() LocalInitResult {
	return m.result
}

func parseCommaSeparated(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
