package screens

import "charm.land/bubbles/v2/textinput"

// HistoryBrowser provides arrow-key browsing through a list of recent values.
// Used by commit and pull-request screens for message history.
type HistoryBrowser struct {
	items      []string
	cursor     int
	typedValue string
}

// NewHistoryBrowser creates a HistoryBrowser with the given history items.
// Items are expected newest-first (index 0 = most recent).
func NewHistoryBrowser(items []string) HistoryBrowser {
	return HistoryBrowser{
		items:  items,
		cursor: -1,
	}
}

// BrowseUp moves up through history (older). On first press, saves the current
// input value so it can be restored later.
func (h *HistoryBrowser) BrowseUp(input *textinput.Model) {
	if len(h.items) == 0 {
		return
	}
	if h.cursor == -1 {
		h.typedValue = input.Value()
	}
	h.cursor++
	if h.cursor >= len(h.items) {
		h.cursor = len(h.items) - 1
	}
	input.SetValue(h.items[h.cursor])
	input.CursorEnd()
}

// BrowseDown moves down through history (newer). When cursor drops below 0,
// restores the value the user had typed before browsing. No-op if not active.
func (h *HistoryBrowser) BrowseDown(input *textinput.Model) {
	if h.cursor < 0 {
		return
	}
	h.cursor--
	if h.cursor < 0 {
		h.cursor = -1
		input.SetValue(h.typedValue)
	} else {
		input.SetValue(h.items[h.cursor])
	}
	input.CursorEnd()
}

// IsActive returns true when the user is browsing history (cursor >= 0).
func (h *HistoryBrowser) IsActive() bool {
	return h.cursor >= 0
}

// Reset exits history browsing mode. Call on any non-arrow keypress.
func (h *HistoryBrowser) Reset() {
	h.cursor = -1
}

// Len returns the number of history items.
func (h *HistoryBrowser) Len() int {
	return len(h.items)
}
