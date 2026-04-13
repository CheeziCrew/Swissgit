package screens

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/CheeziCrew/curd"

	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

type prStep int

const (
	prStepMessage prStep = iota
	prStepBranch
	prStepChanges
	prStepBreaking
	prStepRepoSelect
	prStepProgress
	prStepResults
)

const fmtTwoCol = "%s  %s"

// prTaskDoneMsg is sent when a single repo's PR task finishes.
type prTaskDoneMsg struct {
	index  int
	result ops.PRResult
}

// PullRequestModel handles the PR creation flow.
type PullRequestModel struct {
	step prStep

	messageInput textinput.Model
	branchInput  textinput.Model

	changeTypes    []string
	changeCursor   int
	changeSelected map[int]bool

	breaking       bool
	breakingConfirm curd.ConfirmModel

	history HistoryBrowser

	repoSelect RepoSelectModel
	progress   components.ProgressModel
	results    components.ResultModel
	viewport   viewport.Model
	viewReady  bool

	message         string
	branch          string
	target          string
	changes         []string
	repos           []string
	throttle        *taskThrottle
	preselectedRepo string
	height          int
}

// WithRepo returns a PullRequestModel pre-configured with a single repo path,
// skipping the repo selection step.
func (m PullRequestModel) WithRepo(path string) PullRequestModel {
	m.preselectedRepo = path
	return m
}

func NewPullRequestModel(recentMessages []string) PullRequestModel {
	mi := newStyledInput("PR title / commit message")
	mi.Focus()
	mi.CharLimit = 200

	bi := newStyledInput("feature branch name (e.g. UF-123)")
	bi.CharLimit = 100

	cfg := ops.LoadConfig()

	return PullRequestModel{
		step:           prStepMessage,
		messageInput:   mi,
		branchInput:    bi,
		target:         cfg.TargetBranch,
		changeTypes:    cfg.ChangeTypes,
		changeSelected: make(map[int]bool),
		history:        NewHistoryBrowser(recentMessages),
		breakingConfirm: curd.NewConfirmModel(curd.ConfirmConfig{
			Question: "Breaking change?",
			Caller:   "breaking",
			Palette:  palette,
		}),
	}
}

func (m PullRequestModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m PullRequestModel) Update(msg tea.Msg) (PullRequestModel, tea.Cmd) {
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
	case prStepMessage:
		return m.updateMessage(msg)
	case prStepBranch:
		return m.updateBranch(msg)
	case prStepChanges:
		return m.updateChanges(msg)
	case prStepBreaking:
		return m.updateBreaking(msg)
	case prStepRepoSelect:
		return m.updateRepoSelect(msg)
	case prStepProgress:
		return m.updateProgress(msg)
	case prStepResults:
		return m.updateResults(msg)
	}
	return m, nil
}

func (m PullRequestModel) updateMessage(msg tea.Msg) (PullRequestModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			val := m.messageInput.Value()
			if val == "" {
				return m, nil
			}
			m.message = val
			m.step = prStepBranch
			m.branchInput.Focus()
			return m, tea.Batch(
				textinput.Blink,
				func() tea.Msg { return SaveHistoryMsg{Category: "pr_message", Value: val} },
			)
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("up"))):
			if m.history.Len() > 0 {
				m.history.BrowseUp(&m.messageInput)
				return m, nil
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down"))):
			if m.history.IsActive() {
				m.history.BrowseDown(&m.messageInput)
				return m, nil
			}
		}
	}
	if _, ok := msg.(tea.KeyPressMsg); ok {
		m.history.Reset()
	}
	var cmd tea.Cmd
	m.messageInput, cmd = m.messageInput.Update(msg)
	return m, cmd
}

func (m PullRequestModel) updateBranch(msg tea.Msg) (PullRequestModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			val := m.branchInput.Value()
			if val == "" {
				return m, nil
			}
			m.branch = val
			m.step = prStepChanges
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.step = prStepMessage
			m.messageInput.Focus()
			return m, textinput.Blink
		}
	}
	var cmd tea.Cmd
	m.branchInput, cmd = m.branchInput.Update(msg)
	return m, cmd
}

func (m *PullRequestModel) collectSelectedChanges() {
	for i, ct := range m.changeTypes {
		if m.changeSelected[i] {
			m.changes = append(m.changes, ct)
		}
	}
}

func (m PullRequestModel) updateChanges(msg tea.Msg) (PullRequestModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			m.changeCursor--
			if m.changeCursor < 0 {
				m.changeCursor = len(m.changeTypes) - 1
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			m.changeCursor++
			if m.changeCursor >= len(m.changeTypes) {
				m.changeCursor = 0
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("space"))):
			m.changeSelected[m.changeCursor] = !m.changeSelected[m.changeCursor]
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			m.collectSelectedChanges()
			m.step = prStepBreaking
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.step = prStepBranch
			m.branchInput.Focus()
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m PullRequestModel) updateBreaking(msg tea.Msg) (PullRequestModel, tea.Cmd) {
	switch msg := msg.(type) {
	case curd.ConfirmMsg:
		m.breaking = msg.Confirmed
		if m.preselectedRepo != "" {
			m.repos = []string{m.preselectedRepo}
			return m, m.startPRTasks()
		}
		m.step = prStepRepoSelect
		preview := titleStyle.Render("🚀 Pull Request") + "\n\n" + m.showSummary()
		m.repoSelect = NewRepoSelectModel("pullrequest", ".", lipgloss.Height(preview), m.height)
		return m, m.repoSelect.Init()
	case curd.BackToMenuMsg:
		m.step = prStepChanges
		return m, nil
	default:
		_ = msg
	}

	var cmd tea.Cmd
	m.breakingConfirm, cmd = m.breakingConfirm.Update(msg)
	return m, cmd
}

func (m PullRequestModel) updateRepoSelect(msg tea.Msg) (PullRequestModel, tea.Cmd) {
	switch msg := msg.(type) {
	case RepoSelectDoneMsg:
		m.repos = msg.Paths
		return m, m.startPRTasks()
	case BackToMenuMsg:
		m.step = prStepBreaking
		return m, nil
	}

	var cmd tea.Cmd
	m.repoSelect, cmd = m.repoSelect.Update(msg)
	return m, cmd
}

func (m *PullRequestModel) startPRTasks() tea.Cmd {
	var tasks []components.RepoTask
	for _, p := range m.repos {
		name := p
		if n, err := ops.GetRepoNameForPath(p); err == nil {
			name = n
		}
		tasks = append(tasks, components.RepoTask{
			Name:   name,
			Path:   p,
			Status: components.TaskPending,
		})
	}

	m.progress = components.NewProgressModel(tasks)
	m.step = prStepProgress

	var taskCmds []tea.Cmd
	for i, p := range m.repos {
		idx := i
		path := p
		branch := m.branch
		message := m.message
		target := m.target
		changes := m.changes
		breaking := m.breaking

		taskCmds = append(taskCmds, func() tea.Msg {
			result := ops.CommitAndCreatePR(path, branch, message, target, changes, breaking)
			return prTaskDoneMsg{index: idx, result: result}
		})
	}

	m.throttle = newThrottle(taskCmds)
	initial := m.throttle.Start(&m.progress)
	return tea.Batch(append([]tea.Cmd{m.progress.Init()}, initial...)...)
}

func (m PullRequestModel) updateProgress(msg tea.Msg) (PullRequestModel, tea.Cmd) {
	switch msg := msg.(type) {
	case prTaskDoneMsg:
		status := components.TaskDone
		result := msg.result.PRURL
		errStr := ""
		if !msg.result.Success {
			status = components.TaskFailed
			errStr = msg.result.Error
		}
		updateMsg := components.RepoTaskUpdateMsg{
			Index:  msg.index,
			Status: status,
			Result: result,
			Error:  errStr,
		}
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(updateMsg)
		if next := m.throttle.Dispatch(&m.progress); next != nil {
			return m, tea.Batch(cmd, next)
		}
		return m, cmd

	case components.AllTasksDoneMsg:
		m.results = components.NewResultModel("Pull Requests", m.progress.Tasks)
		m.step = prStepResults
		m.viewport.SetContent(m.results.View())
		return m, nil
	}

	var cmd tea.Cmd
	m.progress, cmd = m.progress.Update(msg)
	return m, cmd
}

func (m PullRequestModel) updateResults(msg tea.Msg) (PullRequestModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m PullRequestModel) viewChanges() string {
	s := m.showSummary()
	s += prLabelStyle.Render("What changes does this contain?") + "\n\n"
	for i, ct := range m.changeTypes {
		check := uncheckStyle.Render("○")
		if m.changeSelected[i] {
			check = checkStyle.Render("●")
		}
		if i == m.changeCursor {
			s += menuActiveItem.Render(fmt.Sprintf(fmtTwoCol, check, menuActiveName.Render(ct))) + "\n"
		} else {
			s += menuInactiveItem.Render(fmt.Sprintf(fmtTwoCol, check, menuInactiveName.Render(ct))) + "\n"
		}
	}
	s += curd.RenderHintBar(st, []curd.Hint{
		{Key: "space", Desc: "toggle"},
		{Key: "enter", Desc: "confirm"},
		{Key: "esc", Desc: "back"},
	})
	return s
}

func (m PullRequestModel) viewBreaking() string {
	s := m.showSummary()
	s += m.breakingConfirm.View()
	return s
}

func (m PullRequestModel) View() string {
	var s string
	s += titleStyle.Render("🚀 Pull Request") + "\n\n"

	switch m.step {
	case prStepMessage:
		var content string
		content += prLabelStyle.Render("PR title / commit message") + "\n"
		content += m.messageInput.View()
		if m.history.Len() > 0 {
			content += "\n" + helpStyle.Render(fmt.Sprintf("↑↓ recent (%d)", m.history.Len()))
		}
		return s + inputBox.Render(content) + "\n\n" + curd.RenderHintBar(st, []curd.Hint{
			{Key: "enter", Desc: "submit"},
			{Key: "esc", Desc: "back"},
		})

	case prStepBranch:
		s += m.showSummary()
		var content string
		content += prLabelStyle.Render("Feature branch") + "\n"
		content += m.branchInput.View()
		return s + inputBox.Render(content) + "\n\n" + curd.RenderHintBar(st, []curd.Hint{
			{Key: "enter", Desc: "submit"},
			{Key: "esc", Desc: "back"},
		})

	case prStepChanges:
		return s + m.viewChanges()

	case prStepBreaking:
		return s + m.viewBreaking()

	case prStepRepoSelect:
		s += m.showSummary()
		s += m.repoSelect.View()

	case prStepProgress:
		s += m.showSummary()
		s += m.progress.View()

	case prStepResults:
		if m.viewReady {
			s += m.viewport.View() + "\n"
		} else {
			s += m.results.View() + "\n"
		}
		return s + curd.RenderHintBar(st, []curd.Hint{
			{Key: "esc/q", Desc: "menu"},
		})
	}

	return s
}

func (m PullRequestModel) showSummary() string {
	var parts []string
	if m.message != "" {
		parts = append(parts, summaryLine("message", m.message))
	}
	if m.branch != "" {
		parts = append(parts, summaryLine("branch", m.branch))
	}
	if len(m.changes) > 0 && m.step > prStepChanges {
		for _, c := range m.changes {
			// Truncate to fit inside summary box (52 width - 6 padding/border - 4 label)
			if len(c) > 40 {
				c = c[:37] + "…"
			}
			parts = append(parts, summaryLine("  ·", c))
		}
	}
	if len(parts) > 0 {
		return summaryBlock(parts...)
	}
	return ""
}
