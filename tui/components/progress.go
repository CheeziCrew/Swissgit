package components

import (
	"github.com/CheeziCrew/curd"
)

// Re-export shared types.
type RepoTaskStatus = curd.RepoTaskStatus
type RepoTask = curd.RepoTask
type RepoTaskUpdateMsg = curd.RepoTaskUpdateMsg
type AllTasksDoneMsg = curd.AllTasksDoneMsg
type ProgressModel = curd.ProgressModel

const (
	TaskPending = curd.TaskPending
	TaskRunning = curd.TaskRunning
	TaskDone    = curd.TaskDone
	TaskFailed  = curd.TaskFailed
)

func NewProgressModel(tasks []curd.RepoTask) curd.ProgressModel {
	return curd.NewProgressModel(tasks, curd.SwissgitPalette)
}
