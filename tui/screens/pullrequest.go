package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	breakingCursor int // 0=no, 1=yes

	recentMessages []string
	historyCursor  int
	typedValue     string

	repoSelect RepoSelectModel
	progress   components.ProgressModel
	results    components.ResultModel
	viewport   viewport.Model
	viewReady  bool

	message string
	branch  string
	target  string
	changes []string
	repos   []string
	height  int
}

func NewPullRequestModel(recentMessages []string) PullRequestModel {
	mi := textinput.New()
	mi.Placeholder = "PR title / commit message"
	mi.Focus()
	mi.CharLimit = 200
	mi.Width = 60
	mi.PromptStyle = lipgloss.NewStyle().Foreground(colorBrMag)
	mi.TextStyle = lipgloss.NewStyle().Foreground(colorFg)

	bi := textinput.New()
	bi.Placeholder = "feature branch name (e.g. UF-123)"
	bi.CharLimit = 100
	bi.Width = 60
	bi.PromptStyle = lipgloss.NewStyle().Foreground(colorBrMag)
	bi.TextStyle = lipgloss.NewStyle().Foreground(colorFg)

	return PullRequestModel{
		step:           prStepMessage,
		messageInput:   mi,
		branchInput:    bi,
		target:         "main",
		changeTypes:    ops.ChangeTypes,
		changeSelected: make(map[int]bool),
		recentMessages: recentMessages,
		historyCursor:  -1,
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
			m.viewport = viewport.New(msg.Width-6, msg.Height-10)
			m.viewReady = true
		} else {
			m.viewport.Width = msg.Width - 6
			m.viewport.Height = msg.Height - 10
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
	case tea.KeyMsg:
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
			if len(m.recentMessages) > 0 {
				if m.historyCursor == -1 {
					m.typedValue = m.messageInput.Value()
				}
				m.historyCursor++
				if m.historyCursor >= len(m.recentMessages) {
					m.historyCursor = len(m.recentMessages) - 1
				}
				m.messageInput.SetValue(m.recentMessages[m.historyCursor])
				m.messageInput.CursorEnd()
				return m, nil
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down"))):
			if m.historyCursor >= 0 {
				m.historyCursor--
				if m.historyCursor < 0 {
					m.messageInput.SetValue(m.typedValue)
				} else {
					m.messageInput.SetValue(m.recentMessages[m.historyCursor])
				}
				m.messageInput.CursorEnd()
				return m, nil
			}
		}
	}
	if _, ok := msg.(tea.KeyMsg); ok {
		m.historyCursor = -1
	}
	var cmd tea.Cmd
	m.messageInput, cmd = m.messageInput.Update(msg)
	return m, cmd
}

func (m PullRequestModel) updateBranch(msg tea.Msg) (PullRequestModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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

func (m PullRequestModel) updateChanges(msg tea.Msg) (PullRequestModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
		case key.Matches(msg, key.NewBinding(key.WithKeys(" "))):
			m.changeSelected[m.changeCursor] = !m.changeSelected[m.changeCursor]
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			for i, ct := range m.changeTypes {
				if m.changeSelected[i] {
					m.changes = append(m.changes, ct)
				}
			}
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
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k", "down", "j"))):
			m.breakingCursor = 1 - m.breakingCursor
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			m.breaking = m.breakingCursor == 1
			m.step = prStepRepoSelect
			m.repoSelect = NewRepoSelectModel("pullrequest", ".", m.height)
			return m, m.repoSelect.Init()
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.step = prStepChanges
			return m, nil
		}
	}
	return m, nil
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
			Status: components.TaskRunning,
		})
	}

	m.progress = components.NewProgressModel(tasks)
	m.step = prStepProgress

	var cmds []tea.Cmd
	cmds = append(cmds, m.progress.Init())

	for i, p := range m.repos {
		idx := i
		path := p
		branch := m.branch
		message := m.message
		target := m.target
		changes := m.changes
		breaking := m.breaking

		cmds = append(cmds, func() tea.Msg {
			result := ops.CommitAndCreatePR(path, branch, message, target, changes, breaking)
			return prTaskDoneMsg{index: idx, result: result}
		})
	}

	return tea.Batch(cmds...)
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
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m PullRequestModel) View() string {
	var s string
	s += titleStyle.Render("🚀 Pull Request") + "\n\n"

	switch m.step {
	case prStepMessage:
		var content string
		content += prLabelStyle.Render("PR title / commit message") + "\n"
		content += m.messageInput.View()
		if len(m.recentMessages) > 0 {
			content += "\n" + helpStyle.Render(fmt.Sprintf("↑↓ recent (%d)", len(m.recentMessages)))
		}
		s += inputBox.Render(content) + "\n\n"
		s += menuHelpBox.Render("enter next  •  esc back")

	case prStepBranch:
		s += m.showSummary()
		var content string
		content += prLabelStyle.Render("Feature branch") + "\n"
		content += m.branchInput.View()
		s += inputBox.Render(content) + "\n\n"
		s += menuHelpBox.Render("enter next  •  esc back")

	case prStepChanges:
		s += m.showSummary()
		s += prLabelStyle.Render("What changes does this contain?") + "\n\n"
		for i, ct := range m.changeTypes {
			check := uncheckStyle.Render("○")
			if m.changeSelected[i] {
				check = checkStyle.Render("●")
			}
			if i == m.changeCursor {
				line := fmt.Sprintf("%s  %s", check, menuActiveName.Render(ct))
				s += menuActiveItem.Render(line) + "\n"
			} else {
				line := fmt.Sprintf("%s  %s", check, menuInactiveName.Render(ct))
				s += menuInactiveItem.Render(line) + "\n"
			}
		}
		s += "\n"
		s += menuHelpBox.Render("space toggle  •  enter next  •  esc back")

	case prStepBreaking:
		s += m.showSummary()

		type breakOption struct {
			label string
			desc  string
		}
		options := []breakOption{
			{label: "No", desc: "safe (default)"},
			{label: "Yes", desc: "breaking change"},
		}

		s += prLabelStyle.Render("Breaking change?") + "\n\n"
		for i, opt := range options {
			line := fmt.Sprintf("%s  %s", menuActiveName.Render(opt.label), menuActiveDesc.Render(opt.desc))
			if i == m.breakingCursor {
				s += menuActiveItem.Render(line) + "\n"
			} else {
				inactiveLine := fmt.Sprintf("%s  %s", menuInactiveName.Render(opt.label), menuInactiveDesc.Render(opt.desc))
				s += menuInactiveItem.Render(inactiveLine) + "\n"
			}
		}
		s += "\n"
		s += menuHelpBox.Render("↑↓ navigate  •  enter select  •  esc back")

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
		scrollHint := ""
		if m.viewReady && m.viewport.TotalLineCount() > m.viewport.VisibleLineCount() {
			scrollHint = "  •  ↑↓ scroll"
		}
		s += menuHelpBox.Render("esc/q back to menu" + scrollHint)
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
