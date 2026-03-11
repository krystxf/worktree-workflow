package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/krystof/worktree-workflow/internal/config"
)

type initPhase int

const (
	initEditor initPhase = iota
	initSuffix
	initSeparator
	initDone
)

type InitResult struct {
	Editor    string
	DirSuffix string
	BranchSep string
	Canceled  bool
}

type InitModel struct {
	editor    textinput.Model
	suffix    textinput.Model
	separator textinput.Model
	phase     initPhase
	result    InitResult
	quitting  bool
	defaults  config.GlobalConfig
}

var (
	labelStyle = lipgloss.NewStyle().Bold(true)
	hintStyle  = lipgloss.NewStyle().Faint(true)
)

func NewInitModel(current config.GlobalConfig) InitModel {
	editor := textinput.New()
	editor.Placeholder = current.Editor
	editor.Focus()
	editor.CharLimit = 100
	editor.Width = 40
	if current.Editor != "" {
		editor.SetValue(current.Editor)
	}

	suffix := textinput.New()
	suffix.Placeholder = current.Naming.WorktreeDirSuffix
	suffix.CharLimit = 50
	suffix.Width = 40
	if current.Naming.WorktreeDirSuffix != "" {
		suffix.SetValue(current.Naming.WorktreeDirSuffix)
	}

	separator := textinput.New()
	separator.Placeholder = current.Naming.BranchSeparator
	separator.CharLimit = 20
	separator.Width = 40
	if current.Naming.BranchSeparator != "" {
		separator.SetValue(current.Naming.BranchSeparator)
	}

	return InitModel{
		editor:    editor,
		suffix:    suffix,
		separator: separator,
		phase:     initEditor,
		defaults:  current,
	}
}

func (m InitModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.result.Canceled = true
			m.quitting = true
			return m, tea.Quit
		case "enter":
			switch m.phase {
			case initEditor:
				value := strings.TrimSpace(m.editor.Value())
				if value == "" {
					value = m.defaults.Editor
				}
				m.result.Editor = value
				m.phase = initSuffix
				m.editor.Blur()
				m.suffix.Focus()
				return m, textinput.Blink

			case initSuffix:
				value := strings.TrimSpace(m.suffix.Value())
				if value == "" {
					value = m.defaults.Naming.WorktreeDirSuffix
				}
				m.result.DirSuffix = value
				m.phase = initSeparator
				m.suffix.Blur()
				m.separator.Focus()
				return m, textinput.Blink

			case initSeparator:
				value := strings.TrimSpace(m.separator.Value())
				if value == "" {
					value = m.defaults.Naming.BranchSeparator
				}
				m.result.BranchSep = value
				m.phase = initDone
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	switch m.phase {
	case initEditor:
		m.editor, cmd = m.editor.Update(msg)
	case initSuffix:
		m.suffix, cmd = m.suffix.Update(msg)
	case initSeparator:
		m.separator, cmd = m.separator.Update(msg)
	}
	return m, cmd
}

func (m InitModel) View() string {
	if m.phase == initDone {
		return ""
	}

	var b strings.Builder
	fmt.Fprint(&b, "\n")

	// Editor (always shown)
	if m.phase == initEditor {
		fmt.Fprintf(&b, "  %s\n", labelStyle.Render("Editor command"))
		fmt.Fprintf(&b, "  %s\n\n", hintStyle.Render("e.g. cursor, code, nvim, zed"))
		fmt.Fprintf(&b, "  %s\n", m.editor.View())
	} else {
		fmt.Fprintf(&b, "  %s %s\n", checkStyle.Render("✓"), fmt.Sprintf("Editor: %s", m.result.Editor))
	}

	// Suffix
	if m.phase == initSuffix {
		fmt.Fprintf(&b, "\n  %s\n", labelStyle.Render("Worktree directory suffix"))
		fmt.Fprintf(&b, "  %s\n\n", hintStyle.Render("appended to repo name for worktree parent dir"))
		fmt.Fprintf(&b, "  repo%s/repo%sbranch\n\n", hintStyle.Render("<suffix>"), hintStyle.Render("<sep>"))
		fmt.Fprintf(&b, "  %s\n", m.suffix.View())
	} else if m.phase > initSuffix {
		fmt.Fprintf(&b, "  %s %s\n", checkStyle.Render("✓"), fmt.Sprintf("Directory suffix: %s", m.result.DirSuffix))
	}

	// Separator
	if m.phase == initSeparator {
		fmt.Fprintf(&b, "\n  %s\n", labelStyle.Render("Branch separator"))
		fmt.Fprintf(&b, "  %s\n\n", hintStyle.Render("separates repo name from branch in worktree dir"))
		fmt.Fprintf(&b, "  repo%s/repo%sbranch\n\n", hintStyle.Render(m.result.DirSuffix), hintStyle.Render("<sep>"))
		fmt.Fprintf(&b, "  %s\n", m.separator.View())
	}

	fmt.Fprintf(&b, "\n  %s\n\n", hintStyle.Render("enter to confirm · esc to cancel"))
	return b.String()
}

func (m InitModel) Result() InitResult {
	return m.result
}
