package screens

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	repos     []RepoInfo
	cursor    int
	selected  map[int]bool
	caller    string
	loading   bool
	spinner   spinner.Model
	rootPath  string
	winHeight int // visible rows for the repo list
	winOffset int // first visible index
}

func NewRepoSelectModel(caller, rootPath string, height ...int) RepoSelectModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorMagenta)

	h := 0
	if len(height) > 0 && height[0] > 0 {
		h = height[0] - 8
		if h < 5 {
			h = 5
		}
	}

	return RepoSelectModel{
		selected:  make(map[int]bool),
		caller:    caller,
		loading:   true,
		spinner:   s,
		rootPath:  rootPath,
		winHeight: h,
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

func (m *RepoSelectModel) ensureCursorVisible() {
	if m.winHeight <= 0 {
		return
	}
	if m.cursor < m.winOffset {
		m.winOffset = m.cursor
	}
	if m.cursor >= m.winOffset+m.winHeight {
		m.winOffset = m.cursor - m.winHeight + 1
	}
}

func (m RepoSelectModel) Update(msg tea.Msg) (RepoSelectModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Reserve lines for title, help text, and padding
		m.winHeight = msg.Height - 8
		if m.winHeight < 5 {
			m.winHeight = 5
		}
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

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.repos) - 1
				m.winOffset = len(m.repos) - m.winHeight
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

		case key.Matches(msg, key.NewBinding(key.WithKeys(" "))):
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
		return inputBox.Render(prDimStyle.Render("No git repositories found.")) + "\n\n" + menuHelpBox.Render("esc back")
	}

	var s string
	s += titleStyle.Render("Select Repos") + "\n\n"

	// Determine visible window
	visibleStart := m.winOffset
	visibleEnd := visibleStart + m.winHeight
	if visibleEnd > len(m.repos) || m.winHeight <= 0 {
		visibleEnd = len(m.repos)
	}
	if visibleStart > len(m.repos) {
		visibleStart = 0
	}

	// Show scroll-up indicator
	if visibleStart > 0 {
		s += helpStyle.Render(fmt.Sprintf("  ↑ %d more above", visibleStart)) + "\n"
	}

	for i := visibleStart; i < visibleEnd; i++ {
		r := m.repos[i]
		isCursor := i == m.cursor
		isSelected := m.selected[i]

		check := uncheckStyle.Render("○")
		if isSelected {
			check = checkStyle.Render("●")
		}

		// Name style: cursor > selected > unselected
		name := repoUnselectedName.Render(r.Name)
		if isCursor {
			name = repoCursorName.Render(r.Name)
		} else if isSelected {
			name = repoSelectedName.Render(r.Name)
		}

		// Info badges — dim when unselected and not cursor
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

	// Show scroll-down indicator
	remaining := len(m.repos) - visibleEnd
	if remaining > 0 {
		s += helpStyle.Render(fmt.Sprintf("  ↓ %d more below", remaining)) + "\n"
	}

	selected := 0
	for _, v := range m.selected {
		if v {
			selected++
		}
	}

	s += "\n" + menuHelpBox.Render(fmt.Sprintf(
		"space toggle  •  ctrl+a all  •  enter confirm (%d/%d selected)  •  esc back",
		selected, len(m.repos),
	))

	return s
}
