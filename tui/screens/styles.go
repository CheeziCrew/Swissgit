package screens

import (
	"charm.land/lipgloss/v2"
	"github.com/CheeziCrew/curd"
)

// Package-level palette and styles for all screen files.
var (
	palette = curd.SwissgitPalette
	st      = palette.Styles()
)

// Re-export colors that screen files reference directly.
var (
	colorBg      = curd.ColorBg
	colorRed     = curd.ColorRed
	colorGreen   = curd.ColorGreen
	colorYellow  = curd.ColorYellow
	colorBlue    = curd.ColorBlue
	colorMagenta = curd.ColorMagenta
	colorCyan    = curd.ColorCyan
	colorFg      = curd.ColorFg
	colorGray    = curd.ColorGray
	colorBrRed   = curd.ColorBrRed
	colorBrGreen = curd.ColorBrGreen
	colorBrYlow  = curd.ColorBrYellow
	colorBrBlue  = curd.ColorBrBlue
	colorBrMag   = curd.ColorBrMag
	colorBrCyan  = curd.ColorBrCyan
	colorBrWhite = curd.ColorBrWhite
)

// Re-export common styles that screens use.
var (
	titleStyle         = st.Title
	subtitleStyle      = st.Subtitle
	inputBox           = st.InputBox
	helpStyle          = st.HelpMargin
	selectedStyle      = st.Selected
	normalStyle        = st.Normal
	descStyle          = st.Dim
	accentStyle        = st.AccentStyle
	summaryBoxStyle    = st.SummaryBox
	summaryLabelStyle  = st.Dim
	summaryValueStyle  = st.AccentStyle
	prLabelStyle       = lipgloss.NewStyle().Foreground(curd.ColorBrBlue).Bold(true)
	prDimStyle         = st.Dim
	checkStyle         = st.CheckStyle
	uncheckStyle       = st.UncheckStyle
	dirtyStyle         = st.DirtyStyle
	cleanMark          = st.CleanMark
	branchMark         = st.BranchMark
	repoActiveItem     = st.RepoActiveItem
	repoInactiveItem   = st.RepoInactiveItem
	repoCursorName     = st.RepoCursorName
	repoSelectedName   = st.RepoSelectedName
	repoUnselectedName = st.RepoUnselectedName

	// Menu styles used by cleanup, pullrequest, clone etc.
	menuActiveItem   = st.MenuActiveItem
	menuInactiveItem = st.MenuInactiveItem
	menuActiveName   = st.MenuActiveName
	menuInactiveName = st.MenuInactiveName
	menuActiveDesc   = st.MenuActiveDesc
	menuInactiveDesc = st.MenuInactiveDesc
	cursorMark       = st.CursorMark
)
