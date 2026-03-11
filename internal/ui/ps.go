package ui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/krystof/worktree-workflow/internal/config"
	gitpkg "github.com/krystof/worktree-workflow/internal/git"
)

type worktreeItem struct {
	branch string
	path   string
}

func (i worktreeItem) Title() string       { return i.branch }
func (i worktreeItem) Description() string { return i.path }
func (i worktreeItem) FilterValue() string { return i.branch }

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
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func NewPickerModel(title string, worktrees []gitpkg.Worktree, action PickerAction) PickerModel {
	items := make([]list.Item, len(worktrees))
	for i, wt := range worktrees {
		items[i] = worktreeItem{branch: wt.Branch, path: wt.Path}
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = title
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	return PickerModel{
		list:          l,
		title:         title,
		action:        action,
		state:         stateBrowsing,
		forcePromptFn: forceRemoveAction,
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

	case tea.KeyMsg:
		// Force confirm prompt
		if m.state == stateConfirmForce {
			switch msg.String() {
			case "y", "Y":
				if m.forceItem != nil && m.forcePromptFn != nil {
					return m, m.forcePromptFn(*m.forceItem)
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
			if item, ok := m.list.SelectedItem().(worktreeItem); ok {
				m.selected = &item
				return m, m.action(item)
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

var (
	warnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	promptStyle = lipgloss.NewStyle().Bold(true)
)

func (m PickerModel) View() string {
	if m.state == stateConfirmForce && m.forceItem != nil {
		var b strings.Builder
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  %s Worktree '%s' contains modified or untracked files.\n\n",
			warnStyle.Render("!"), m.forceItem.branch))
		b.WriteString(fmt.Sprintf("  %s %s\n\n",
			promptStyle.Render("Force remove?"),
			dimStyle.Render("[y/N]")))
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

// OpenEditorAction returns a PickerAction that opens the selected worktree in an editor.
func OpenEditorAction(globalCfg config.GlobalConfig) PickerAction {
	return func(item worktreeItem) tea.Cmd {
		return func() tea.Msg {
			cmd := exec.Command(globalCfg.Editor, item.path)
			out, err := cmd.CombinedOutput()
			if err != nil {
				return actionDoneMsg{err: fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))}
			}
			return actionDoneMsg{result: fmt.Sprintf("Opened %s in %s", item.path, globalCfg.Editor)}
		}
	}
}

// RemoveWorktreeAction returns a PickerAction that removes the selected worktree and prunes.
// If remove fails due to modified/untracked files, it sends needsForceMsg to prompt the user.
func RemoveWorktreeAction(force bool) PickerAction {
	return func(item worktreeItem) tea.Cmd {
		return func() tea.Msg {
			out, err := gitpkg.WorktreeRemove(item.path, force)
			if err != nil {
				if !force && strings.Contains(err.Error(), "modified or untracked") {
					return needsForceMsg{item: item}
				}
				return actionDoneMsg{err: err}
			}

			pruneOut, pruneErr := gitpkg.WorktreePrune()
			_, _ = out, pruneOut

			if pruneErr != nil {
				return actionDoneMsg{result: fmt.Sprintf("Removed worktree '%s' (prune warning: %s)", item.branch, pruneErr)}
			}

			return actionDoneMsg{result: fmt.Sprintf("Removed worktree '%s'", item.branch)}
		}
	}
}

func forceRemoveAction(item worktreeItem) tea.Cmd {
	return func() tea.Msg {
		out, err := gitpkg.WorktreeRemove(item.path, true)
		if err != nil {
			return actionDoneMsg{err: err}
		}

		pruneOut, pruneErr := gitpkg.WorktreePrune()
		_, _ = out, pruneOut

		if pruneErr != nil {
			return actionDoneMsg{result: fmt.Sprintf("Force removed worktree '%s' (prune warning: %s)", item.branch, pruneErr)}
		}

		return actionDoneMsg{result: fmt.Sprintf("Force removed worktree '%s'", item.branch)}
	}
}
