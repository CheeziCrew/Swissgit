package screens

import (
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

// newStyledInput creates a textinput with our standard theme applied.
func newStyledInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetWidth(60)
	s := textinput.DefaultStyles(true)
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(colorBrMag)
	s.Focused.Text = lipgloss.NewStyle().Foreground(colorFg)
	ti.SetStyles(s)
	return ti
}
