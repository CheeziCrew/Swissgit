package utils

import (
	"fmt"
	"time"
)

// ShowSpinner displays a spinner in the terminal until done is signaled
func ShowSpinner(repoName string, done chan bool) {
	spinnerChars := `-\|/`
	i := 0

	for {
		select {
		case <-done:
			// Clear the spinner line
			return
		default:
			// Update the spinner
			fmt.Printf("\r%s [%c] ", repoName, spinnerChars[i%len(spinnerChars)]) // Extra space to cover up leftovers
			i++
			time.Sleep(100 * time.Millisecond)
		}
	}
}
