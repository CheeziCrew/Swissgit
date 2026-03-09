package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Base16 ANSI colors — respects terminal theme
var (
	colorBg      = lipgloss.Color("0")  // background
	colorRed     = lipgloss.Color("1")  // errors
	colorGreen   = lipgloss.Color("2")  // success
	colorYellow  = lipgloss.Color("3")  // warnings
	colorBlue    = lipgloss.Color("4")  // info
	colorMagenta = lipgloss.Color("5")  // accent
	colorCyan    = lipgloss.Color("6")  // secondary
	colorFg      = lipgloss.Color("7")  // foreground
	colorGray    = lipgloss.Color("8")  // dim/muted
	colorBrRed   = lipgloss.Color("9")  // bright red
	colorBrGreen = lipgloss.Color("10") // bright green
	colorBrYlow  = lipgloss.Color("11") // bright yellow
	colorBrBlue  = lipgloss.Color("12") // bright blue
	colorBrMag   = lipgloss.Color("13") // bright magenta — primary
	colorBrCyan  = lipgloss.Color("14") // bright cyan
	colorBrWhite = lipgloss.Color("15") // bright white
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBrMag).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorMagenta).
			Italic(true)

	inputBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorGray).
		Padding(0, 2)

	summaryBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMagenta).
			Padding(0, 2).
			Width(52)

	summaryLabelStyle = lipgloss.NewStyle().
				Foreground(colorGray)

	summaryValueStyle = lipgloss.NewStyle().
				Foreground(colorBrMag).
				Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(colorBrMag).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(colorFg)

	descStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			MarginTop(1)

	// Shared styles used by other screens
	prLabelStyle = lipgloss.NewStyle().
			Foreground(colorBrBlue).
			Bold(true)

	prDimStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	accentStyle = lipgloss.NewStyle().
			Foreground(colorBrMag)

	logoBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorMagenta).
		Padding(1, 3).
		MarginBottom(1)

	logoGradientColors = []lipgloss.Color{
		colorBrMag,
		colorMagenta,
		colorBrBlue,
		colorBlue,
	}

	taglineStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Italic(true)

	versionStyle = lipgloss.NewStyle().
			Foreground(colorMagenta).
			Bold(true)

	menuActiveItem = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBrMag).
			Padding(0, 1).
			Width(50)

	menuInactiveItem = lipgloss.NewStyle().
				Padding(0, 2).
				Width(52)

	menuActiveName = lipgloss.NewStyle().
			Foreground(colorBrMag).
			Bold(true)

	menuInactiveName = lipgloss.NewStyle().
				Foreground(colorGray)

	menuActiveDesc = lipgloss.NewStyle().
			Foreground(colorMagenta)

	menuInactiveDesc = lipgloss.NewStyle().
				Foreground(colorGray)

	menuHelpBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGray).
			Padding(0, 2).
			Foreground(colorGray)

	// Shared styles for cursor/check in other screens
	cursorMark   = lipgloss.NewStyle().Foreground(colorBrMag).Bold(true)
	checkStyle   = lipgloss.NewStyle().Foreground(colorBrMag).Bold(true)
	uncheckStyle = lipgloss.NewStyle().Foreground(colorGray)
)

// MenuSelectionMsg is sent when the user picks a command.
type MenuSelectionMsg struct {
	Command string
}

type menuItem struct {
	icon    string
	name    string
	command string
	desc    string
}

// MenuModel is the main command picker.
type MenuModel struct {
	items    []menuItem
	cursor   int
	selected string
}

func NewMenuModel() MenuModel {
	return MenuModel{
		items: []menuItem{
			{icon: "🚀", name: "Pull Request", command: "pullrequest", desc: "commit, push & create PR"},
			{icon: "🧹", name: "Cleanup", command: "cleanup", desc: "reset, update main, prune branches"},
			{icon: "📦", name: "Commit", command: "commit", desc: "stage, commit & push changes"},
			{icon: "📊", name: "Status", command: "status", desc: "check repo status & changes"},
			{icon: "🌿", name: "Branches", command: "branches", desc: "list local, remote & stale branches"},
			{icon: "📥", name: "Clone", command: "clone", desc: "clone repo or entire org"},
			{icon: "🔀", name: "Automerge", command: "automerge", desc: "enable auto-merge on PRs"},
			{icon: "🔃", name: "Merge PRs", command: "mergeprs", desc: "merge approved pull requests"},
			{icon: "⚙", name: "Enable Workflows", command: "enableworkflows", desc: "re-enable disabled CI workflows"},
		},
	}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.items) - 1
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			m.cursor++
			if m.cursor >= len(m.items) {
				m.cursor = 0
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			m.selected = m.items[m.cursor].command
			return m, func() tea.Msg {
				return MenuSelectionMsg{Command: m.selected}
			}
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	var s strings.Builder

	banner := []string{
		" ___       _         ___ _ _   ",
		"/ __|_ __ (_)___ ___/ __(_) |_ ",
		"\\__ \\ V  V / (_-<(_-< (_ | |  _|",
		"|___/\\_/\\_/|_/__/__/\\___|_|\\__|",
	}

	var logoLines []string
	for i, line := range banner {
		ci := i % len(logoGradientColors)
		logoLines = append(logoLines, lipgloss.NewStyle().Foreground(logoGradientColors[ci]).Bold(true).Render(line))
	}
	logoContent := strings.Join(logoLines, "\n")
	logoContent += "\n" + taglineStyle.Render("  your multi-repo sidekick") + "  " + versionStyle.Render("⚙")

	s.WriteString(logoBox.Render(logoContent) + "\n\n")

	for i, item := range m.items {
		if i == m.cursor {
			line := fmt.Sprintf("%s  %s  %s", item.icon, menuActiveName.Render(item.name), menuActiveDesc.Render(item.desc))
			s.WriteString(menuActiveItem.Render(line) + "\n")
		} else {
			line := fmt.Sprintf("%s  %s  %s", item.icon, menuInactiveName.Render(item.name), menuInactiveDesc.Render(item.desc))
			s.WriteString(menuInactiveItem.Render(line) + "\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(menuHelpBox.Render("↑↓ navigate  •  enter select  •  q quit"))

	return s.String()
}
