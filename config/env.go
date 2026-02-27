package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds runtime configuration from environment variables
type RuntimeConfig struct {
	// DryRun mode - don't actually execute commands
	DryRun bool

	// Debug mode - verbose output
	Debug bool

	// ForceHooks forces hook execution even if unchanged
	ForceHooks bool

	// RunPostinstallOnUpdate runs postinstall hooks on update
	RunPostinstallOnUpdate bool

	// StateDir - custom state directory (default: $XDG_CACHE_HOME/tools/state/)
	StateDir string

	// CUEVersion - version of cue to bootstrap
	CUEVersion string

	// MiseShimsDefault - path to mise shims
	MiseShimsDefault string

	// MiseDataDir - mise data directory
	MiseDataDir string

	// MiseShimsCustom - custom shims directory
	MiseShimsCustom string
}

// LoadRuntimeConfig loads configuration from environment variables
func LoadRuntimeConfig() *RuntimeConfig {
	cfg := &RuntimeConfig{
		// Default CUE version
		CUEVersion: os.Getenv("CUE_VERSION"),
	}

	// Dry run mode
	if os.Getenv("DRY_RUN") != "" {
		cfg.DryRun = true
	}

	// Debug mode
	if os.Getenv("DEBUG") != "" {
		cfg.Debug = true
	}

	// Force hooks
	if os.Getenv("FORCE_HOOKS") != "" {
		cfg.ForceHooks = true
	}

	// Run postinstall on update
	if os.Getenv("RUN_POSTINSTALL_ON_UPDATE") != "" {
		cfg.RunPostinstallOnUpdate = true
	}

	// Custom state directory
	if stateDir := os.Getenv("STATE_DIR"); stateDir != "" {
		cfg.StateDir = stateDir
	}

	// Mise paths
	cfg.MiseShimsDefault = os.Getenv("MISE_SHIMS_DEFAULT")
	if cfg.MiseShimsDefault == "" {
		// Try common locations
		home := os.Getenv("HOME")
		if home != "" {
			cfg.MiseShimsDefault = filepath.Join(home, ".local", "share", "mise", "shims")
		}
	}

	cfg.MiseDataDir = os.Getenv("MISE_DATA_DIR")
	if cfg.MiseDataDir == "" {
		home := os.Getenv("HOME")
		if home != "" {
			cfg.MiseDataDir = filepath.Join(home, ".local", "share", "mise")
		}
	}

	cfg.MiseShimsCustom = os.Getenv("MISE_SHIMS_CUSTOM")

	return cfg
}

// SetupEnvironment prepares the environment for mise
func (c *RuntimeConfig) SetupEnvironment() error {
	// Build PATH with mise shims
	pathEntries := []string{}

	// Custom shims first (highest priority)
	if c.MiseShimsCustom != "" {
		if err := validatePath(c.MiseShimsCustom); err != nil {
			return fmt.Errorf("MISE_SHIMS_CUSTOM invalid: %w", err)
		}
		pathEntries = append(pathEntries, c.MiseShimsCustom)
	}

	// Default mise shims
	if c.MiseShimsDefault != "" {
		if err := validatePath(c.MiseShimsDefault); err != nil {
			return fmt.Errorf("MISE_SHIMS_DEFAULT invalid: %w", err)
		}
		pathEntries = append(pathEntries, c.MiseShimsDefault)
	}

	// Add current PATH
	currentPath := os.Getenv("PATH")
	if currentPath != "" {
		pathEntries = append(pathEntries, currentPath)
	}

	// Set new PATH
	newPath := joinPaths(pathEntries...)
	if err := os.Setenv("PATH", newPath); err != nil {
		return fmt.Errorf("failed to set PATH: %w", err)
	}

	return nil
}

// validatePath checks if a path exists and is a directory
func validatePath(p string) error {
	info, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Path doesn't exist yet, that's OK
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", p)
	}
	return nil
}

// joinPaths joins multiple path entries with os.PathListSeparator
func joinPaths(paths ...string) string {
	result := ""
	for _, p := range paths {
		if p == "" {
			continue
		}
		if result == "" {
			result = p
		} else {
			result = result + string(os.PathListSeparator) + p
		}
	}
	return result
}
