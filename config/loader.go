package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Loader handles unified config loading
type Loader struct {
	explicitFormat string
}

// NewLoader creates a new config loader
func NewLoader() *Loader {
	return &Loader{}
}

// SetFormat explicitly sets the config format
func (l *Loader) SetFormat(format string) {
	l.explicitFormat = format
}

// Parse loads a config file and returns a Config
func (l *Loader) Parse(path string) (*Config, error) {
	// Determine format
	format := l.explicitFormat
	if format == "" {
		var err error
		format, err = DetectFormat(path)
		if err != nil {
			return nil, err
		}
	}

	// Check file exists
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("config file not found: %w", err)
	}

	// Parse based on format
	switch format {
	case "json":
		return ParseJSON(path)
	case "yaml", "yml":
		return ParseYAML(path)
	case "toml":
		return ParseTOML(path)
	case "cue":
		return ParseCUE(path)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// ParseFile is a convenience function to parse a config file
func ParseFile(path string) (*Config, error) {
	loader := NewLoader()
	return loader.Parse(path)
}

// ParseFileWithFormat parses a config file with explicit format
func ParseFileWithFormat(path, format string) (*Config, error) {
	loader := NewLoader()
	loader.SetFormat(format)
	return loader.Parse(path)
}

// GetToolOrder returns the tools_order from a config
func GetToolOrder(cfg *Config) []string {
	if cfg == nil {
		return nil
	}
	return cfg.ToolsOrder
}

// GetTools returns the tools map from a config
func GetTools(cfg *Config) map[string]Tool {
	if cfg == nil {
		return nil
	}
	return cfg.Tools
}

// GetDefaults returns the defaults from a config
func GetDefaults(cfg *Config) *Defaults {
	if cfg == nil {
		return nil
	}
	return cfg.Defaults
}

// GetSettings returns the settings from a config
func GetSettings(cfg *Config) *Settings {
	if cfg == nil {
		return nil
	}
	return cfg.Settings
}

// NormalizePath converts a path to absolute and resolves relative references
func NormalizePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}
	return absPath, nil
}
