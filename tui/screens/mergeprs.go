package screens

import (
	"fmt"
	"os"
	"time"

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

const (
	defaultBatchSize = 5
	defaultWaitMin   = 10
)

type mergePRsStep int

const (
	mergePRsStepInput mergePRsStep = iota
	mergePRsStepFetching
	mergePRsStepProgress
	mergePRsStepWaiting
	mergePRsStepResults
)

type mergePRsFetchedMsg struct {
	prs []ops.PRInfo
	err error
}

type mergePRTaskDoneMsg struct {
	index  int
	result ops.MergePRResult
}

type mergeWaitTickMsg struct{}

// MergePRsModel handles the merge-prs flow with batch + wait.
type MergePRsModel struct {
	step mergePRsStep

	orgInput textinput.Model
	org      string

	spinner   spinner.Model
	prs       []ops.PRInfo
	progress  components.ProgressModel
	viewport  viewport.Model
	viewReady bool
	height    int

	// Batching state
	batchSize  int
	waitMin    int
	batchIndex int // which cycle we're on (1-based)

	// Cycle-level tracking (reset each cycle)
	cycleMerged      int               // successful merges in current cycle
	cycleTarget      int               // = batchSize
	cycleFailed      map[string]string  // "repo #number" -> error message
	cycleFailedOrder []string           // ordered keys for display
	cycleRetry       bool              // true when re-fetching within same cycle

	// Waiting countdown
	waitRemaining int // seconds left

	// Accumulate results across all batches
	allResults []components.RepoTask
	merged     int
	failed     int
}

func NewMergePRsModel() MergePRsModel {
	defaultOrg := os.Getenv("GITHUB_ORG")
	if defaultOrg == "" {
		defaultOrg = "Sundsvallskommun"
	}

	oi := newStyledInput("GitHub org name")
	oi.SetValue(defaultOrg)
	oi.Focus()
	oi.CharLimit = 100

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorMagenta)

	return MergePRsModel{
		step:        mergePRsStepInput,
		orgInput:    oi,
		spinner:     s,
		batchSize:   defaultBatchSize,
		waitMin:     defaultWaitMin,
		cycleFailed: make(map[string]string),
	}
}

func (m MergePRsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m MergePRsModel) Update(msg tea.Msg) (MergePRsModel, tea.Cmd) {
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
	case mergePRsStepInput:
		return m.updateInput(msg)
	case mergePRsStepFetching:
		return m.updateFetching(msg)
	case mergePRsStepProgress:
		return m.updateProgress(msg)
	case mergePRsStepWaiting:
		return m.updateWaiting(msg)
	case mergePRsStepResults:
		return m.updateResults(msg)
	}
	return m, nil
}

func (m MergePRsModel) updateInput(msg tea.Msg) (MergePRsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			val := m.orgInput.Value()
			if val == "" {
				return m, nil
			}
			m.org = val
			m.step = mergePRsStepFetching
			return m, tea.Batch(m.spinner.Tick, m.fetchPRs())
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}
	var cmd tea.Cmd
	m.orgInput, cmd = m.orgInput.Update(msg)
	return m, cmd
}

func (m MergePRsModel) fetchPRs() tea.Cmd {
	org := m.org
	return func() tea.Msg {
		prs, err := ops.FetchApprovedPRs(org)
		return mergePRsFetchedMsg{prs: prs, err: err}
	}
}

func (m MergePRsModel) updateFetching(msg tea.Msg) (MergePRsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case mergePRsFetchedMsg:
		if msg.err != nil {
			m.allResults = append(m.allResults, components.RepoTask{
				Name:   "fetch PRs",
				Status: components.TaskFailed,
				Error:  msg.err.Error(),
			})
			return m.goToResults()
		}

		if len(msg.prs) == 0 {
			if len(m.allResults) == 0 {
				m.allResults = append(m.allResults, components.RepoTask{
					Name:   "search",
					Status: components.TaskDone,
					Result: "no approved PRs found",
				})
			}
			return m.goToResults()
		}

		m.prs = msg.prs

		// Fresh cycle: reset cycle tracking
		if !m.cycleRetry {
			m.batchIndex++
			m.cycleMerged = 0
			m.cycleTarget = m.batchSize
			m.cycleFailed = make(map[string]string)
			m.cycleFailedOrder = nil
		}

		// Check if any PRs are eligible (not already failed this cycle)
		hasEligible := false
		for _, pr := range m.prs {
			if _, failed := m.cycleFailed[prKey(pr)]; !failed {
				hasEligible = true
				break
			}
		}

		if !hasEligible {
			// No eligible PRs — cycle is done
			if m.cycleRetry {
				// We were retrying within a cycle; enter wait for next cycle
				return m, m.startWait()
			}
			// Fresh fetch returned nothing eligible — we're done
			return m.goToResults()
		}

		return m, m.startBatch()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyPressMsg:
		if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
			return m.goToResults()
		}
	}
	return m, nil
}

func (m *MergePRsModel) startBatch() tea.Cmd {
	remaining := m.cycleTarget - m.cycleMerged

	// Filter out PRs that already failed in this cycle
	var eligible []ops.PRInfo
	for _, pr := range m.prs {
		if _, failed := m.cycleFailed[prKey(pr)]; !failed {
			eligible = append(eligible, pr)
		}
	}

	// Take up to 'remaining' from eligible
	end := remaining
	if end > len(eligible) {
		end = len(eligible)
	}
	batch := eligible[:end]

	var tasks []components.RepoTask
	for _, pr := range batch {
		tasks = append(tasks, components.RepoTask{
			Name:   prKey(pr),
			Status: components.TaskRunning,
		})
	}

	m.progress = components.NewProgressModel(tasks)
	m.step = mergePRsStepProgress

	var cmds []tea.Cmd
	cmds = append(cmds, m.progress.Init())

	for i, pr := range batch {
		idx := i
		info := pr
		org := m.org

		cmds = append(cmds, func() tea.Msg {
			result := ops.MergePR(org, info.Repo, info.Number)
			result.Title = info.Title
			return mergePRTaskDoneMsg{index: idx, result: result}
		})
	}

	return tea.Batch(cmds...)
}

func (m MergePRsModel) updateProgress(msg tea.Msg) (MergePRsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case mergePRTaskDoneMsg:
		status := components.TaskDone
		resultStr := msg.result.Title
		errStr := ""
		if !msg.result.Success {
			status = components.TaskFailed
			errStr = msg.result.Error
		}

		updateMsg := components.RepoTaskUpdateMsg{
			Index: msg.index, Status: status, Result: resultStr, Error: errStr,
		}
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(updateMsg)
		return m, cmd

	case components.AllTasksDoneMsg:
		for _, t := range m.progress.Tasks {
			m.allResults = append(m.allResults, t)
			switch t.Status {
			case components.TaskDone:
				m.merged++
				m.cycleMerged++
			case components.TaskFailed:
				m.failed++
				if _, exists := m.cycleFailed[t.Name]; !exists {
					m.cycleFailed[t.Name] = t.Error
					m.cycleFailedOrder = append(m.cycleFailedOrder, t.Name)
				}
			}
		}

		// Cycle target met — wait before next cycle
		if m.cycleMerged >= m.cycleTarget {
			return m, m.startWait()
		}

		// Still need more merges — re-fetch immediately within the same cycle
		m.cycleRetry = true
		m.step = mergePRsStepFetching
		return m, tea.Batch(m.spinner.Tick, m.fetchPRs())
	}

	var cmd tea.Cmd
	m.progress, cmd = m.progress.Update(msg)
	return m, cmd
}

func (m *MergePRsModel) startWait() tea.Cmd {
	m.step = mergePRsStepWaiting
	m.waitRemaining = m.waitMin * 60
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return mergeWaitTickMsg{}
	})
}

func (m MergePRsModel) updateWaiting(msg tea.Msg) (MergePRsModel, tea.Cmd) {
	switch msg.(type) {
	case mergeWaitTickMsg:
		m.waitRemaining--
		if m.waitRemaining <= 0 {
			m.cycleRetry = false
			m.step = mergePRsStepFetching
			return m, tea.Batch(m.spinner.Tick, m.fetchPRs())
		}
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return mergeWaitTickMsg{}
		})
	case tea.KeyPressMsg:
		msg := msg.(tea.KeyPressMsg)
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			m.cycleRetry = false
			m.step = mergePRsStepFetching
			return m, tea.Batch(m.spinner.Tick, m.fetchPRs())
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
			return m.goToResults()
		}
	}
	return m, nil
}

func (m *MergePRsModel) goToResults() (MergePRsModel, tea.Cmd) {
	results := components.NewResultModel("Merge PRs", m.allResults)
	m.step = mergePRsStepResults
	m.viewport.SetContent(results.View())
	return *m, nil
}

func (m MergePRsModel) updateResults(msg tea.Msg) (MergePRsModel, tea.Cmd) {
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

func (m MergePRsModel) View() string {
	var s string
	s += titleStyle.Render("🔀 Merge PRs") + "\n\n"

	switch m.step {
	case mergePRsStepInput:
		var content string
		content += prLabelStyle.Render("GitHub organization") + "\n"
		content += m.orgInput.View()
		s += inputBox.Render(content)
		s += curd.RenderHintBar(st, []curd.Hint{
			{Key: "enter", Desc: "submit"},
			{Key: "esc", Desc: "back"},
		})
		return s

	case mergePRsStepFetching:
		s += m.summaryView()
		label := "Searching for approved PRs…"
		if m.batchIndex > 0 {
			label = "Re-fetching approved PRs…"
		}
		content := fmt.Sprintf("%s %s", m.spinner.View(), label)
		s += inputBox.Render(content)

	case mergePRsStepProgress:
		s += m.summaryView()
		s += m.progress.View()

	case mergePRsStepWaiting:
		s += m.summaryView()
		min := m.waitRemaining / 60
		sec := m.waitRemaining % 60
		waitStr := fmt.Sprintf("⏳ Waiting %d:%02d before next batch…", min, sec)
		s += inputBox.Render(waitStr)
		s += curd.RenderHintBar(st, []curd.Hint{
			{Key: "enter", Desc: "skip wait"},
			{Key: "esc", Desc: "back"},
		})
		return s

	case mergePRsStepResults:
		if m.viewReady {
			s += m.viewport.View() + "\n"
		}
		s += curd.RenderHintBar(st, []curd.Hint{
			{Key: "esc/q", Desc: "menu"},
		})
		return s
	}

	return s
}

func prKey(pr ops.PRInfo) string {
	return fmt.Sprintf("%s #%d", pr.Repo, pr.Number)
}

func (m MergePRsModel) summaryView() string {
	lines := []string{
		summaryLine("org", m.org),
	}
	if m.batchIndex > 0 {
		cycleInfo := fmt.Sprintf("#%d (%d merged", m.batchIndex, m.cycleMerged)
		if len(m.cycleFailed) > 0 {
			cycleInfo += fmt.Sprintf(", %d failed", len(m.cycleFailed))
		}
		cycleInfo += fmt.Sprintf(", wait %dm)", m.waitMin)
		lines = append(lines, summaryLine("cycle", cycleInfo))
	}
	if m.merged > 0 || m.failed > 0 {
		lines = append(lines, summaryLine("total", fmt.Sprintf("%d merged, %d failed", m.merged, m.failed)))
	}

	s := summaryBlock(lines...)

	// Show per-PR failure details for the current cycle
	if len(m.cycleFailedOrder) > 0 {
		failStyle := lipgloss.NewStyle().Foreground(colorRed)
		s += "\n"
		for _, name := range m.cycleFailedOrder {
			errMsg := m.cycleFailed[name]
			if len(errMsg) > 80 {
				errMsg = errMsg[len(errMsg)-77:] + "..."
			}
			s += fmt.Sprintf("  %s %s\n", failStyle.Render("✗"), name)
			s += fmt.Sprintf("    %s\n", descStyle.Render(errMsg))
		}
	}

	return s
}
