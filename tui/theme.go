package tui

import "github.com/CheeziCrew/curd"

// Re-export colors for backward compat in screen files that use them directly.
var (
	ColorRed   = curd.ColorRed
	ColorGreen = curd.ColorGreen
	ColorBlue  = curd.ColorBlue
	ColorMag   = curd.ColorMagenta
	ColorFg    = curd.ColorFg
	ColorGray  = curd.ColorGray
	ColorBrMag = curd.ColorBrMag
)

// App palette and derived styles.
var (
	AppPalette = curd.SwissgitPalette
	Styles     = AppPalette.Styles()
)

// Backward-compat style aliases.
var (
	TitleStyle = Styles.Title
	HelpStyle  = Styles.HelpMargin
)
