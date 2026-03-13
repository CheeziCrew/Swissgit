package components

import (
	"testing"

	"github.com/CheeziCrew/curd"
)

func TestNewProgressModel(t *testing.T) {
	tasks := []curd.RepoTask{
		{Name: "repo-a", Path: "/tmp/a", Status: TaskRunning},
		{Name: "repo-b", Path: "/tmp/b", Status: TaskRunning},
	}

	m := NewProgressModel(tasks)
	if len(m.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(m.Tasks))
	}

	v := m.View()
	if v == "" {
		t.Error("View() returned empty string")
	}
}

func TestProgressModel_Init(t *testing.T) {
	tasks := []curd.RepoTask{
		{Name: "repo-a", Status: TaskRunning},
	}
	m := NewProgressModel(tasks)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil cmd")
	}
}

func TestRepoTaskStatusConstants(t *testing.T) {
	if TaskPending != curd.TaskPending {
		t.Errorf("TaskPending mismatch")
	}
	if TaskRunning != curd.TaskRunning {
		t.Errorf("TaskRunning mismatch")
	}
	if TaskDone != curd.TaskDone {
		t.Errorf("TaskDone mismatch")
	}
	if TaskFailed != curd.TaskFailed {
		t.Errorf("TaskFailed mismatch")
	}
}

func TestProgressModel_Update(t *testing.T) {
	tasks := []curd.RepoTask{
		{Name: "repo-a", Status: TaskRunning},
	}
	m := NewProgressModel(tasks)

	updateMsg := RepoTaskUpdateMsg{
		Index:  0,
		Status: TaskDone,
		Result: "success",
	}

	m, _ = m.Update(updateMsg)
	if m.Tasks[0].Status != TaskDone {
		t.Errorf("task status = %d, want TaskDone (%d)", m.Tasks[0].Status, TaskDone)
	}
}
