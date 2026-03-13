package tui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/CheeziCrew/swissgit/tui/screens"
)

// Screen identifies which screen is active.
type Screen int

const (
	ScreenMenu Screen = iota
	ScreenRepoSelect
	ScreenPullRequest
	ScreenCommit
	ScreenCleanup
	ScreenStatus
	ScreenBranches
	ScreenClone
	ScreenAutomerge
	ScreenMergePRs
	ScreenEnableWorkflows
	ScreenTeamPRs
	ScreenMyPRs
)

// NavigateMsg tells the root model to switch screens.
type NavigateMsg struct {
	Screen Screen
}

// Model is the root Bubble Tea model that routes to sub-screens.
type Model struct {
	current Screen
	menu    screens.MenuModel
	width   int
	height  int
	history *History
	pr         screens.PullRequestModel
	cleanup    screens.CleanupModel
	commit     screens.CommitModel
	status     screens.StatusModel
	branches   screens.BranchesModel
	clone      screens.CloneModel
	automerge       screens.AutomergeModel
	mergePRs        screens.MergePRsModel
	enableWorkflows screens.EnableWorkflowsModel
	teamPRs         screens.TeamPRsModel
	myPRs           screens.MyPRsModel
	repoSelect      screens.RepoSelectModel
}

// New creates a fresh root model.
func New() Model {
	return Model{
		current: ScreenMenu,
		menu:    screens.NewMenuModel(),
		history: LoadHistory(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		if key.Matches(msg, DefaultKeyMap.Quit) && m.current == ScreenMenu {
			return m, tea.Quit
		}

	case NavigateMsg:
		m.current = msg.Screen
		return m, nil

	case screens.MenuSelectionMsg:
		return m, m.handleMenuSelection(msg)

	case screens.BackToMenuMsg:
		m.current = ScreenMenu
		m.menu = screens.NewMenuModel()
		return m, func() tea.Msg {
			return tea.WindowSizeMsg{Width: m.width, Height: m.height}
		}

	case screens.SaveHistoryMsg:
		m.history.Add(msg.Category, msg.Value)
		return m, nil

	case screens.StatusActionMsg:
		return m, m.handleStatusAction(msg)
	}

	var cmd tea.Cmd
	switch m.current {
	case ScreenMenu:
		m.menu, cmd = m.menu.Update(msg)
	case ScreenPullRequest:
		m.pr, cmd = m.pr.Update(msg)
	case ScreenCleanup:
		m.cleanup, cmd = m.cleanup.Update(msg)
	case ScreenCommit:
		m.commit, cmd = m.commit.Update(msg)
	case ScreenStatus:
		m.status, cmd = m.status.Update(msg)
	case ScreenBranches:
		m.branches, cmd = m.branches.Update(msg)
	case ScreenClone:
		m.clone, cmd = m.clone.Update(msg)
	case ScreenAutomerge:
		m.automerge, cmd = m.automerge.Update(msg)
	case ScreenMergePRs:
		m.mergePRs, cmd = m.mergePRs.Update(msg)
	case ScreenEnableWorkflows:
		m.enableWorkflows, cmd = m.enableWorkflows.Update(msg)
	case ScreenTeamPRs:
		m.teamPRs, cmd = m.teamPRs.Update(msg)
	case ScreenMyPRs:
		m.myPRs, cmd = m.myPRs.Update(msg)
	case ScreenRepoSelect:
		m.repoSelect, cmd = m.repoSelect.Update(msg)
	}
	return m, cmd
}

func (m Model) View() tea.View {
	var content string
	switch m.current {
	case ScreenMenu:
		content = m.menu.View()
	case ScreenPullRequest:
		content = m.pr.View()
	case ScreenCleanup:
		content = m.cleanup.View()
	case ScreenCommit:
		content = m.commit.View()
	case ScreenStatus:
		content = m.status.View()
	case ScreenBranches:
		content = m.branches.View()
	case ScreenClone:
		content = m.clone.View()
	case ScreenAutomerge:
		content = m.automerge.View()
	case ScreenMergePRs:
		content = m.mergePRs.View()
	case ScreenEnableWorkflows:
		content = m.enableWorkflows.View()
	case ScreenTeamPRs:
		content = m.teamPRs.View()
	case ScreenMyPRs:
		content = m.myPRs.View()
	case ScreenRepoSelect:
		content = m.repoSelect.View()
	}
	v := tea.NewView(lipgloss.NewStyle().Padding(1, 2, 0, 2).Render(content))
	v.AltScreen = true
	return v
}

func (m *Model) handleMenuSelection(msg screens.MenuSelectionMsg) tea.Cmd {
	var initCmd tea.Cmd

	switch msg.Command {
	case "pullrequest":
		m.current = ScreenPullRequest
		m.pr = screens.NewPullRequestModel(m.history.Get("pr_message"))
		initCmd = m.pr.Init()
	case "cleanup":
		m.current = ScreenCleanup
		m.cleanup = screens.NewCleanupModel()
		initCmd = m.cleanup.Init()
	case "commit":
		m.current = ScreenCommit
		m.commit = screens.NewCommitModel(m.history.Get("commit_message"))
		initCmd = m.commit.Init()
	case "status":
		m.current = ScreenStatus
		m.status = screens.NewStatusModel()
		initCmd = m.status.Init()
	case "branches":
		m.current = ScreenBranches
		m.branches = screens.NewBranchesModel()
		initCmd = m.branches.Init()
	case "clone":
		m.current = ScreenClone
		m.clone = screens.NewCloneModel()
		initCmd = m.clone.Init()
	case "automerge":
		m.current = ScreenAutomerge
		m.automerge = screens.NewAutomergeModel()
		initCmd = m.automerge.Init()
	case "mergeprs":
		m.current = ScreenMergePRs
		m.mergePRs = screens.NewMergePRsModel()
		initCmd = m.mergePRs.Init()
	case "enableworkflows":
		m.current = ScreenEnableWorkflows
		m.enableWorkflows = screens.NewEnableWorkflowsModel()
		initCmd = m.enableWorkflows.Init()
	case "teamprs":
		m.current = ScreenTeamPRs
		m.teamPRs = screens.NewTeamPRsModel()
		initCmd = m.teamPRs.Init()
	case "myprs":
		m.current = ScreenMyPRs
		m.myPRs = screens.NewMyPRsModel()
		initCmd = m.myPRs.Init()
	default:
		return nil
	}

	// Every new screen needs the terminal dimensions immediately.
	// WindowSizeMsg is only sent at program start (to the menu) and on resize,
	// so freshly created screens would never know the terminal size without this.
	sizeCmd := func() tea.Msg {
		return tea.WindowSizeMsg{Width: m.width, Height: m.height}
	}

	return tea.Batch(initCmd, sizeCmd)
}

func (m *Model) handleStatusAction(msg screens.StatusActionMsg) tea.Cmd {
	sizeCmd := func() tea.Msg {
		return tea.WindowSizeMsg{Width: m.width, Height: m.height}
	}

	switch msg.Action {
	case "commit":
		m.current = ScreenCommit
		m.commit = screens.NewCommitModel(m.history.Get("commit_message")).WithRepo(msg.Path)
		return tea.Batch(m.commit.Init(), sizeCmd)
	case "pullrequest":
		m.current = ScreenPullRequest
		m.pr = screens.NewPullRequestModel(m.history.Get("pr_message")).WithRepo(msg.Path)
		return tea.Batch(m.pr.Init(), sizeCmd)
	}
	return nil
}
