package screens

import "testing"

func TestNewMenuModel(t *testing.T) {
	m := NewMenuModel()
	v := m.View()
	if v == "" {
		t.Error("NewMenuModel().View() returned empty string")
	}
}

func TestNewStatusModel(t *testing.T) {
	m := NewStatusModel()
	if m.step != statusStepProgress {
		t.Errorf("initial step = %d, want statusStepProgress (%d)", m.step, statusStepProgress)
	}
	v := m.View()
	if v == "" {
		t.Error("NewStatusModel().View() returned empty string")
	}
}

func TestNewCommitModel(t *testing.T) {
	m := NewCommitModel(nil)
	if m.step != commitStepMessage {
		t.Errorf("initial step = %d, want commitStepMessage (%d)", m.step, commitStepMessage)
	}
	v := m.View()
	if v == "" {
		t.Error("NewCommitModel().View() returned empty string")
	}
}

func TestNewCommitModel_WithHistory(t *testing.T) {
	msgs := []string{"fix: bug", "feat: new stuff"}
	m := NewCommitModel(msgs)
	if len(m.recentMessages) != 2 {
		t.Errorf("recentMessages count = %d, want 2", len(m.recentMessages))
	}
	v := m.View()
	if v == "" {
		t.Error("NewCommitModel(with history).View() returned empty string")
	}
}

func TestNewCleanupModel(t *testing.T) {
	m := NewCleanupModel()
	if m.step != cleanupStepDrop {
		t.Errorf("initial step = %d, want cleanupStepDrop (%d)", m.step, cleanupStepDrop)
	}
	if m.dropCursor != 0 {
		t.Errorf("initial dropCursor = %d, want 0", m.dropCursor)
	}
	v := m.View()
	if v == "" {
		t.Error("NewCleanupModel().View() returned empty string")
	}
}

func TestNewBranchesModel(t *testing.T) {
	m := NewBranchesModel()
	if m.step != branchesStepProgress {
		t.Errorf("initial step = %d, want branchesStepProgress (%d)", m.step, branchesStepProgress)
	}
	v := m.View()
	if v == "" {
		t.Error("NewBranchesModel().View() returned empty string")
	}
}

func TestNewPullRequestModel(t *testing.T) {
	m := NewPullRequestModel(nil)
	if m.step != prStepMessage {
		t.Errorf("initial step = %d, want prStepMessage (%d)", m.step, prStepMessage)
	}
	v := m.View()
	if v == "" {
		t.Error("NewPullRequestModel().View() returned empty string")
	}
}

func TestNewCloneModel(t *testing.T) {
	m := NewCloneModel()
	if m.step != cloneStepInput {
		t.Errorf("initial step = %d, want cloneStepInput (%d)", m.step, cloneStepInput)
	}
	v := m.View()
	if v == "" {
		t.Error("NewCloneModel().View() returned empty string")
	}
}

func TestNewAutomergeModel(t *testing.T) {
	m := NewAutomergeModel()
	if m.step != automergeStepTarget {
		t.Errorf("initial step = %d, want automergeStepTarget (%d)", m.step, automergeStepTarget)
	}
	v := m.View()
	if v == "" {
		t.Error("NewAutomergeModel().View() returned empty string")
	}
}

func TestNewMergePRsModel(t *testing.T) {
	m := NewMergePRsModel()
	if m.step != mergePRsStepInput {
		t.Errorf("initial step = %d, want mergePRsStepInput (%d)", m.step, mergePRsStepInput)
	}
	v := m.View()
	if v == "" {
		t.Error("NewMergePRsModel().View() returned empty string")
	}
}

func TestNewEnableWorkflowsModel(t *testing.T) {
	m := NewEnableWorkflowsModel()
	if m.step != enableWFStepInput {
		t.Errorf("initial step = %d, want enableWFStepInput (%d)", m.step, enableWFStepInput)
	}
	v := m.View()
	if v == "" {
		t.Error("NewEnableWorkflowsModel().View() returned empty string")
	}
}

func TestNewTeamPRsModel(t *testing.T) {
	m := NewTeamPRsModel()
	v := m.View()
	if v == "" {
		t.Error("NewTeamPRsModel().View() returned empty string")
	}
}

func TestNewMyPRsModel(t *testing.T) {
	m := NewMyPRsModel()
	if m.step != myPRsStepFetching {
		t.Errorf("initial step = %d, want myPRsStepFetching (%d)", m.step, myPRsStepFetching)
	}
	v := m.View()
	if v == "" {
		t.Error("NewMyPRsModel().View() returned empty string")
	}
}
