package screens

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/CheeziCrew/swissgit/git"
)

var (
	dirtyStyle = lipgloss.NewStyle().Foreground(colorYellow)
	cleanMark  = lipgloss.NewStyle().Foreground(colorGreen)
	branchMark = lipgloss.NewStyle().Foreground(colorMagenta)

	// Cursor row — thick left border accent
	repoActiveItem = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), false, false, false, true).
			BorderForeground(colorBlue).
			PaddingLeft(1)

	// Non-cursor rows — just padding to align
	repoInactiveItem = lipgloss.NewStyle().
				PaddingLeft(3)

	// Repo name when cursor is on it — use bright cyan to stand out from selected
	repoCursorName = lipgloss.NewStyle().
			Foreground(colorBlue).
			Bold(true)

	// Repo name when selected (checked) but not cursor
	repoSelectedName = lipgloss.NewStyle().
				Foreground(colorMagenta).
				Bold(true)

	// Repo name when unselected — dimmed
	repoUnselectedName = lipgloss.NewStyle().
				Foreground(colorGray)
)

// reposScanResultMsg is sent after scanning completes.
type reposScanResultMsg struct {
	repos []RepoInfo
}

// RepoSelectModel lets the user pick repos from discovered subdirectories.
type RepoSelectModel struct {
	repos        []RepoInfo
	cursor       int
	selected     map[int]bool
	caller       string
	loading      bool
	spinner      spinner.Model
	rootPath     string
	termHeight   int // raw terminal height
	winOffset    int // first visible index
	parentOffset int // lines consumed by parent screen + app padding
}

// measureOwnChrome returns the exact number of lines reposelect's non-repo
// content occupies (title + scroll indicators).
func measureOwnChrome(scrollable bool) int {
	header := titleStyle.Render("Select Repos") + "\n"
	chrome := lipgloss.Height(header)
	if scrollable {
		chrome += 2 // scroll-up + scroll-down indicator lines
	}
	return chrome
}

// NewRepoSelectModel creates a repo selector.
// parentOffset = lines the parent screen renders above/below this component
// (title, summary box, app.go padding, etc.)
// termHeight = current terminal height so we can size the list immediately.
func NewRepoSelectModel(caller, rootPath string, parentOffset, termHeight int) RepoSelectModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorMagenta)

	return RepoSelectModel{
		selected:     make(map[int]bool),
		caller:       caller,
		loading:      true,
		spinner:      s,
		rootPath:     rootPath,
		parentOffset: parentOffset,
		termHeight:   termHeight,
	}
}

func (m RepoSelectModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.scanRepos())
}

func (m RepoSelectModel) scanRepos() tea.Cmd {
	root := m.rootPath
	return func() tea.Msg {
		if root == "" {
			root = "."
		}

		paths, err := git.DiscoverRepos(root)
		if err != nil {
			return reposScanResultMsg{repos: nil}
		}

		var repos []RepoInfo
		for _, p := range paths {
			name := filepath.Base(p)
			repoName, err := git.GetRepoName(p)
			if err == nil {
				name = repoName
			}

			changes, _ := git.CountChangesShell(p)
			defaultBranch := git.DefaultBranch(p, "main")
			branch := getBranchShell(p)

			repos = append(repos, RepoInfo{
				Path:          p,
				Name:          name,
				Branch:        branch,
				DefaultBranch: defaultBranch,
				Modified:      changes.Modified,
				Added:         changes.Added,
				Deleted:       changes.Deleted,
				Untracked:     changes.Untracked,
				IsDirty:       changes.HasChanges() || (branch != "" && branch != defaultBranch),
			})
		}

		sort.Slice(repos, func(i, j int) bool {
			return repos[i].Name < repos[j].Name
		})

		return reposScanResultMsg{repos: repos}
	}
}

func getBranchShell(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// visibleRepoCount returns how many repo lines fit given the current state.
func (m *RepoSelectModel) visibleRepoCount() int {
	scrollable := len(m.repos) > 10 // assume scrollable for initial sizing
	chrome := measureOwnChrome(scrollable)
	wh := m.termHeight - m.parentOffset - chrome - 1 // -1 for app.go top padding
	if wh < 5 {
		wh = 5
	}
	return wh
}

func (m *RepoSelectModel) ensureCursorVisible() {
	wh := m.visibleRepoCount()
	if m.cursor < m.winOffset {
		m.winOffset = m.cursor
	}
	if m.cursor >= m.winOffset+wh {
		m.winOffset = m.cursor - wh + 1
	}
}

func (m RepoSelectModel) Update(msg tea.Msg) (RepoSelectModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termHeight = msg.Height
		m.ensureCursorVisible()

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case reposScanResultMsg:
		m.loading = false
		m.repos = msg.repos
		// Smart pre-select: check dirty repos
		for i, r := range m.repos {
			if r.IsDirty {
				m.selected[i] = true
			}
		}
		return m, nil

	case tea.KeyPressMsg:
		if m.loading {
			return m, nil
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.repos) - 1
				wh := m.visibleRepoCount()
				m.winOffset = len(m.repos) - wh
				if m.winOffset < 0 {
					m.winOffset = 0
				}
			}
			m.ensureCursorVisible()

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			m.cursor++
			if m.cursor >= len(m.repos) {
				m.cursor = 0
				m.winOffset = 0
			}
			m.ensureCursorVisible()

		case key.Matches(msg, key.NewBinding(key.WithKeys("space"))):
			m.selected[m.cursor] = !m.selected[m.cursor]

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+a"))):
			allSelected := true
			for i := range m.repos {
				if !m.selected[i] {
					allSelected = false
					break
				}
			}
			if allSelected {
				m.selected = make(map[int]bool)
			} else {
				for i := range m.repos {
					m.selected[i] = true
				}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			var paths []string
			for i, r := range m.repos {
				if m.selected[i] {
					paths = append(paths, r.Path)
				}
			}
			return m, func() tea.Msg {
				return RepoSelectDoneMsg{Paths: paths, Caller: m.caller}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}
	return m, nil
}

func (m RepoSelectModel) View() string {
	if m.loading {
		content := fmt.Sprintf("%s Scanning repos…", m.spinner.View())
		return inputBox.Render(content)
	}

	if len(m.repos) == 0 {
		return inputBox.Render(prDimStyle.Render("No git repositories found."))
	}

	// Dynamically compute how many repo lines fit — no hand-counted constants
	wh := m.visibleRepoCount()
	scrollable := len(m.repos) > wh

	visibleStart := m.winOffset
	visibleEnd := visibleStart + wh
	if visibleEnd > len(m.repos) {
		visibleEnd = len(m.repos)
	}
	if visibleStart > len(m.repos) {
		visibleStart = 0
	}

	var s string
	s += titleStyle.Render("Select Repos") + "\n"

	// Always reserve the scroll-up line when scrollable (prevents layout jumping)
	if visibleStart > 0 {
		s += helpStyle.Render(fmt.Sprintf("  ↑ %d more above", visibleStart)) + "\n"
	} else if scrollable {
		s += "\n"
	}

	for i := visibleStart; i < visibleEnd; i++ {
		r := m.repos[i]
		isCursor := i == m.cursor
		isSelected := m.selected[i]

		check := uncheckStyle.Render("○")
		if isSelected {
			check = checkStyle.Render("●")
		}

		name := repoUnselectedName.Render(r.Name)
		if isCursor {
			name = repoCursorName.Render(r.Name)
		} else if isSelected {
			name = repoSelectedName.Render(r.Name)
		}

		var info string
		totalChanges := r.Modified + r.Added + r.Deleted + r.Untracked
		if totalChanges > 0 {
			info = dirtyStyle.Render(fmt.Sprintf(" %dΔ", totalChanges))
		} else if isSelected || isCursor {
			info = cleanMark.Render(" ✓")
		}

		branchInfo := ""
		if r.Branch != "" && r.Branch != r.DefaultBranch {
			branchInfo = branchMark.Render(fmt.Sprintf(" (%s)", r.Branch))
		}

		line := fmt.Sprintf("%s  %s%s%s", check, name, info, branchInfo)
		if isCursor {
			s += repoActiveItem.Render(line) + "\n"
		} else {
			s += repoInactiveItem.Render(line) + "\n"
		}
	}

	// Always reserve the scroll-down line when scrollable (prevents layout jumping)
	remaining := len(m.repos) - visibleEnd
	if remaining > 0 {
		s += helpStyle.Render(fmt.Sprintf("  ↓ %d more below", remaining)) + "\n"
	} else if scrollable {
		s += "\n"
	}

	return s
}
