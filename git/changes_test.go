package git

import (
	"testing"

	gogit "github.com/go-git/go-git/v5"
)

func TestChanges_HasChanges(t *testing.T) {
	tests := []struct {
		name string
		c    Changes
		want bool
	}{
		{"zero value", Changes{}, false},
		{"modified only", Changes{Modified: 1}, true},
		{"added only", Changes{Added: 2}, true},
		{"deleted only", Changes{Deleted: 1}, true},
		{"untracked only", Changes{Untracked: 3}, true},
		{"all types", Changes{Modified: 1, Added: 2, Deleted: 3, Untracked: 4}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.HasChanges(); got != tt.want {
				t.Errorf("HasChanges() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChanges_Total(t *testing.T) {
	tests := []struct {
		name string
		c    Changes
		want int
	}{
		{"zero", Changes{}, 0},
		{"one of each", Changes{Modified: 1, Added: 1, Deleted: 1, Untracked: 1}, 4},
		{"mixed", Changes{Modified: 3, Untracked: 7}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Total(); got != tt.want {
				t.Errorf("Total() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifyStatusLine(t *testing.T) {
	tests := []struct {
		name      string
		x, y      byte
		wantMod   int
		wantAdd   int
		wantDel   int
		wantUntr  int
	}{
		{"untracked ??", '?', '?', 0, 0, 0, 1},
		{"modified in worktree", ' ', 'M', 0, 0, 0, 0},   // x=' ', y='M' -> Modified++
		{"modified in index", 'M', ' ', 0, 0, 0, 0},       // x='M', y=' ' -> Modified++
		{"added in index", 'A', ' ', 0, 0, 0, 0},           // x='A' -> Added++
		{"deleted in worktree", ' ', 'D', 0, 0, 0, 0},     // x=' ', y='D' -> Deleted++
		{"deleted in index", 'D', ' ', 0, 0, 0, 0},         // x='D' -> Deleted++
		{"renamed in index", 'R', ' ', 0, 0, 0, 0},         // x='R' -> Added++
		{"copied in index", 'C', ' ', 0, 0, 0, 0},          // x='C' -> Added++
	}

	// The table above isn't fully testing counts because each case is isolated.
	// Let's use direct assertions instead.
	_ = tests

	directTests := []struct {
		name string
		x, y byte
		want Changes
	}{
		{"untracked", '?', '?', Changes{Untracked: 1}},
		{"modified worktree", ' ', 'M', Changes{Modified: 1}},
		{"modified index", 'M', ' ', Changes{Modified: 1}},
		{"both modified", 'M', 'M', Changes{Modified: 1}},
		{"added index", 'A', ' ', Changes{Added: 1}},
		{"added worktree", ' ', 'A', Changes{Added: 1}},
		{"deleted index", 'D', ' ', Changes{Deleted: 1}},
		{"deleted worktree", ' ', 'D', Changes{Deleted: 1}},
		{"renamed index", 'R', ' ', Changes{Added: 1}},
		{"renamed worktree", ' ', 'R', Changes{Added: 1}},
		{"copied index", 'C', ' ', Changes{Added: 1}},
		{"copied worktree", ' ', 'C', Changes{Added: 1}},
		{"no change", ' ', ' ', Changes{}},
		{"modified and deleted", 'M', 'D', Changes{Modified: 1, Deleted: 1}},
	}
	for _, tt := range directTests {
		t.Run(tt.name, func(t *testing.T) {
			var got Changes
			classifyStatusLine(tt.x, tt.y, &got)
			if got != tt.want {
				t.Errorf("classifyStatusLine(%q, %q) = %+v, want %+v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestCountChanges(t *testing.T) {
	tests := []struct {
		name   string
		status gogit.Status
		want   Changes
	}{
		{
			"empty status",
			gogit.Status{},
			Changes{},
		},
		{
			"untracked file",
			gogit.Status{
				"new.txt": &gogit.FileStatus{Worktree: gogit.Untracked, Staging: gogit.Untracked},
			},
			Changes{Untracked: 1},
		},
		{
			"modified in worktree",
			gogit.Status{
				"file.go": &gogit.FileStatus{Worktree: gogit.Modified, Staging: gogit.Unmodified},
			},
			Changes{Modified: 1},
		},
		{
			"modified in staging",
			gogit.Status{
				"file.go": &gogit.FileStatus{Worktree: gogit.Unmodified, Staging: gogit.Modified},
			},
			Changes{Modified: 1},
		},
		{
			"added in staging",
			gogit.Status{
				"new.go": &gogit.FileStatus{Worktree: gogit.Unmodified, Staging: gogit.Added},
			},
			Changes{Added: 1},
		},
		{
			"deleted in worktree",
			gogit.Status{
				"old.go": &gogit.FileStatus{Worktree: gogit.Deleted, Staging: gogit.Unmodified},
			},
			Changes{Deleted: 1},
		},
		{
			"multiple files",
			gogit.Status{
				"a.go": &gogit.FileStatus{Worktree: gogit.Modified, Staging: gogit.Unmodified},
				"b.go": &gogit.FileStatus{Worktree: gogit.Untracked, Staging: gogit.Untracked},
				"c.go": &gogit.FileStatus{Worktree: gogit.Unmodified, Staging: gogit.Deleted},
			},
			Changes{Modified: 1, Untracked: 1, Deleted: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountChanges(tt.status)
			if got != tt.want {
				t.Errorf("CountChanges() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
