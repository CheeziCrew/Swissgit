package screens

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/CheeziCrew/curd"

	"github.com/CheeziCrew/swissgit/git"
	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

type automergeStep int

const (
	automergeStepTarget automergeStep = iota
	automergeStepProgress
	automergeStepResults
)

type automergeTaskDoneMsg struct {
	index  int
	result ops.AutomergeResult
}

// AutomergeModel handles the automerge flow.
// No repo selection — automatically discovers and processes all repos.
type AutomergeModel struct {
	step automergeStep

	targetInput textinput.Model
	target      string

	progress  components.ProgressModel
	results   components.ResultModel
	repos     []string
	height    int
	viewport  viewport.Model
	viewReady bool
}

func NewAutomergeModel() AutomergeModel {
	ti := textinput.New()
	ti.Placeholder = "PR search target (e.g. branch name)"
	ti.Focus()
	ti.CharLimit = 200
	ti.SetWidth(60)

	return AutomergeModel{
		step:        automergeStepTarget,
		targetInput: ti,
	}
}

func (m AutomergeModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m AutomergeModel) Update(msg tea.Msg) (AutomergeModel, tea.Cmd) {
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.height = wsm.Height
		if !m.viewReady {
			m.viewport = viewport.New(viewport.WithWidth(wsm.Width-6), viewport.WithHeight(wsm.Height-10))
			m.viewReady = true
		} else {
			m.viewport.SetWidth(wsm.Width - 6)
			m.viewport.SetHeight(wsm.Height - 10)
		}
	}
	switch m.step {
	case automergeStepTarget:
		return m.updateTarget(msg)
	case automergeStepProgress:
		return m.updateProgress(msg)
	case automergeStepResults:
		return m.updateResults(msg)
	}
	return m, nil
}

func (m AutomergeModel) updateTarget(msg tea.Msg) (AutomergeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			val := m.targetInput.Value()
			if val == "" {
				return m, nil
			}
			m.target = val
			return m, m.startAutomergeTasks()
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}
	var cmd tea.Cmd
	m.targetInput, cmd = m.targetInput.Update(msg)
	return m, cmd
}

func (m *AutomergeModel) startAutomergeTasks() tea.Cmd {
	paths, err := git.DiscoverRepos(".")
	if err != nil || len(paths) == 0 {
		// Nothing to do — go back
		return func() tea.Msg { return BackToMenuMsg{} }
	}
	m.repos = paths

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
	m.step = automergeStepProgress

	var cmds []tea.Cmd
	cmds = append(cmds, m.progress.Init())

	for i, p := range m.repos {
		idx := i
		path := p
		target := m.target

		cmds = append(cmds, func() tea.Msg {
			result := ops.EnableAutomerge(target, path)
			return automergeTaskDoneMsg{index: idx, result: result}
		})
	}

	return tea.Batch(cmds...)
}

func (m AutomergeModel) updateProgress(msg tea.Msg) (AutomergeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case automergeTaskDoneMsg:
		status := components.TaskDone
		resultStr := ""
		errStr := ""
		if !msg.result.Success {
			status = components.TaskFailed
			errStr = msg.result.Error
		} else {
			resultStr = "PR #" + msg.result.PRNumber
		}

		updateMsg := components.RepoTaskUpdateMsg{
			Index: msg.index, Status: status, Result: resultStr, Error: errStr,
		}
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(updateMsg)
		return m, cmd

	case components.AllTasksDoneMsg:
		m.results = components.NewResultModel("Automerge", m.progress.Tasks)
		m.step = automergeStepResults
		m.viewport.SetContent(m.results.View())
		return m, nil
	}

	var cmd tea.Cmd
	m.progress, cmd = m.progress.Update(msg)
	return m, cmd
}

func (m AutomergeModel) updateResults(msg tea.Msg) (AutomergeModel, tea.Cmd) {
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

func (m AutomergeModel) View() string {
	var s string
	s += titleStyle.Render("🔀 Automerge") + "\n\n"

	switch m.step {
	case automergeStepTarget:
		var content string
		content += prLabelStyle.Render("PR head branch (exact match)") + "\n"
		content += m.targetInput.View()
		s += inputBox.Render(content) + "\n\n"
		s += curd.RenderHintBar(st, []curd.Hint{
			{Key: "enter", Desc: "submit"},
			{Key: "esc", Desc: "back"},
		})
		return s

	case automergeStepProgress:
		s += summaryBlock(summaryLine("head", m.target))
		s += m.progress.View()

	case automergeStepResults:
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
