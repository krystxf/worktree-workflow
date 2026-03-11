package ui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/krystof/worktree-workflow/internal/config"
	"github.com/krystof/worktree-workflow/internal/git"
	"github.com/krystof/worktree-workflow/internal/sync"
)

type phase int

const (
	phaseCreating phase = iota
	phaseSyncing
	phaseHooks
	phaseOpening
	phaseDone
	phaseFailed
)

type phaseResult struct {
	phase phase
	logs  string
	err   error
}

type hookResult struct {
	index int
	logs  string
	err   error
}

type CreateModel struct {
	branch      string
	root        string
	worktreeDir string
	globalCfg   config.GlobalConfig
	projectCfg  config.ProjectConfig

	current      phase
	failedAt     phase // which phase failed
	spinner      spinner.Model
	logs         []string
	err          error
	quitting     bool
	hookIndex    int // index of the hook currently running (0-based)
	hooksDone    int // number of completed hooks
	hooksTotal   int
	hookStatuses []hookStatus
}

type hookStatus struct {
	command string
	done    bool
	failed  bool
}

var (
	checkStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	failStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	dimStyle     = lipgloss.NewStyle().Faint(true)
)

func NewCreateModel(branch, root, worktreeDir string, globalCfg config.GlobalConfig, projectCfg config.ProjectConfig) CreateModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	statuses := make([]hookStatus, len(projectCfg.PostCopyHooks))
	for i, hook := range projectCfg.PostCopyHooks {
		statuses[i] = hookStatus{command: hook}
	}

	return CreateModel{
		branch:       branch,
		root:         root,
		worktreeDir:  worktreeDir,
		globalCfg:    globalCfg,
		projectCfg:   projectCfg,
		current:      phaseCreating,
		spinner:      s,
		hooksTotal:   len(projectCfg.PostCopyHooks),
		hookStatuses: statuses,
	}
}

func (m CreateModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.runCreate())
}

func (m CreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case phaseResult:
		if msg.logs != "" {
			for _, line := range strings.Split(strings.TrimSpace(msg.logs), "\n") {
				if line != "" {
					m.logs = append(m.logs, line)
				}
			}
		}
		if msg.err != nil {
			m.err = msg.err
			m.failedAt = msg.phase
			m.current = phaseFailed
			return m, tea.Quit
		}

		switch msg.phase {
		case phaseCreating:
			if m.projectCfg.SyncIgnored {
				m.current = phaseSyncing
				return m, m.runSync()
			}
			if m.hooksTotal > 0 {
				m.current = phaseHooks
				return m, m.runHook(0)
			}
			m.current = phaseOpening
			return m, m.runOpen()

		case phaseSyncing:
			if m.hooksTotal > 0 {
				m.current = phaseHooks
				return m, m.runHook(0)
			}
			m.current = phaseOpening
			return m, m.runOpen()

		case phaseOpening:
			m.current = phaseDone
			return m, tea.Quit
		}

	case hookResult:
		if msg.logs != "" {
			for _, line := range strings.Split(strings.TrimSpace(msg.logs), "\n") {
				if line != "" {
					m.logs = append(m.logs, line)
				}
			}
		}
		if msg.err != nil {
			m.hookStatuses[msg.index].failed = true
			m.err = msg.err
			m.failedAt = phaseHooks
			m.current = phaseFailed
			return m, tea.Quit
		}

		m.hookStatuses[msg.index].done = true
		m.hooksDone++

		// More hooks to run?
		if m.hooksDone < m.hooksTotal {
			m.hookIndex = m.hooksDone
			return m, m.runHook(m.hookIndex)
		}

		// All hooks done
		m.current = phaseOpening
		return m, m.runOpen()
	}

	return m, nil
}

func (m CreateModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	fmt.Fprint(&b, "\n")

	// Phase lines
	phases := []struct {
		p    phase
		name string
	}{
		{phaseCreating, "Creating worktree"},
		{phaseSyncing, "Syncing gitignored files"},
		{phaseHooks, "Running post-copy hooks"},
		{phaseOpening, "Opening in editor"},
	}

	for _, ph := range phases {
		if ph.p == phaseSyncing && !m.projectCfg.SyncIgnored {
			continue
		}
		if ph.p == phaseHooks && m.hooksTotal == 0 {
			continue
		}

		if m.current == phaseFailed && ph.p == m.failedAt {
			fmt.Fprintf(&b, "  %s  %s\n", failStyle.Render("✗"), ph.name)
		} else if m.current == phaseFailed && ph.p > m.failedAt {
			fmt.Fprintf(&b, "     %s\n", dimStyle.Render(ph.name))
		} else if ph.p < m.current || (m.current == phaseFailed && ph.p < m.failedAt) {
			fmt.Fprintf(&b, "  %s  %s\n", checkStyle.Render("✓"), ph.name)
		} else if ph.p == m.current && m.current != phaseDone {
			fmt.Fprintf(&b, "  %s %s\n", m.spinner.View(), ph.name)
		} else {
			fmt.Fprintf(&b, "     %s\n", dimStyle.Render(ph.name))
		}

		// Show individual hook steps when in hooks phase
		if ph.p == phaseHooks && m.current >= phaseHooks && m.hooksTotal > 0 {
			for i, hs := range m.hookStatuses {
				label := fmt.Sprintf("[%d/%d] %s", i+1, m.hooksTotal, hs.command)
				if hs.done {
					fmt.Fprintf(&b, "      %s  %s\n", checkStyle.Render("✓"), label)
				} else if hs.failed {
					fmt.Fprintf(&b, "      %s  %s\n", failStyle.Render("✗"), label)
				} else if m.current == phaseHooks && i == m.hookIndex {
					fmt.Fprintf(&b, "      %s %s\n", m.spinner.View(), label)
				} else {
					fmt.Fprintf(&b, "         %s\n", dimStyle.Render(label))
				}
			}
		}
	}

	if m.current == phaseDone {
		fmt.Fprintf(&b, "\n  %s  %s\n", checkStyle.Render("✓"), "Done!")
		fmt.Fprintf(&b, "      %s\n", dimStyle.Render(m.worktreeDir))
	}

	if m.current == phaseFailed {
		fmt.Fprintf(&b, "\n  %s %s\n", failStyle.Render("✗"), failStyle.Render(m.err.Error()))
	}

	// Show recent logs
	if len(m.logs) > 0 {
		fmt.Fprint(&b, "\n")
		start := 0
		if len(m.logs) > 8 {
			start = len(m.logs) - 8
		}
		for _, line := range m.logs[start:] {
			fmt.Fprintf(&b, "    %s\n", dimStyle.Render(line))
		}
	}

	fmt.Fprint(&b, "\n")
	return b.String()
}

func (m CreateModel) runCreate() tea.Cmd {
	return func() tea.Msg {
		out, err := git.WorktreeAdd(m.worktreeDir, m.branch)
		return phaseResult{phase: phaseCreating, logs: out, err: err}
	}
}

func (m CreateModel) runSync() tea.Cmd {
	return func() tea.Msg {
		out, err := sync.SyncIgnored(m.root, m.worktreeDir, m.projectCfg.SyncExcludes)
		return phaseResult{phase: phaseSyncing, logs: out, err: err}
	}
}

func (m CreateModel) runHook(index int) tea.Cmd {
	hook := m.projectCfg.PostCopyHooks[index]
	return func() tea.Msg {
		cmd := exec.Command("sh", "-c", hook)
		cmd.Dir = m.worktreeDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return hookResult{index: index, logs: string(out), err: fmt.Errorf("hook %q failed: %s", hook, strings.TrimSpace(string(out)))}
		}
		return hookResult{index: index, logs: string(out)}
	}
}

func (m CreateModel) runOpen() tea.Cmd {
	return func() tea.Msg {
		if !m.globalCfg.AutoOpenEditor {
			return phaseResult{phase: phaseOpening}
		}
		cmd := exec.Command(m.globalCfg.Editor, m.worktreeDir)
		out, err := cmd.CombinedOutput()
		return phaseResult{phase: phaseOpening, logs: string(out), err: err}
	}
}
