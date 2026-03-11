package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type initPhase int

const (
	initEditor initPhase = iota
	initDone
)

type InitResult struct {
	Editor   string
	Canceled bool
}

type InitModel struct {
	editor   textinput.Model
	phase    initPhase
	result   InitResult
	quitting bool
}

var (
	labelStyle = lipgloss.NewStyle().Bold(true)
	hintStyle  = lipgloss.NewStyle().Faint(true)
)

func NewInitModel(currentEditor string) InitModel {
	ti := textinput.New()
	ti.Placeholder = "cursor"
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40

	if currentEditor != "" {
		ti.SetValue(currentEditor)
	}

	return InitModel{
		editor: ti,
		phase:  initEditor,
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
			value := strings.TrimSpace(m.editor.Value())
			if value == "" {
				value = "cursor"
			}
			m.result.Editor = value
			m.phase = initDone
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)
	return m, cmd
}

func (m InitModel) View() string {
	if m.phase == initDone {
		return ""
	}

	var b strings.Builder
	fmt.Fprint(&b, "\n")
	fmt.Fprintf(&b, "  %s\n", labelStyle.Render("Editor command"))
	fmt.Fprintf(&b, "  %s\n\n", hintStyle.Render("e.g. cursor, code, nvim, zed"))
	fmt.Fprintf(&b, "  %s\n\n", m.editor.View())
	fmt.Fprintf(&b, "  %s\n", hintStyle.Render("enter to confirm · esc to cancel"))
	fmt.Fprint(&b, "\n")
	return b.String()
}

func (m InitModel) Result() InitResult {
	return m.result
}
