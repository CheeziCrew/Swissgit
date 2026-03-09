package screens

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/CheeziCrew/swissgit/ops"
)

var (
	tpRepoName   = lipgloss.NewStyle().Foreground(colorBrMag).Bold(true)
	tpRepoCount  = lipgloss.NewStyle().Foreground(colorMagenta)
	tpAuthor     = lipgloss.NewStyle().Foreground(colorCyan)
	tpLink       = lipgloss.NewStyle().Foreground(colorBrBlue).Underline(true)
	tpDim        = lipgloss.NewStyle().Foreground(colorGray)
	tpBullet     = lipgloss.NewStyle().Foreground(colorCyan).Render("▸")
	tpBotBullet  = lipgloss.NewStyle().Foreground(colorGray).Render("▸")
	tpDot        = lipgloss.NewStyle().Foreground(colorGray).Render("·")
	tpDraftMark  = "✏️"
	tpBotMark    = "🤖"
	tpAgeFresh   = lipgloss.NewStyle().Foreground(colorGray)
	tpAgeStale   = lipgloss.NewStyle().Foreground(colorYellow)
	tpAgeOld     = lipgloss.NewStyle().Foreground(colorRed)
)

// formatAge returns a human-readable age string with color based on staleness.
func formatAge(created time.Time) string {
	d := time.Since(created)
	days := int(d.Hours() / 24)

	var label string
	switch {
	case days == 0:
		label = "today"
	case days == 1:
		label = "1d"
	case days < 7:
		label = fmt.Sprintf("%dd", days)
	case days < 30:
		label = fmt.Sprintf("%dw", days/7)
	default:
		label = fmt.Sprintf("%dmo", days/30)
	}

	style := tpAgeFresh
	switch {
	case days > 30:
		style = tpAgeOld
	case days > 7:
		style = tpAgeStale
	}
	return style.Render(label)
}

type teamPRsStep int

const (
	teamPRsStepInput teamPRsStep = iota
	teamPRsStepFetching
	teamPRsStepResults
)

type teamPRsReposFetchedMsg struct {
	repos []string
	err   error
}

type teamPRsFetchedMsg struct {
	prs []ops.TeamPR
	err error
}

// TeamPRsModel handles the team-prs flow.
type TeamPRsModel struct {
	step teamPRsStep

	orgInput   textinput.Model
	teamInput  textinput.Model
	focusIndex int

	org  string
	team string

	spinner   spinner.Model
	prs       []ops.TeamPR
	viewport  viewport.Model
	viewReady bool
	fetchMsg  string
	height    int
}

func NewTeamPRsModel() TeamPRsModel {
	oi := newStyledInput("GitHub org name")
	oi.SetValue("Sundsvallskommun")
	oi.Focus()
	oi.CharLimit = 100

	ti := newStyledInput("Team slug")
	ti.SetValue("team-unmasked")
	ti.CharLimit = 100

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorMagenta)

	return TeamPRsModel{
		step:      teamPRsStepInput,
		orgInput:  oi,
		teamInput: ti,
		spinner:   s,
	}
}

func (m TeamPRsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TeamPRsModel) Update(msg tea.Msg) (TeamPRsModel, tea.Cmd) {
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.height = wsm.Height
		if !m.viewReady {
			m.viewport = viewport.New(viewport.WithWidth(wsm.Width-6), viewport.WithHeight(wsm.Height-14))
			m.viewReady = true
		} else {
			m.viewport.SetWidth(wsm.Width - 6)
			m.viewport.SetHeight(wsm.Height - 10)
		}
	}

	switch m.step {
	case teamPRsStepInput:
		return m.updateInput(msg)
	case teamPRsStepFetching:
		return m.updateFetching(msg)
	case teamPRsStepResults:
		return m.updateResults(msg)
	}
	return m, nil
}

func (m TeamPRsModel) updateInput(msg tea.Msg) (TeamPRsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			m.focusIndex = (m.focusIndex + 1) % 2
			return m, m.focusActive()
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			org := m.orgInput.Value()
			team := m.teamInput.Value()
			if org == "" || team == "" {
				return m, nil
			}
			m.org = org
			m.team = team
			m.step = teamPRsStepFetching
			m.fetchMsg = "Fetching team repos…"
			return m, tea.Batch(m.spinner.Tick, m.fetchTeamRepos())
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}

	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.orgInput, cmd = m.orgInput.Update(msg)
	case 1:
		m.teamInput, cmd = m.teamInput.Update(msg)
	}
	return m, cmd
}

func (m *TeamPRsModel) focusActive() tea.Cmd {
	m.orgInput.Blur()
	m.teamInput.Blur()

	switch m.focusIndex {
	case 0:
		m.orgInput.Focus()
	case 1:
		m.teamInput.Focus()
	}
	return textinput.Blink
}

func (m TeamPRsModel) fetchTeamRepos() tea.Cmd {
	org := m.org
	team := m.team
	return func() tea.Msg {
		excludePrefixes := []string{"web-app-", "Camunda-"}
		repos, err := ops.FetchTeamRepoNames(org, team, excludePrefixes)
		return teamPRsReposFetchedMsg{repos: repos, err: err}
	}
}

func (m TeamPRsModel) updateFetching(msg tea.Msg) (TeamPRsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case teamPRsReposFetchedMsg:
		if msg.err != nil {
			m.step = teamPRsStepResults
			m.viewport.SetContent(tpDim.Render(fmt.Sprintf("Failed to fetch team repos: %s", msg.err)))
			return m, nil
		}
		if len(msg.repos) == 0 {
			m.step = teamPRsStepResults
			m.viewport.SetContent(tpDim.Render("No repos found for this team."))
			return m, nil
		}

		m.fetchMsg = fmt.Sprintf("Found %d repos, searching PRs…", len(msg.repos))
		return m, m.fetchPRs(msg.repos)

	case teamPRsFetchedMsg:
		if msg.err != nil {
			m.step = teamPRsStepResults
			m.viewport.SetContent(tpDim.Render(fmt.Sprintf("Failed to fetch PRs: %s", msg.err)))
			return m, nil
		}
		m.prs = msg.prs
		m.step = teamPRsStepResults
		m.viewport.SetContent(m.renderTable())
		return m, nil

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

func (m TeamPRsModel) fetchPRs(repos []string) tea.Cmd {
	org := m.org
	return func() tea.Msg {
		prs, err := ops.FetchTeamPRs(org, repos)
		return teamPRsFetchedMsg{prs: prs, err: err}
	}
}

func (m TeamPRsModel) updateResults(msg tea.Msg) (TeamPRsModel, tea.Cmd) {
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

func prLink(url, text string) string {
	return tpLink.Hyperlink(url).Render(text)
}

func (m TeamPRsModel) renderTable() string {
	if len(m.prs) == 0 {
		return tpDim.Render("No open PRs found for this team.")
	}

	sort.Slice(m.prs, func(i, j int) bool {
		if m.prs[i].Repo != m.prs[j].Repo {
			return m.prs[i].Repo < m.prs[j].Repo
		}
		return m.prs[i].Number < m.prs[j].Number
	})

	grouped := make(map[string][]ops.TeamPR)
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
			bullet := tpBullet
			title := prLink(pr.URL, pr.Title)
			meta := tpAuthor.Render(pr.Author)
			if ops.IsBot(pr.Author) {
				bullet = tpBotBullet
				meta = tpBotMark
			}
			if pr.Draft {
				meta += " " + tpDraftMark
			}
			if !pr.CreatedAt.IsZero() {
				meta += " " + tpDot + " " + formatAge(pr.CreatedAt)
			}
			b.WriteString(fmt.Sprintf("    %s %s %s %s\n", bullet, title, tpDot, meta))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m TeamPRsModel) View() string {
	var s string
	s += titleStyle.Render("👥 Team PRs") + "\n\n"

	switch m.step {
	case teamPRsStepInput:
		labels := []string{"Organization", "Team"}
		inputs := []string{m.orgInput.View(), m.teamInput.View()}

		var content string
		for i, label := range labels {
			marker := "  "
			if i == m.focusIndex {
				marker = cursorMark.Render("▸ ")
			}
			content += fmt.Sprintf("%s%s\n  %s\n\n", marker, prLabelStyle.Render(label), inputs[i])
		}
		s += inputBox.Render(content)
		return s

	case teamPRsStepFetching:
		s += summaryBlock(
			summaryLine("org", m.org),
			summaryLine("team", m.team),
		)
		content := fmt.Sprintf("%s %s", m.spinner.View(), m.fetchMsg)
		s += inputBox.Render(content)

	case teamPRsStepResults:
		s += summaryBlock(
			summaryLine("org", m.org),
			summaryLine("team", m.team),
			summaryLine("open PRs", fmt.Sprintf("%d", len(m.prs))),
		)
		if m.viewReady {
			s += m.viewport.View() + "\n"
		}
		return s
	}

	return s
}
