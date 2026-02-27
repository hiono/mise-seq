package hooks

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// StateManager manages hook execution state with SHA256 markers
type StateManager struct {
	// StateDir is the directory to store state files
	// Default: $XDG_CACHE_HOME/tools/state/ or ~/.cache/tools/state/
	StateDir string

	// ForceHooks forces hook execution even if markers match
	ForceHooks bool

	// RunPostinstallOnUpdate runs postinstall hooks even if tool is already installed
	RunPostinstallOnUpdate bool
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	stateDir := os.Getenv("XDG_CACHE_HOME")
	if stateDir == "" {
		home := os.Getenv("HOME")
		if home != "" {
			stateDir = filepath.Join(home, ".cache")
		}
	}
	if stateDir == "" {
		stateDir = "/tmp"
	}
	stateDir = filepath.Join(stateDir, "tools", "state")

	return &StateManager{
		StateDir:               stateDir,
		ForceHooks:             false,
		RunPostinstallOnUpdate: false,
	}
}

// GetToolStateDir returns the state directory for a specific tool
func (s *StateManager) GetToolStateDir(toolName string) string {
	return filepath.Join(s.StateDir, toolName)
}

// computeSHA256 computes SHA256 hash of a string
func computeSHA256(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// ReadMarker reads a marker file and returns its content (SHA256 hash)
func (s *StateManager) ReadMarker(toolName, hookType string) (string, error) {
	markerPath := filepath.Join(s.GetToolStateDir(toolName), hookType+".sha256")
	data, err := os.ReadFile(markerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read marker: %w", err)
	}
	return string(data), nil
}

// WriteMarker writes a SHA256 hash to a marker file
func (s *StateManager) WriteMarker(toolName, hookType, hash string) error {
	toolDir := s.GetToolStateDir(toolName)
	if err := os.MkdirAll(toolDir, 0755); err != nil {
		return fmt.Errorf("failed to create tool state dir: %w", err)
	}

	markerPath := filepath.Join(toolDir, hookType+".sha256")
	if err := os.WriteFile(markerPath, []byte(hash), 0644); err != nil {
		return fmt.Errorf("failed to write marker: %w", err)
	}
	return nil
}

// ShouldRunHook determines if a hook should be executed
// Returns (shouldRun, existingHash, error)
func (s *StateManager) ShouldRunHook(toolName, hookType, script string) (bool, string, error) {
	currentHash := computeSHA256(script)

	// Check for existing marker
	existingHash, err := s.ReadMarker(toolName, hookType)
	if err != nil {
		return false, "", err
	}

	// If force hooks is enabled, always run
	if s.ForceHooks {
		return true, existingHash, nil
	}

	// If no existing marker, run the hook
	if existingHash == "" {
		return true, existingHash, nil
	}

	// If hashes don't match, run the hook
	if existingHash != currentHash {
		return true, existingHash, nil
	}

	// Hashes match - skip unless RunPostinstallOnUpdate is set for postinstall
	if hookType == "postinstall" && s.RunPostinstallOnUpdate {
		return true, existingHash, nil
	}

	// Skip - hashes match and no override
	return false, existingHash, nil
}

// SaveHookState saves the hook state after execution
func (s *StateManager) SaveHookState(toolName, hookType, script string) error {
	hash := computeSHA256(script)
	return s.WriteMarker(toolName, hookType, hash)
}

// ClearToolState removes all state files for a tool
func (s *StateManager) ClearToolState(toolName string) error {
	toolDir := s.GetToolStateDir(toolName)
	if err := os.RemoveAll(toolDir); err != nil {
		return fmt.Errorf("failed to clear tool state: %w", err)
	}
	return nil
}

// ClearAllState removes all state files
func (s *StateManager) ClearAllState() error {
	if err := os.RemoveAll(s.StateDir); err != nil {
		return fmt.Errorf("failed to clear all state: %w", err)
	}
	return nil
}
