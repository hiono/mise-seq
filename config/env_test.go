package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRuntimeConfig_Defaults(t *testing.T) {
	// Clear all env vars
	os.Unsetenv("DRY_RUN")
	os.Unsetenv("DEBUG")
	os.Unsetenv("FORCE_HOOKS")
	os.Unsetenv("RUN_POSTINSTALL_ON_UPDATE")
	os.Unsetenv("STATE_DIR")
	os.Unsetenv("CUE_VERSION")
	os.Unsetenv("MISE_SHIMS_DEFAULT")
	os.Unsetenv("MISE_DATA_DIR")
	os.Unsetenv("MISE_SHIMS_CUSTOM")

	cfg := LoadRuntimeConfig()

	if cfg.DryRun {
		t.Error("Expected DryRun=false by default")
	}
	if cfg.Debug {
		t.Error("Expected Debug=false by default")
	}
	if cfg.ForceHooks {
		t.Error("Expected ForceHooks=false by default")
	}
	if cfg.RunPostinstallOnUpdate {
		t.Error("Expected RunPostinstallOnUpdate=false by default")
	}
	if cfg.CUEVersion != "" {
		t.Error("Expected empty CUEVersion by default")
	}
}

func TestLoadRuntimeConfig_EnvVars(t *testing.T) {
	// Set env vars
	os.Setenv("DRY_RUN", "1")
	os.Setenv("DEBUG", "1")
	os.Setenv("FORCE_HOOKS", "1")
	os.Setenv("RUN_POSTINSTALL_ON_UPDATE", "1")
	os.Setenv("STATE_DIR", "/custom/state")
	os.Setenv("CUE_VERSION", "v0.9.0")
	os.Setenv("MISE_SHIMS_CUSTOM", "/custom/shims")
	defer func() {
		// Clean up
		os.Unsetenv("DRY_RUN")
		os.Unsetenv("DEBUG")
		os.Unsetenv("FORCE_HOOKS")
		os.Unsetenv("RUN_POSTINSTALL_ON_UPDATE")
		os.Unsetenv("STATE_DIR")
		os.Unsetenv("CUE_VERSION")
		os.Unsetenv("MISE_SHIMS_CUSTOM")
	}()

	cfg := LoadRuntimeConfig()

	if !cfg.DryRun {
		t.Error("Expected DryRun=true")
	}
	if !cfg.Debug {
		t.Error("Expected Debug=true")
	}
	if !cfg.ForceHooks {
		t.Error("Expected ForceHooks=true")
	}
	if !cfg.RunPostinstallOnUpdate {
		t.Error("Expected RunPostinstallOnUpdate=true")
	}
	if cfg.StateDir != "/custom/state" {
		t.Errorf("Expected StateDir=/custom/state, got %s", cfg.StateDir)
	}
	if cfg.CUEVersion != "v0.9.0" {
		t.Errorf("Expected CUEVersion=v0.9.0, got %s", cfg.CUEVersion)
	}
	if cfg.MiseShimsCustom != "/custom/shims" {
		t.Errorf("Expected MiseShimsCustom=/custom/shims, got %s", cfg.MiseShimsCustom)
	}
}

func TestJoinPaths(t *testing.T) {
	tests := []struct {
		paths    []string
		expected string
	}{
		{[]string{}, ""},
		{[]string{"/path1"}, "/path1"},
		{[]string{"/path1", "/path2"}, "/path1" + string(os.PathListSeparator) + "/path2"},
		{[]string{"", "/path2"}, "/path2"},
		{[]string{"/path1", ""}, "/path1"},
		{[]string{"", ""}, ""},
	}

	for _, tt := range tests {
		result := joinPaths(tt.paths...)
		if result != tt.expected {
			t.Errorf("joinPaths(%v) = %s, expected %s", tt.paths, result, tt.expected)
		}
	}
}

func TestRuntimeConfig_SetupEnvironment(t *testing.T) {
	// Save original PATH
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)

	tmpDir := t.TempDir()
	cfg := &RuntimeConfig{
		MiseShimsCustom: tmpDir,
	}

	err := cfg.SetupEnvironment()
	if err != nil {
		t.Fatalf("SetupEnvironment failed: %v", err)
	}

	newPath := os.Getenv("PATH")
	expected := tmpDir + string(os.PathListSeparator) + origPath
	if newPath != expected {
		t.Errorf("PATH = %s, expected %s", newPath, expected)
	}
}

func TestValidatePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Test existing directory
	if err := validatePath(tmpDir); err != nil {
		t.Errorf("validatePath(%s) failed: %v", tmpDir, err)
	}

	// Test non-existent path (should be OK)
	nonExistent := filepath.Join(tmpDir, "nonexistent")
	if err := validatePath(nonExistent); err != nil {
		t.Errorf("validatePath(%s) failed: %v", nonExistent, err)
	}

	// Test file (not directory)
	tmpFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if err := validatePath(tmpFile); err == nil {
		t.Error("Expected error for non-directory path")
	}
}
