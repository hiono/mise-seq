package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// ParseJSON parses a JSON config file
func ParseJSON(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &cfg, nil
}

// ParseYAML parses a YAML config file
func ParseYAML(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &cfg, nil
}

// ParseTOML parses a TOML config file
func ParseTOML(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cfg Config
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	return &cfg, nil
}

// ParseCUE parses a CUE config file
// Note: CUE parsing requires full cue build - simplified for now
func ParseCUE(path string) (*Config, error) {
	// TODO: Implement CUE parsing with full cuelang.org/go dependency
	// For now, return unsupported error
	return nil, fmt.Errorf("CUE parsing not yet implemented - requires full CUE SDK")
}

// DetectFormat detects the config format from file extension
func DetectFormat(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return "json", nil
	case ".yaml", ".yml":
		return "yaml", nil
	case ".toml":
		return "toml", nil
	case ".cue":
		return "cue", nil
	default:
		return "", fmt.Errorf("unknown config format for file: %s", path)
	}
}
