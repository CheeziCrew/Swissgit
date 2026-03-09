package screens

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/CheeziCrew/swissgit/ops"
)

type myPRsStep int

const (
	myPRsStepFetching myPRsStep = iota
	myPRsStepResults
)

type myPRsFetchedMsg struct {
	prs []ops.MyPR
	err error
}

// MyPRsModel handles the my-prs flow.
type MyPRsModel struct {
	step myPRsStep

	spinner   spinner.Model
	prs       []ops.MyPR
	viewport  viewport.Model
	viewReady bool
	height    int
}

func NewMyPRsModel() MyPRsModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorMagenta)

	return MyPRsModel{
		step:    myPRsStepFetching,
		spinner: s,
	}
}

func (m MyPRsModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchMyPRs())
}

func fetchMyPRs() tea.Cmd {
	return func() tea.Msg {
		prs, err := ops.FetchMyPRs()
		return myPRsFetchedMsg{prs: prs, err: err}
	}
}

func (m MyPRsModel) Update(msg tea.Msg) (MyPRsModel, tea.Cmd) {
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.height = wsm.Height
		if !m.viewReady {
			m.viewport = viewport.New(viewport.WithWidth(wsm.Width-6), viewport.WithHeight(wsm.Height-10))
			m.viewReady = true
		} else {
			m.viewport.SetWidth(wsm.Width - 6)
			m.viewport.SetHeight(wsm.Height - 10)
		}
	}

	switch m.step {
	case myPRsStepFetching:
		return m.updateFetching(msg)
	case myPRsStepResults:
		return m.updateResults(msg)
	}
	return m, nil
}

func (m MyPRsModel) updateFetching(msg tea.Msg) (MyPRsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case myPRsFetchedMsg:
		if msg.err != nil {
			m.step = myPRsStepResults
			m.viewport.SetContent(tpDim.Render(fmt.Sprintf("Failed to fetch PRs: %s", msg.err)))
			return m, nil
		}
		m.prs = msg.prs
		m.step = myPRsStepResults
		m.viewport.SetContent(m.renderTable())
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyPressMsg:
		if key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))) {
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}
	return m, nil
}

func (m MyPRsModel) updateResults(msg tea.Msg) (MyPRsModel, tea.Cmd) {
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

func (m MyPRsModel) renderTable() string {
	if len(m.prs) == 0 {
		return tpDim.Render("No open PRs found.")
	}

	sort.Slice(m.prs, func(i, j int) bool {
		if m.prs[i].Repo != m.prs[j].Repo {
			return m.prs[i].Repo < m.prs[j].Repo
		}
		return m.prs[i].Number < m.prs[j].Number
	})

	grouped := make(map[string][]ops.MyPR)
	var repoOrder []string
	for _, pr := range m.prs {
		if _, exists := grouped[pr.Repo]; !exists {
			repoOrder = append(repoOrder, pr.Repo)
		}
		grouped[pr.Repo] = append(grouped[pr.Repo], pr)
	}

	var b strings.Builder
	for i, repo := range repoOrder {
		if i > 0 {
			b.WriteString("\n")
		}
		prs := grouped[repo]
		b.WriteString("  " + tpRepoName.Render(repo) + " " + tpRepoCount.Render(fmt.Sprintf("(%d)", len(prs))) + "\n")
		for _, pr := range prs {
			title := prLink(pr.URL, pr.Title)
			meta := ""
			if pr.Draft {
				meta += " " + tpDraftMark
			}
			if !pr.CreatedAt.IsZero() {
				meta += " " + tpDot + " " + formatAge(pr.CreatedAt)
			}
			b.WriteString(fmt.Sprintf("    %s %s%s\n", tpBullet, title, meta))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m MyPRsModel) View() string {
	var s string
	s += titleStyle.Render("🔖 My PRs") + "\n\n"

	switch m.step {
	case myPRsStepFetching:
		content := fmt.Sprintf("%s Fetching your open PRs…", m.spinner.View())
		s += inputBox.Render(content)

	case myPRsStepResults:
		s += summaryBlock(
			summaryLine("open PRs", fmt.Sprintf("%d", len(m.prs))),
		)
		if m.viewReady {
			s += m.viewport.View() + "\n"
		}
		return s
	}

	return s
}
