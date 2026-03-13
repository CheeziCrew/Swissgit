package ops

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.TargetBranch != "main" {
		t.Errorf("TargetBranch = %q, want main", cfg.TargetBranch)
	}
	if len(cfg.ChangeTypes) == 0 {
		t.Error("ChangeTypes should not be empty")
	}
	if cfg.TemplatePath != "" {
		t.Errorf("TemplatePath = %q, want empty", cfg.TemplatePath)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	cfg := LoadConfig()
	if cfg.TargetBranch != "main" {
		t.Errorf("TargetBranch = %q, want main", cfg.TargetBranch)
	}
	if len(cfg.ChangeTypes) != len(defaultChangeTypes) {
		t.Errorf("ChangeTypes len = %d, want %d", len(cfg.ChangeTypes), len(defaultChangeTypes))
	}
}

func TestLoadConfig_CustomFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	sgDir := filepath.Join(dir, ".swissgit")
	os.MkdirAll(sgDir, 0o755)

	cfg := Config{
		ChangeTypes:  []string{"Feature", "Bugfix"},
		TargetBranch: "develop",
		TemplatePath: "/tmp/custom.md",
	}
	data, _ := json.Marshal(cfg)
	os.WriteFile(filepath.Join(sgDir, "config.json"), data, 0o644)

	loaded := LoadConfig()
	if loaded.TargetBranch != "develop" {
		t.Errorf("TargetBranch = %q, want develop", loaded.TargetBranch)
	}
	if len(loaded.ChangeTypes) != 2 {
		t.Errorf("ChangeTypes len = %d, want 2", len(loaded.ChangeTypes))
	}
	if loaded.TemplatePath != "/tmp/custom.md" {
		t.Errorf("TemplatePath = %q", loaded.TemplatePath)
	}
}

func TestLoadConfig_PartialOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	sgDir := filepath.Join(dir, ".swissgit")
	os.MkdirAll(sgDir, 0o755)

	// Only override target branch, keep default change types
	data := []byte(`{"targetBranch": "master"}`)
	os.WriteFile(filepath.Join(sgDir, "config.json"), data, 0o644)

	loaded := LoadConfig()
	if loaded.TargetBranch != "master" {
		t.Errorf("TargetBranch = %q, want master", loaded.TargetBranch)
	}
	if len(loaded.ChangeTypes) != len(defaultChangeTypes) {
		t.Errorf("ChangeTypes should fall back to defaults, got %d", len(loaded.ChangeTypes))
	}
}
