package screens

import (
	"fmt"
	"path/filepath"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/CheeziCrew/curd"

	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

type cloneStep int

const (
	cloneStepInput cloneStep = iota
	cloneStepProgress
	cloneStepResults
)

type cloneTaskDoneMsg struct {
	index  int
	result ops.CloneResult
}

type cloneOrgFetchedMsg struct {
	repos []ops.Repository
	err   error
}

// CloneModel handles the clone flow.
type CloneModel struct {
	step cloneStep

	repoInput textinput.Model
	orgInput  textinput.Model
	teamInput textinput.Model
	pathInput textinput.Model

	focusIndex int // 0=repo, 1=org, 2=team, 3=path

	progress  components.ProgressModel
	results   components.ResultModel
	viewport  viewport.Model
	viewReady bool
	height    int
}

func NewCloneModel() CloneModel {
	ri := textinput.New()
	ri.Placeholder = "repo SSH URL (or leave empty for org clone)"
	ri.Focus()
	ri.CharLimit = 300
	ri.SetWidth(60)

	oi := textinput.New()
	oi.Placeholder = "GitHub org name"
	oi.SetValue("Sundsvallskommun")
	oi.CharLimit = 100
	oi.SetWidth(60)

	ti := textinput.New()
	ti.Placeholder = "team name (optional)"
	ti.SetValue("api-team")
	ti.CharLimit = 100
	ti.SetWidth(60)

	pi := textinput.New()
	pi.Placeholder = "destination path (default: .)"
	pi.SetValue(".")
	pi.CharLimit = 200
	pi.SetWidth(60)

	return CloneModel{
		step:      cloneStepInput,
		repoInput: ri,
		orgInput:  oi,
		teamInput: ti,
		pathInput: pi,
	}
}

func (m CloneModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CloneModel) Update(msg tea.Msg) (CloneModel, tea.Cmd) {
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
	case cloneStepInput:
		return m.updateInput(msg)
	case cloneStepProgress:
		return m.updateProgress(msg)
	case cloneStepResults:
		return m.updateResults(msg)
	}
	return m, nil
}

func (m CloneModel) updateInput(msg tea.Msg) (CloneModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			m.focusIndex = (m.focusIndex + 1) % 4
			return m, m.focusActive()
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return m, m.startClone()
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}

	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.repoInput, cmd = m.repoInput.Update(msg)
	case 1:
		m.orgInput, cmd = m.orgInput.Update(msg)
	case 2:
		m.teamInput, cmd = m.teamInput.Update(msg)
	case 3:
		m.pathInput, cmd = m.pathInput.Update(msg)
	}
	return m, cmd
}

func (m *CloneModel) focusActive() tea.Cmd {
	m.repoInput.Blur()
	m.orgInput.Blur()
	m.teamInput.Blur()
	m.pathInput.Blur()

	switch m.focusIndex {
	case 0:
		m.repoInput.Focus()
	case 1:
		m.orgInput.Focus()
	case 2:
		m.teamInput.Focus()
	case 3:
		m.pathInput.Focus()
	}
	return textinput.Blink
}

func (m *CloneModel) startClone() tea.Cmd {
	repoURL := m.repoInput.Value()
	orgName := m.orgInput.Value()
	teamName := m.teamInput.Value()
	destPath := m.pathInput.Value()
	if destPath == "" {
		destPath = "."
	}

	if repoURL != "" {
		// Single repo clone
		var tasks []components.RepoTask
		tasks = append(tasks, components.RepoTask{
			Name:   filepath.Base(repoURL),
			Path:   destPath,
			Status: components.TaskRunning,
		})
		m.progress = components.NewProgressModel(tasks)
		m.step = cloneStepProgress

		return tea.Batch(m.progress.Init(), func() tea.Msg {
			result := ops.CloneFromURL(repoURL, destPath)
			return cloneTaskDoneMsg{index: 0, result: result}
		})
	}

	if orgName != "" {
		// Fetch org repos then clone
		m.step = cloneStepProgress
		return func() tea.Msg {
			repos, err := ops.GetOrgRepositories(orgName, teamName)
			return cloneOrgFetchedMsg{repos: repos, err: err}
		}
	}

	return nil
}

func (m CloneModel) updateProgress(msg tea.Msg) (CloneModel, tea.Cmd) {
	switch msg := msg.(type) {
	case cloneOrgFetchedMsg:
		if msg.err != nil {
			// Create a single failed task
			tasks := []components.RepoTask{{
				Name:   "org fetch",
				Status: components.TaskFailed,
				Error:  msg.err.Error(),
			}}
			m.results = components.NewResultModel("Clone", tasks)
			m.step = cloneStepResults
			m.viewport.SetContent(m.results.View())
			return m, nil
		}

		destPath := m.pathInput.Value()
		if destPath == "" {
			destPath = "."
		}

		var tasks []components.RepoTask
		for _, r := range msg.repos {
			tasks = append(tasks, components.RepoTask{
				Name:   r.Name,
				Path:   filepath.Join(destPath, r.Name),
				Status: components.TaskRunning,
			})
		}
		m.progress = components.NewProgressModel(tasks)

		var cmds []tea.Cmd
		cmds = append(cmds, m.progress.Init())
		for i, r := range msg.repos {
			idx := i
			repo := r
			dest := filepath.Join(destPath, r.Name)
			cmds = append(cmds, func() tea.Msg {
				result := ops.CloneRepository(repo, dest)
				return cloneTaskDoneMsg{index: idx, result: result}
			})
		}
		return m, tea.Batch(cmds...)

	case cloneTaskDoneMsg:
		status := components.TaskDone
		resultStr := ""
		errStr := ""
		if !msg.result.Success {
			status = components.TaskFailed
			errStr = msg.result.Error
		} else if msg.result.Skipped {
			resultStr = "already cloned"
		}

		updateMsg := components.RepoTaskUpdateMsg{
			Index: msg.index, Status: status, Result: resultStr, Error: errStr,
		}
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(updateMsg)
		return m, cmd

	case components.AllTasksDoneMsg:
		m.results = components.NewResultModel("Clone", m.progress.Tasks)
		m.step = cloneStepResults
		m.viewport.SetContent(m.results.View())
		return m, nil
	}

	var cmd tea.Cmd
	m.progress, cmd = m.progress.Update(msg)
	return m, cmd
}

func (m CloneModel) updateResults(msg tea.Msg) (CloneModel, tea.Cmd) {
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

func (m CloneModel) View() string {
	var s string
	s += titleStyle.Render("📥 Clone") + "\n\n"

	switch m.step {
	case cloneStepInput:
		labels := []string{"Repo URL", "Org name", "Team", "Path"}
		inputs := []string{
			m.repoInput.View(),
			m.orgInput.View(),
			m.teamInput.View(),
			m.pathInput.View(),
		}

		var content string
		for i, label := range labels {
			marker := "  "
			if i == m.focusIndex {
				marker = cursorMark.Render("▸ ")
			}
			content += fmt.Sprintf("%s%s\n  %s\n\n", marker, prLabelStyle.Render(label), inputs[i])
		}

		s += inputBox.Render(content)
		s += curd.RenderHintBar(st, []curd.Hint{
			{Key: "tab", Desc: "next"},
			{Key: "enter", Desc: "submit"},
			{Key: "esc", Desc: "back"},
		})
		return s

	case cloneStepProgress:
		s += m.progress.View()

	case cloneStepResults:
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
