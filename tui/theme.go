package tui

import "github.com/charmbracelet/lipgloss"

// Base16 ANSI colors — respects terminal theme
var (
	ColorRed   = lipgloss.Color("1")
	ColorGreen = lipgloss.Color("2")
	ColorBlue  = lipgloss.Color("4")
	ColorMag   = lipgloss.Color("5")
	ColorFg    = lipgloss.Color("7")
	ColorGray  = lipgloss.Color("8")
	ColorBrMag = lipgloss.Color("13")
)

// Styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBrMag).
			MarginBottom(1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorGray)
)
