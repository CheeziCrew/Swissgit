package components

import (
	"testing"

	"github.com/CheeziCrew/curd"
)

func TestNewResultModel(t *testing.T) {
	tasks := []curd.RepoTask{
		{Name: "repo-a", Status: TaskDone, Result: "done"},
		{Name: "repo-b", Status: TaskFailed, Error: "error"},
	}

	m := NewResultModel("Test Results", tasks)
	v := m.View()
	if v == "" {
		t.Error("View() returned empty string")
	}
}

func TestNewResultModel_Empty(t *testing.T) {
	m := NewResultModel("Empty", nil)
	v := m.View()
	if v == "" {
		t.Error("View() returned empty string for empty task list")
	}
}
