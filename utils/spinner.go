package utils

import (
	"fmt"
	"strings"
	"time"
)

// showSpinner displays a spinner in the terminal until done is signaled
func ShowSpinner(repoName string, done chan bool) {
	spinnerChars := `-\|/`
	i := 0
	prevLength := 1

	for {
		select {
		case <-done:
			// Received signal to stop the spinner
			prevLength = len(repoName) + 4
			fmt.Printf("\r%s\r", strings.Repeat(" ", prevLength))
			return
		default:
			// Update the spinner
			fmt.Printf("\r%s [%c]", repoName, spinnerChars[i%len(spinnerChars)])
			i++
			time.Sleep(100 * time.Millisecond)
		}
	}
}
