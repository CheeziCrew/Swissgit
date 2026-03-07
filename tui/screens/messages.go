package screens

// BackToMenuMsg signals the root model to return to the menu.
type BackToMenuMsg struct{}

// SaveHistoryMsg tells the root model to persist a message to history.
type SaveHistoryMsg struct {
	Category string
	Value    string
}

// RepoSelectDoneMsg is sent when repos have been selected.
type RepoSelectDoneMsg struct {
	Paths  []string
	Caller string // which screen requested the selection
}

// RepoInfo holds metadata about a discovered repo.
type RepoInfo struct {
	Path          string
	Name          string
	Branch        string
	DefaultBranch string
	Modified      int
	Added         int
	Deleted       int
	Untracked     int
	Ahead         int
	Behind        int
	IsDirty       bool
}
