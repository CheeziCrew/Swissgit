package screens

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/CheeziCrew/curd"

	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

type commitStep int

const (
	commitStepMessage commitStep = iota
	commitStepBranch
	commitStepRepoSelect
	commitStepProgress
	commitStepResults
)

type commitTaskDoneMsg struct {
	index  int
	result ops.CommitResult
}

// CommitModel handles the commit flow.
type CommitModel struct {
	step commitStep

	messageInput textinput.Model
	branchInput  textinput.Model

	message        string
	branch         string
	preselectedRepo string

	history HistoryBrowser

	repoSelect RepoSelectModel
	progress   components.ProgressModel
	results    components.ResultModel
	repos      []string
	viewport   viewport.Model
	viewReady  bool
	height     int
}

// WithRepo returns a CommitModel pre-configured with a single repo path,
// skipping the repo selection step.
func (m CommitModel) WithRepo(path string) CommitModel {
	m.preselectedRepo = path
	return m
}

func NewCommitModel(recentMessages []string) CommitModel {
	mi := newStyledInput("commit message")
	mi.Focus()
	mi.CharLimit = 200

	bi := newStyledInput("branch (optional, leave empty for current)")
	bi.CharLimit = 100

	return CommitModel{
		step:         commitStepMessage,
		messageInput: mi,
		branchInput:  bi,
		history:      NewHistoryBrowser(recentMessages),
	}
}

func (m CommitModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CommitModel) Update(msg tea.Msg) (CommitModel, tea.Cmd) {
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
	case commitStepMessage:
		return m.updateMessage(msg)
	case commitStepBranch:
		return m.updateBranch(msg)
	case commitStepRepoSelect:
		return m.updateRepoSelect(msg)
	case commitStepProgress:
		return m.updateProgress(msg)
	case commitStepResults:
		return m.updateResults(msg)
	}
	return m, nil
}

func (m CommitModel) updateMessage(msg tea.Msg) (CommitModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			val := m.messageInput.Value()
			if val == "" {
				return m, nil
			}
			m.message = val
			m.step = commitStepBranch
			m.branchInput.Focus()
			return m, tea.Batch(
				textinput.Blink,
				func() tea.Msg { return SaveHistoryMsg{Category: "commit_message", Value: val} },
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

func (m CommitModel) updateBranch(msg tea.Msg) (CommitModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			m.branch = m.branchInput.Value()
			if m.preselectedRepo != "" {
				m.repos = []string{m.preselectedRepo}
				return m, m.startCommitTasks()
			}
			m.step = commitStepRepoSelect
			parts := []string{summaryLine("message", m.message)}
			if m.branch != "" {
				parts = append(parts, summaryLine("branch", m.branch))
			}
			preview := titleStyle.Render("📦 Commit & Push") + "\n\n" + summaryBlock(parts...)
			m.repoSelect = NewRepoSelectModel("commit", ".", lipgloss.Height(preview), m.height)
			return m, m.repoSelect.Init()
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.step = commitStepMessage
			m.messageInput.Focus()
			return m, textinput.Blink
		}
	}
	var cmd tea.Cmd
	m.branchInput, cmd = m.branchInput.Update(msg)
	return m, cmd
}

func (m CommitModel) updateRepoSelect(msg tea.Msg) (CommitModel, tea.Cmd) {
	switch msg := msg.(type) {
	case RepoSelectDoneMsg:
		m.repos = msg.Paths
		return m, m.startCommitTasks()
	case BackToMenuMsg:
		m.step = commitStepBranch
		m.branchInput.Focus()
		return m, textinput.Blink
	}
	var cmd tea.Cmd
	m.repoSelect, cmd = m.repoSelect.Update(msg)
	return m, cmd
}

func (m *CommitModel) startCommitTasks() tea.Cmd {
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
	m.step = commitStepProgress

	var cmds []tea.Cmd
	cmds = append(cmds, m.progress.Init())

	for i, p := range m.repos {
		idx := i
		path := p
		branch := m.branch
		message := m.message

		cmds = append(cmds, func() tea.Msg {
			result := ops.CommitAndPush(path, branch, message)
			return commitTaskDoneMsg{index: idx, result: result}
		})
	}

	return tea.Batch(cmds...)
}

func (m CommitModel) updateProgress(msg tea.Msg) (CommitModel, tea.Cmd) {
	switch msg := msg.(type) {
	case commitTaskDoneMsg:
		status := components.TaskDone
		errStr := ""
		if !msg.result.Success {
			status = components.TaskFailed
			errStr = msg.result.Error
		}
		updateMsg := components.RepoTaskUpdateMsg{
			Index: msg.index, Status: status, Result: msg.result.Branch, Error: errStr,
		}
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(updateMsg)
		return m, cmd

	case components.AllTasksDoneMsg:
		m.results = components.NewResultModel("Commit & Push", m.progress.Tasks)
		m.step = commitStepResults
		m.viewport.SetContent(m.results.View())
		return m, nil
	}

	var cmd tea.Cmd
	m.progress, cmd = m.progress.Update(msg)
	return m, cmd
}

func (m CommitModel) updateResults(msg tea.Msg) (CommitModel, tea.Cmd) {
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

func (m CommitModel) viewMessage() string {
	var content string
	content += prLabelStyle.Render("Commit message") + "\n"
	content += m.messageInput.View()
	if m.history.Len() > 0 {
		content += "\n" + helpStyle.Render(fmt.Sprintf("↑↓ recent (%d)", m.history.Len()))
	}
	return inputBox.Render(content) + "\n\n" + curd.RenderHintBar(st, []curd.Hint{
		{Key: "enter", Desc: "submit"},
		{Key: "esc", Desc: "back"},
	})
}

func (m CommitModel) viewBranch() string {
	s := summaryBlock(summaryLine("message", m.message))
	var content string
	content += prLabelStyle.Render("Branch (optional)") + "\n"
	content += m.branchInput.View()
	return s + inputBox.Render(content) + "\n\n" + curd.RenderHintBar(st, []curd.Hint{
		{Key: "enter", Desc: "submit"},
		{Key: "esc", Desc: "back"},
	})
}

func (m CommitModel) View() string {
	var s string
	s += titleStyle.Render("📦 Commit & Push") + "\n\n"

	switch m.step {
	case commitStepMessage:
		return s + m.viewMessage()

	case commitStepBranch:
		return s + m.viewBranch()

	case commitStepRepoSelect:
		parts := []string{summaryLine("message", m.message)}
		if m.branch != "" {
			parts = append(parts, summaryLine("branch", m.branch))
		}
		s += summaryBlock(parts...)
		s += m.repoSelect.View()

	case commitStepProgress:
		s += m.progress.View()

	case commitStepResults:
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

func summaryLine(label, value string) string {
	return summaryLabelStyle.Render(label+" ") + summaryValueStyle.Render(value)
}

func summaryBlock(lines ...string) string {
	content := strings.Join(lines, "\n")
	return summaryBoxStyle.Render(content) + "\n\n"
}
