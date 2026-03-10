package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/joho/godotenv"

	"github.com/CheeziCrew/swissgit/cli"
	"github.com/CheeziCrew/swissgit/tui"
)

func init() {
	// Load .env from the directory of the executable (not CWD).
	// Resolve symlinks so it works when invoked via e.g. ~/.local/bin/ symlink.
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("Error getting executable path: %v\n", err)
		return
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		log.Printf("Error resolving symlinks: %v\n", err)
		return
	}
	envPath := filepath.Join(filepath.Dir(exePath), ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("No .env file found at %s. Using environment variables from the system.\n", envPath)
	}
}

func main() {
	// If any args are passed (beyond the binary name), use the fast CLI path.
	// Otherwise, launch the full TUI.
	if len(os.Args) > 1 {
		root := cli.BuildCLI()
		if err := root.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	// Launch the Bubble Tea TUI
	p := tea.NewProgram(tui.New())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
