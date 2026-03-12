package screens

import (
	"github.com/CheeziCrew/curd"
)

// Re-export for the app's use.
type MenuSelectionMsg = curd.MenuSelectionMsg
type MenuModel = curd.MenuModel

func NewMenuModel() curd.MenuModel {
	return curd.NewMenuModel(curd.MenuConfig{
		Banner: []string{
			" ___       _         ___ _ _   ",
			"/ __|_ __ (_)___ ___/ __(_) |_ ",
			"\\__ \\ V  V / (_-<(_-< (_ | |  _|",
			"|___/\\_/\\_/|_/__/__/\\___|_|\\__|",
		},
		Tagline: "your multi-repo sidekick",
		Palette: curd.SwissgitPalette,
		Items: []curd.MenuItem{
			{Icon: "🚀", Name: "Pull Request", Command: "pullrequest", Desc: "commit, push & create PR"},
			{Icon: "🧹", Name: "Cleanup", Command: "cleanup", Desc: "reset, update main, prune branches"},
			{Icon: "📦", Name: "Commit", Command: "commit", Desc: "stage, commit & push changes"},
			{Icon: "📊", Name: "Status", Command: "status", Desc: "check repo status & changes"},
			{Icon: "🌿", Name: "Branches", Command: "branches", Desc: "list local, remote & stale branches"},
			{Icon: "📥", Name: "Clone", Command: "clone", Desc: "clone repo or entire org"},
			{Icon: "🔀", Name: "Automerge", Command: "automerge", Desc: "enable auto-merge on PRs"},
			{Icon: "🔃", Name: "Merge PRs", Command: "mergeprs", Desc: "merge approved pull requests"},
			{Icon: "⚙", Name: "Enable Workflows", Command: "enableworkflows", Desc: "re-enable disabled CI workflows"},
			{Icon: "👥", Name: "Team PRs", Command: "teamprs", Desc: "list open PRs across team repos"},
			{Icon: "🔖", Name: "My PRs", Command: "myprs", Desc: "list your open pull requests"},
		},
	})
}
