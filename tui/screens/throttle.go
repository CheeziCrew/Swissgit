package screens

import (
	tea "charm.land/bubbletea/v2"

	"github.com/CheeziCrew/swissgit/tui/components"
)

const maxConcurrentOps = 5

// taskThrottle dispatches tea.Cmds in batches of maxConcurrentOps.
type taskThrottle struct {
	cmds []tea.Cmd
	next int
}

func newThrottle(cmds []tea.Cmd) *taskThrottle {
	return &taskThrottle{cmds: cmds}
}

// Start returns the initial batch of cmds and marks those tasks as running.
func (t *taskThrottle) Start(progress *components.ProgressModel) []tea.Cmd {
	end := maxConcurrentOps
	if end > len(t.cmds) {
		end = len(t.cmds)
	}
	t.next = end
	for i := 0; i < end; i++ {
		progress.Tasks[i].Status = components.TaskRunning
	}
	return t.cmds[:end]
}

// Dispatch returns the next cmd and marks it as running, or nil if all dispatched.
func (t *taskThrottle) Dispatch(progress *components.ProgressModel) tea.Cmd {
	if t == nil || t.next >= len(t.cmds) {
		return nil
	}
	progress.Tasks[t.next].Status = components.TaskRunning
	cmd := t.cmds[t.next]
	t.next++
	return cmd
}
