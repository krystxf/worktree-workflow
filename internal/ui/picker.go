package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	gitpkg "github.com/krystof/worktree-workflow/internal/git"
)

const mainWorktreeLabel = " (main)"

type worktreeItem struct {
	branch   string
	path     string
	disabled bool
	isMain   bool
}

func (i worktreeItem) Title() string {
	if i.disabled {
		return dimStyle.Render(i.branch) + dimStyle.Render(mainWorktreeLabel)
	}
	if i.isMain {
		return i.branch + dimStyle.Render(mainWorktreeLabel)
	}
	return i.branch
}

func (i worktreeItem) Description() string {
	if i.disabled {
		return dimStyle.Render("main worktree — not removable")
	}
	return i.path
}

func (i worktreeItem) FilterValue() string { return i.branch }

// worktreeDelegate renders list items; for main worktree, "(main)" is always dim (no hover color).
type worktreeDelegate struct {
	list.DefaultDelegate
}

func (d worktreeDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	wi, ok := item.(worktreeItem)
	if !ok || !wi.isMain {
		d.DefaultDelegate.Render(w, m, index, item)
		return
	}
	s := &d.Styles
	width := m.Width()
	if width <= 0 {
		return
	}
	textwidth := width - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight()
	fullTitle := wi.branch + mainWorktreeLabel
	fullTitle = ansi.Truncate(fullTitle, textwidth, "...")
	var branchPart, mainSuffix string
	if strings.HasSuffix(fullTitle, mainWorktreeLabel) {
		branchPart = strings.TrimSuffix(fullTitle, mainWorktreeLabel)
		mainSuffix = mainWorktreeLabel
	} else {
		branchPart = fullTitle
		mainSuffix = ""
	}
	desc := wi.Description()
	if d.ShowDescription {
		desc = ansi.Truncate(desc, textwidth, "...")
	}
	isSelected := index == m.Index()
	emptyFilter := m.FilterState() == list.Filtering && m.FilterValue() == ""
	if emptyFilter {
		_, _ = fmt.Fprintf(w, "%s%s\n%s", s.DimmedTitle.Render(branchPart), dimStyle.Render(mainSuffix), s.DimmedDesc.Render(desc))
	} else if isSelected && m.FilterState() != list.Filtering {
		_, _ = fmt.Fprintf(w, "%s%s\n%s", s.SelectedTitle.Render(branchPart), dimStyle.Render(mainSuffix), s.SelectedDesc.Render(desc))
	} else {
		_, _ = fmt.Fprintf(w, "%s%s\n%s", s.NormalTitle.Render(branchPart), dimStyle.Render(mainSuffix), s.NormalDesc.Render(desc))
	}
}

// actionDoneMsg is sent when the post-selection action completes.
type actionDoneMsg struct {
	result string
	err    error
}

// needsForceMsg is sent when remove fails due to modified/untracked files.
type needsForceMsg struct {
	item worktreeItem
}

// PickerAction defines what happens when a worktree is selected.
type PickerAction func(item worktreeItem) tea.Cmd

type pickerState int

const (
	stateBrowsing pickerState = iota
	stateConfirmForce
	stateLoading
)

type PickerModel struct {
	list          list.Model
	title         string
	action        PickerAction
	quitting      bool
	selected      *worktreeItem
	result        string
	err           error
	state         pickerState
	forceItem     *worktreeItem
	forcePromptFn func(item worktreeItem) tea.Cmd // action to run on force confirm
	spinner       spinner.Model
	loadingLabel  string
}

var (
	docStyle    = lipgloss.NewStyle().Margin(1, 2)
	warnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	promptStyle = lipgloss.NewStyle().Bold(true)
)

func NewPickerModel(title string, worktrees []gitpkg.Worktree, action PickerAction, mainPath, disabledPath, loadingLabel string) PickerModel {
	items := make([]list.Item, len(worktrees))
	for i, wt := range worktrees {
		items[i] = worktreeItem{
			branch:   wt.Branch,
			path:     wt.Path,
			disabled: wt.Path == disabledPath,
			isMain:   mainPath != "" && wt.Path == mainPath,
		}
	}

	delegate := worktreeDelegate{DefaultDelegate: list.NewDefaultDelegate()}
	l := list.New(items, delegate, 0, 0)
	l.Title = title
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return PickerModel{
		list:          l,
		title:         title,
		action:        action,
		state:         stateBrowsing,
		forcePromptFn: forceRemoveAction,
		spinner:       s,
		loadingLabel:  loadingLabel,
	}
}

func (m PickerModel) Init() tea.Cmd {
	return nil
}

func (m PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case spinner.TickMsg:
		if m.state == stateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		// Loading — only allow ctrl+c
		if m.state == stateLoading {
			if msg.String() == "ctrl+c" {
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil
		}

		// Force confirm prompt
		if m.state == stateConfirmForce {
			switch msg.String() {
			case "y", "Y":
				if m.forceItem != nil && m.forcePromptFn != nil {
					m.selected = m.forceItem
					m.state = stateLoading
					return m, tea.Batch(m.spinner.Tick, m.forcePromptFn(*m.forceItem))
				}
			case "n", "N", "esc", "q":
				m.state = stateBrowsing
				m.forceItem = nil
				m.selected = nil
				return m, nil
			}
			return m, nil
		}

		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(worktreeItem); ok && !item.disabled {
				m.selected = &item
				m.state = stateLoading
				return m, tea.Batch(m.spinner.Tick, m.action(item))
			}
		}

	case needsForceMsg:
		m.state = stateConfirmForce
		m.forceItem = &msg.item
		return m, nil

	case actionDoneMsg:
		m.result = msg.result
		m.err = msg.err
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m PickerModel) View() string {
	if m.state == stateLoading && m.selected != nil {
		return fmt.Sprintf("\n  %s %s '%s'...\n", m.spinner.View(), m.loadingLabel, m.selected.branch)
	}
	if m.state == stateConfirmForce && m.forceItem != nil {
		var b strings.Builder
		fmt.Fprint(&b, "\n")
		fmt.Fprintf(&b, "  %s Worktree '%s' contains modified or untracked files.\n\n",
			warnStyle.Render("!"), m.forceItem.branch)
		fmt.Fprintf(&b, "  %s %s\n\n",
			promptStyle.Render("Force remove?"),
			dimStyle.Render("[y/N]"))
		return b.String()
	}
	return docStyle.Render(m.list.View())
}

// ResultMessage returns the result to print after the program exits (outside alt screen).
func (m PickerModel) ResultMessage() string {
	if m.selected == nil {
		return ""
	}
	if m.err != nil {
		return fmt.Sprintf("\n  %s %s\n", failStyle.Render("✗"), m.err)
	}
	if m.result != "" {
		return fmt.Sprintf("\n  %s %s\n", checkStyle.Render("✓"), m.result)
	}
	return ""
}

// SelectedPath returns the path of the selected worktree, or empty string if none selected.
func (m PickerModel) SelectedPath() string {
	if m.selected == nil {
		return ""
	}
	return m.selected.path
}
