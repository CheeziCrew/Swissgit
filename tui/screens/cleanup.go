package screens

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/CheeziCrew/curd"

	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

type cleanupStep int

const (
	cleanupStepDrop cleanupStep = iota
	cleanupStepRepoSelect
	cleanupStepProgress
	cleanupStepResults
)

type cleanupTaskDoneMsg struct {
	index  int
	result ops.CleanupResult
}

// CleanupModel handles the cleanup flow.
type CleanupModel struct {
	step cleanupStep

	dropChanges bool
	confirm     curd.ConfirmModel

	repoSelect RepoSelectModel
	progress   components.ProgressModel
	results    components.ResultModel
	repos      []string
	viewport   viewport.Model
	viewReady  bool
	height     int
}

func NewCleanupModel() CleanupModel {
	return CleanupModel{
		step: cleanupStepDrop,
		confirm: curd.NewConfirmModel(curd.ConfirmConfig{
			Question: "Drop uncommitted changes?",
			Caller:   "cleanup",
			Palette:  palette,
		}),
	}
}

func (m CleanupModel) Init() tea.Cmd {
	return nil
}

func (m CleanupModel) Update(msg tea.Msg) (CleanupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		if !m.viewReady {
			m.viewport = viewport.New(viewport.WithWidth(msg.Width-6), viewport.WithHeight(msg.Height-10))
			m.viewReady = true
		} else {
			m.viewport.SetWidth(msg.Width - 6)
			m.viewport.SetHeight(msg.Height - 10)
		}
	default:
		_ = msg
	}

	switch m.step {
	case cleanupStepDrop:
		return m.updateDrop(msg)
	case cleanupStepRepoSelect:
		return m.updateRepoSelect(msg)
	case cleanupStepProgress:
		return m.updateProgress(msg)
	case cleanupStepResults:
		return m.updateResults(msg)
	}
	return m, nil
}

func (m CleanupModel) updateDrop(msg tea.Msg) (CleanupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case curd.ConfirmMsg:
		m.dropChanges = msg.Confirmed
		m.step = cleanupStepRepoSelect
		dropLabel := "no"
		if m.dropChanges {
			dropLabel = "yes"
		}
		preview := titleStyle.Render("🧹 Cleanup") + "\n\n" + summaryBlock(summaryLine("drop changes", dropLabel))
		m.repoSelect = NewRepoSelectModel("cleanup", ".", lipgloss.Height(preview), m.height)
		return m, m.repoSelect.Init()
	case curd.BackToMenuMsg:
		return m, func() tea.Msg { return BackToMenuMsg{} }
	}

	var cmd tea.Cmd
	m.confirm, cmd = m.confirm.Update(msg)
	return m, cmd
}

func (m CleanupModel) updateRepoSelect(msg tea.Msg) (CleanupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case RepoSelectDoneMsg:
		m.repos = msg.Paths
		return m, m.startCleanupTasks()
	case BackToMenuMsg:
		m.step = cleanupStepDrop
		return m, nil
	}
	var cmd tea.Cmd
	m.repoSelect, cmd = m.repoSelect.Update(msg)
	return m, cmd
}

func (m *CleanupModel) startCleanupTasks() tea.Cmd {
	var tasks []components.RepoTask
	for _, p := range m.repos {
		name := p
		if n, err := ops.GetRepoNameForPath(p); err == nil {
			name = n
		}
		tasks = append(tasks, components.RepoTask{
			Name:   name,
			Path:   p,
			Status: components.TaskRunning,
		})
	}

	m.progress = components.NewProgressModel(tasks)
	m.step = cleanupStepProgress

	var cmds []tea.Cmd
	cmds = append(cmds, m.progress.Init())

	for i, p := range m.repos {
		idx := i
		path := p
		drop := m.dropChanges

		cmds = append(cmds, func() tea.Msg {
			result := ops.CleanupRepo(path, drop, "main")
			return cleanupTaskDoneMsg{index: idx, result: result}
		})
	}

	return tea.Batch(cmds...)
}

func (m CleanupModel) updateProgress(msg tea.Msg) (CleanupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case cleanupTaskDoneMsg:
		status := components.TaskDone
		resultStr := ""
		errStr := ""
		if !msg.result.Success {
			status = components.TaskFailed
			errStr = msg.result.Error
		} else {
			var parts []string
			if msg.result.PrunedBranches > 0 {
				parts = append(parts, fmt.Sprintf("pruned %d branches", msg.result.PrunedBranches))
			}
			if msg.result.DroppedChanges {
				parts = append(parts, "dropped changes")
			}
			resultStr = strings.Join(parts, ", ")
		}

		updateMsg := components.RepoTaskUpdateMsg{
			Index: msg.index, Status: status, Result: resultStr, Error: errStr,
		}
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(updateMsg)
		return m, cmd

	case components.AllTasksDoneMsg:
		m.results = components.NewResultModel("Cleanup", m.progress.Tasks)
		m.step = cleanupStepResults
		m.viewport.SetContent(m.results.View())
		return m, nil
	}

	var cmd tea.Cmd
	m.progress, cmd = m.progress.Update(msg)
	return m, cmd
}

func (m CleanupModel) updateResults(msg tea.Msg) (CleanupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q", "enter"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m CleanupModel) View() string {
	var s string
	s += titleStyle.Render("🧹 Cleanup") + "\n\n"

	switch m.step {
	case cleanupStepDrop:
		s += m.confirm.View()
		return s

	case cleanupStepRepoSelect:
		dropLabel := "no"
		if m.dropChanges {
			dropLabel = lipgloss.NewStyle().Foreground(colorRed).Render("yes ⚠")
		}
		s += summaryBlock(summaryLine("drop changes", dropLabel))
		s += m.repoSelect.View()

	case cleanupStepProgress:
		s += m.progress.View()

	case cleanupStepResults:
		if m.viewReady {
			s += m.viewport.View() + "\n"
		} else {
			s += m.results.View() + "\n"
		}
		s += curd.RenderHintBar(st, []curd.Hint{
			{Key: "esc/q", Desc: "menu"},
		})
		return s
	}

	return s
}
