package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseNestedShellExpansion(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	
	content := `{ "GCPProject": "${shell echo ${project}}" }`
	if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set the environment variable that we expect to be expanded
	expectedValue := "my-actual-project"
	os.Setenv("project", expectedValue)
	defer os.Unsetenv("project")

	var cfg struct {
		GCPProject string
	}
	
	err := Parse(context.Background(), configFile, &cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if cfg.GCPProject != expectedValue {
		t.Errorf("Expansion failed.\nExpected: %q\nActual:   %q", expectedValue, cfg.GCPProject)
	}
}

