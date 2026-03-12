package screens

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/CheeziCrew/curd"

	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

type enableWFStep int

const (
	enableWFStepInput enableWFStep = iota
	enableWFStepFetching
	enableWFStepProgress
	enableWFStepResults
)

const enableWorkflowsTitle = "Enable Workflows"

type enableWFReposFetchedMsg struct {
	repos []string
	err   error
}

type enableWFTaskDoneMsg struct {
	index  int
	result ops.EnableWorkflowResult
}

// EnableWorkflowsModel handles the enable-workflows flow.
type EnableWorkflowsModel struct {
	step enableWFStep

	orgInput      textinput.Model
	workflowInput textinput.Model
	prBranchInput textinput.Model
	focusIndex    int

	org          string
	workflowName string
	prBranch     string

	spinner   spinner.Model
	repos     []string
	progress  components.ProgressModel
	results   components.ResultModel
	viewport  viewport.Model
	viewReady bool
	height    int
}

func NewEnableWorkflowsModel() EnableWorkflowsModel {
	defaultOrg := os.Getenv("GITHUB_ORG")
	if defaultOrg == "" {
		defaultOrg = "Sundsvallskommun"
	}

	oi := newStyledInput("GitHub org name")
	oi.SetValue(defaultOrg)
	oi.Focus()
	oi.CharLimit = 100

	wi := newStyledInput("workflow name (empty = all disabled)")
	wi.SetValue("Call Java CI with Maven")
	wi.CharLimit = 200

	pi := newStyledInput("head branch (empty = skip retrigger)")
	pi.CharLimit = 200

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorMagenta)

	return EnableWorkflowsModel{
		step:          enableWFStepInput,
		orgInput:      oi,
		workflowInput: wi,
		prBranchInput: pi,
		spinner:       s,
	}
}

func (m EnableWorkflowsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m EnableWorkflowsModel) Update(msg tea.Msg) (EnableWorkflowsModel, tea.Cmd) {
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
	case enableWFStepInput:
		return m.updateInput(msg)
	case enableWFStepFetching:
		return m.updateFetching(msg)
	case enableWFStepProgress:
		return m.updateProgress(msg)
	case enableWFStepResults:
		return m.updateResults(msg)
	}
	return m, nil
}

func (m EnableWorkflowsModel) updateInput(msg tea.Msg) (EnableWorkflowsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			m.focusIndex = (m.focusIndex + 1) % 3
			return m, m.focusActive()
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			org := m.orgInput.Value()
			if org == "" {
				return m, nil
			}
			m.org = org
			m.workflowName = m.workflowInput.Value()
			m.prBranch = m.prBranchInput.Value()
			m.step = enableWFStepFetching
			return m, tea.Batch(m.spinner.Tick, m.fetchRepos())
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}

	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.orgInput, cmd = m.orgInput.Update(msg)
	case 1:
		m.workflowInput, cmd = m.workflowInput.Update(msg)
	case 2:
		m.prBranchInput, cmd = m.prBranchInput.Update(msg)
	}
	return m, cmd
}

func (m *EnableWorkflowsModel) focusActive() tea.Cmd {
	m.orgInput.Blur()
	m.workflowInput.Blur()
	m.prBranchInput.Blur()

	switch m.focusIndex {
	case 0:
		m.orgInput.Focus()
	case 1:
		m.workflowInput.Focus()
	case 2:
		m.prBranchInput.Focus()
	}
	return textinput.Blink
}

func (m *EnableWorkflowsModel) fetchRepos() tea.Cmd {
	org := m.org
	return func() tea.Msg {
		repos, err := ops.FetchOrgRepoNames(org)
		return enableWFReposFetchedMsg{repos: repos, err: err}
	}
}

func (m EnableWorkflowsModel) updateFetching(msg tea.Msg) (EnableWorkflowsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case enableWFReposFetchedMsg:
		if msg.err != nil {
			tasks := []components.RepoTask{{
				Name:   "fetch repos",
				Status: components.TaskFailed,
				Error:  msg.err.Error(),
			}}
			m.results = components.NewResultModel(enableWorkflowsTitle, tasks)
			m.step = enableWFStepResults
			m.viewport.SetContent(m.results.View())
			return m, nil
		}

		if len(msg.repos) == 0 {
			tasks := []components.RepoTask{{
				Name:   "search",
				Status: components.TaskDone,
				Result: "no repos found in org",
			}}
			m.results = components.NewResultModel(enableWorkflowsTitle, tasks)
			m.step = enableWFStepResults
			m.viewport.SetContent(m.results.View())
			return m, nil
		}

		m.repos = msg.repos
		return m, m.startTasks()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyPressMsg:
		if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}
	return m, nil
}

func (m *EnableWorkflowsModel) startTasks() tea.Cmd {
	var tasks []components.RepoTask
	for _, repo := range m.repos {
		tasks = append(tasks, components.RepoTask{
			Name:   repo,
			Status: components.TaskRunning,
		})
	}

	m.progress = components.NewProgressModel(tasks)
	m.step = enableWFStepProgress

	var cmds []tea.Cmd
	cmds = append(cmds, m.progress.Init())

	for i, repo := range m.repos {
		idx := i
		repoName := repo
		org := m.org
		wfName := m.workflowName
		prefix := m.prBranch

		cmds = append(cmds, func() tea.Msg {
			result := ops.FindAndEnableWorkflows(org, repoName, wfName, prefix)
			return enableWFTaskDoneMsg{index: idx, result: result}
		})
	}

	return tea.Batch(cmds...)
}

func (m EnableWorkflowsModel) updateProgress(msg tea.Msg) (EnableWorkflowsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case enableWFTaskDoneMsg:
		status := components.TaskDone
		resultStr := ""
		errStr := ""
		if !msg.result.Success {
			status = components.TaskFailed
			errStr = msg.result.Error
		} else if msg.result.EnabledCount > 0 || msg.result.RetriggeredPRs > 0 {
			parts := []string{}
			if msg.result.EnabledCount > 0 {
				parts = append(parts, fmt.Sprintf("enabled %d workflow(s)", msg.result.EnabledCount))
			}
			if msg.result.RetriggeredPRs > 0 {
				parts = append(parts, fmt.Sprintf("retriggered %d PR(s)", msg.result.RetriggeredPRs))
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
		m.results = components.NewResultModel(enableWorkflowsTitle, m.progress.Tasks)
		m.step = enableWFStepResults
		m.viewport.SetContent(m.results.View())
		return m, nil
	}

	var cmd tea.Cmd
	m.progress, cmd = m.progress.Update(msg)
	return m, cmd
}

func (m EnableWorkflowsModel) updateResults(msg tea.Msg) (EnableWorkflowsModel, tea.Cmd) {
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

func (m EnableWorkflowsModel) workflowLabel() string {
	if m.workflowName == "" {
		return "(all disabled)"
	}
	return m.workflowName
}

func (m EnableWorkflowsModel) summaryLines(extras ...string) []string {
	lines := []string{
		summaryLine("org", m.org),
		summaryLine("workflow", m.workflowLabel()),
	}
	lines = append(lines, extras...)
	if m.prBranch != "" {
		lines = append(lines, summaryLine("retrigger", m.prBranch))
	}
	return lines
}

func (m EnableWorkflowsModel) View() string {
	var s string
	s += titleStyle.Render("⚙ Enable Workflows") + "\n\n"

	switch m.step {
	case enableWFStepInput:
		labels := []string{"Organization", "Workflow name", "Retrigger branch"}
		inputs := []string{m.orgInput.View(), m.workflowInput.View(), m.prBranchInput.View()}

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

	case enableWFStepFetching:
		s += summaryBlock(m.summaryLines()...)
		s += inputBox.Render(fmt.Sprintf("%s Fetching repos…", m.spinner.View()))

	case enableWFStepProgress:
		s += summaryBlock(m.summaryLines(summaryLine("repos", fmt.Sprintf("%d", len(m.repos))))...)
		s += m.progress.View()

	case enableWFStepResults:
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
