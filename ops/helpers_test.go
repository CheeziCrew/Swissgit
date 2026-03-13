package ops

import (
	"testing"
)

func TestGetRepoNameForPath_InvalidPath(t *testing.T) {
	_, err := GetRepoNameForPath("/nonexistent/repo/path/for/test")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}
