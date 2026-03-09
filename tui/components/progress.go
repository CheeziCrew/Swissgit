package components

import (
	"fmt"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Base16 ANSI colors — respects terminal theme
var (
	cRed   = lipgloss.Color("1")
	cGreen = lipgloss.Color("2")
	cMag   = lipgloss.Color("5")
	cFg    = lipgloss.Color("7")
	cGray  = lipgloss.Color("8")
	cBrMag = lipgloss.Color("13")
)

var (
	successStyle = lipgloss.NewStyle().Foreground(cGreen)
	failStyle    = lipgloss.NewStyle().Foreground(cRed)
	pendingStyle = lipgloss.NewStyle().Foreground(cGray)
	errStyle     = lipgloss.NewStyle().Foreground(cRed)
	resultHint   = lipgloss.NewStyle().Foreground(cGray)
	nameStyle    = lipgloss.NewStyle().Foreground(cFg)
	countStyle   = lipgloss.NewStyle().Bold(true).Foreground(cBrMag)
)

// RepoTaskStatus tracks the state of a single repo operation.
type RepoTaskStatus int

const (
	TaskPending RepoTaskStatus = iota
	TaskRunning
	TaskDone
	TaskFailed
)

// RepoTask represents one repo being processed.
type RepoTask struct {
	Name   string
	Path   string
	Status RepoTaskStatus
	Result string
	Error  string
}

// RepoTaskUpdateMsg is sent when a repo task completes.
type RepoTaskUpdateMsg struct {
	Index  int
	Status RepoTaskStatus
	Result string
	Error  string
}

// AllTasksDoneMsg is sent when all tasks have finished.
type AllTasksDoneMsg struct{}

// ProgressModel shows a progress bar with a spinner for the current task.
type ProgressModel struct {
	Tasks    []RepoTask
	bar      progress.Model
	spinner  spinner.Model
	done     bool
	finished int
}

func NewProgressModel(tasks []RepoTask) ProgressModel {
	bar := progress.New(
		progress.WithoutPercentage(),
		progress.WithWidth(40),
		progress.WithColors(lipgloss.Color("13"), lipgloss.Color("8")),
	)

	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(cBrMag)

	return ProgressModel{
		Tasks:   tasks,
		bar:     bar,
		spinner: s,
	}
}

func (m ProgressModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m ProgressModel) Update(msg tea.Msg) (ProgressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w := msg.Width - 20
		if w > 60 {
			w = 60
		}
		if w < 20 {
			w = 20
		}
		m.bar.SetWidth(w)

	case spinner.TickMsg:
		if m.done {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		var cmd tea.Cmd
		m.bar, cmd = m.bar.Update(msg)
		return m, cmd

	case RepoTaskUpdateMsg:
		if msg.Index < len(m.Tasks) {
			m.Tasks[msg.Index].Status = msg.Status
			m.Tasks[msg.Index].Result = msg.Result
			m.Tasks[msg.Index].Error = msg.Error
		}

		// Recount finished
		m.finished = 0
		allDone := true
		for _, t := range m.Tasks {
			switch t.Status {
			case TaskDone, TaskFailed:
				m.finished++
			default:
				allDone = false
			}
		}

		if allDone {
			m.done = true
			return m, tea.Batch(
				m.bar.SetPercent(1.0),
				func() tea.Msg { return AllTasksDoneMsg{} },
			)
		}

		pct := float64(m.finished) / float64(len(m.Tasks))
		return m, m.bar.SetPercent(pct)
	}
	return m, nil
}

func (m ProgressModel) View() string {
	total := len(m.Tasks)
	pct := float64(m.finished) / float64(total)

	var s string

	// Progress bar
	s += "  " + m.bar.ViewAs(pct) + "\n"

	// Counter
	s += "  " + countStyle.Render(fmt.Sprintf("%d/%d", m.finished, total))

	failed := 0
	for _, t := range m.Tasks {
		if t.Status == TaskFailed {
			failed++
		}
	}
	if failed > 0 {
		s += "  " + errStyle.Render(fmt.Sprintf("(%d failed)", failed))
	}
	s += "\n\n"

	// Show currently running tasks (max 3, compact)
	if !m.done {
		running := 0
		for _, t := range m.Tasks {
			if t.Status == TaskRunning {
				if running < 3 {
					s += fmt.Sprintf("  %s %s\n", m.spinner.View(), nameStyle.Render(t.Name))
				}
				running++
			}
		}
		if running > 3 {
			s += resultHint.Render(fmt.Sprintf("  … and %d more", running-3)) + "\n"
		}
	}

	return s
}

func (m ProgressModel) IsDone() bool {
	return m.done
}
