package screens

import "github.com/CheeziCrew/curd"

// Re-export shared message types.
type BackToMenuMsg = curd.BackToMenuMsg
type RepoSelectDoneMsg = curd.RepoSelectDoneMsg
type RepoInfo = curd.RepoInfo

// SaveHistoryMsg is swissgit-specific (not shared).
type SaveHistoryMsg struct {
	Category string
	Value    string
}
