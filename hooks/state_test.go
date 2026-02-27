package hooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStateManager_NewStateManager(t *testing.T) {
	// Set custom XDG_CACHE_HOME for testing
	os.Setenv("XDG_CACHE_HOME", "/tmp/test-cache")
	defer os.Unsetenv("XDG_CACHE_HOME")

	mgr := NewStateManager()

	expected := "/tmp/test-cache/tools/state"
	if mgr.StateDir != expected {
		t.Errorf("Expected StateDir %s, got %s", expected, mgr.StateDir)
	}
}

func TestStateManager_GetToolStateDir(t *testing.T) {
	mgr := NewStateManager()
	mgr.StateDir = "/tmp/test-state"

	toolDir := mgr.GetToolStateDir("mytool")
	expected := "/tmp/test-state/mytool"

	if toolDir != expected {
		t.Errorf("Expected %s, got %s", expected, toolDir)
	}
}

func TestComputeSHA256(t *testing.T) {
	tests := []struct {
		input    string
	}{
		{"hello"},
		{""},
		{"test-script"},
		{"echo install tool"},
	}

	for _, tt := range tests {
		result := computeSHA256(tt.input)
		if result == "" {
			t.Errorf("computeSHA256(%s) returned empty string", tt.input)
		}
		// Verify it's a valid 64-char hex string
		if len(result) != 64 {
			t.Errorf("computeSHA256(%s) = %s (len=%d), expected 64-char hex", tt.input, result, len(result))
		}
	}
}

func TestStateManager_ShouldRunHook(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewStateManager()
	mgr.StateDir = tmpDir

	toolName := "testtool"
	script1 := "echo hello"
	script2 := "echo world"

	// First run - should run (no marker)
	shouldRun, existingHash, err := mgr.ShouldRunHook(toolName, "preinstall", script1)
	if err != nil {
		t.Fatalf("ShouldRunHook failed: %v", err)
	}
	if !shouldRun {
		t.Error("Expected shouldRun=true on first run")
	}
	if existingHash != "" {
		t.Errorf("Expected empty existingHash on first run, got %s", existingHash)
	}

	// Save state
	if err := mgr.SaveHookState(toolName, "preinstall", script1); err != nil {
		t.Fatalf("SaveHookState failed: %v", err)
	}

	// Same script - should NOT run (hash matches)
	shouldRun, existingHash, err = mgr.ShouldRunHook(toolName, "preinstall", script1)
	if err != nil {
		t.Fatalf("ShouldRunHook failed: %v", err)
	}
	if shouldRun {
		t.Error("Expected shouldRun=false when hash matches")
	}

	// Different script - should run (hash mismatch)
	shouldRun, existingHash, err = mgr.ShouldRunHook(toolName, "preinstall", script2)
	if err != nil {
		t.Fatalf("ShouldRunHook failed: %v", err)
	}
	if !shouldRun {
		t.Error("Expected shouldRun=true when hash differs")
	}

	// Force hooks - should run
	mgr.ForceHooks = true
	shouldRun, _, err = mgr.ShouldRunHook(toolName, "preinstall", script1)
	if err != nil {
		t.Fatalf("ShouldRunHook failed: %v", err)
	}
	if !shouldRun {
		t.Error("Expected shouldRun=true when ForceHooks=true")
	}
}

func TestStateManager_ClearToolState(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewStateManager()
	mgr.StateDir = tmpDir

	toolName := "testtool"

	// Create state
	toolDir := mgr.GetToolStateDir(toolName)
	if err := os.MkdirAll(toolDir, 0755); err != nil {
		t.Fatalf("Failed to create tool dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "preinstall.sha256"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create marker: %v", err)
	}

	// Verify exists
	if _, err := os.Stat(toolDir); os.IsNotExist(err) {
		t.Error("Expected tool dir to exist before clear")
	}

	// Clear
	if err := mgr.ClearToolState(toolName); err != nil {
		t.Fatalf("ClearToolState failed: %v", err)
	}

	// Verify gone
	if _, err := os.Stat(toolDir); !os.IsNotExist(err) {
		t.Error("Expected tool dir to be removed after clear")
	}
}

func TestRunner_DryRun(t *testing.T) {
	runner := NewRunner(true) // dryRun = true

	result, err := runner.Run(nil, "testtool", HookTypePreinstall, "echo test")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	expectedStdout := "[dry-run] Would execute: echo test"
	if result.Stdout != expectedStdout {
		t.Errorf("Expected stdout '%s', got '%s'", expectedStdout, result.Stdout)
	}
}

func TestRunner_EmptyScript(t *testing.T) {
	runner := NewRunner(false)

	result, err := runner.Run(nil, "testtool", HookTypePreinstall, "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0 for empty script, got %d", result.ExitCode)
	}
}
