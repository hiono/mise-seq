package config

import (
	"testing"
)

func TestLogger_Levels(t *testing.T) {
	// Test with debug=false (INFO level)
	logger := NewLogger(false)

	// These should not panic
	logger.LogInfo("info message %s", "test")
	logger.LogWarn("warn message %s", "test")
	logger.LogError("error message %s", "test")

	// Debug should be suppressed
	logger.LogDebug("debug message - should not appear")
}

func TestLogger_DebugMode(t *testing.T) {
	// Test with debug=true (DEBUG level)
	logger := NewLogger(true)

	// All levels should work
	logger.LogInfo("info message")
	logger.LogWarn("warn message")
	logger.LogError("error message")
	logger.LogDebug("debug message")
}

func TestLogger_Global(t *testing.T) {
	// Save original logger
	orig := defaultLogger
	defer func() { defaultLogger = orig }()

	// Initialize with debug
	InitLogger(true)

	// These should not panic
	Info("info test")
	Warn("warn test")
	Error("error test")
	Debug("debug test")
}

func TestLogger_DefaultLogger(t *testing.T) {
	// Save original logger
	orig := defaultLogger
	defer func() { defaultLogger = orig }()

	// Reset to nil
	defaultLogger = nil

	// GetLogger should create a default logger
	logger := GetLogger()
	if logger == nil {
		t.Error("Expected non-nil logger")
	}

	// Should not panic
	logger.LogInfo("test")
}
