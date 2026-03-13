package screens

import (
	"charm.land/bubbles/v2/textinput"
	"github.com/CheeziCrew/curd"
)

// newStyledInput creates a textinput with the swissgit palette applied.
func newStyledInput(placeholder string) textinput.Model {
	return curd.NewStyledInput(placeholder, palette)
}
