package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseJSON(t *testing.T) {
	cfg, err := ParseJSON("testdata/test.json")
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(cfg.ToolsOrder) != 3 {
		t.Errorf("Expected 3 tools in order, got %d", len(cfg.ToolsOrder))
	}

	if _, ok := cfg.Tools["go"]; !ok {
		t.Error("Expected 'go' tool in config")
	}
}

func TestParseYAML(t *testing.T) {
	cfg, err := ParseYAML("testdata/test.yaml")
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if len(cfg.ToolsOrder) != 3 {
		t.Errorf("Expected 3 tools in order, got %d", len(cfg.ToolsOrder))
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"test.json", "json"},
		{"test.yaml", "yaml"},
		{"test.yml", "yaml"},
		{"test.toml", "toml"},
		{"test.cue", "cue"},
	}

	for _, tt := range tests {
		format, err := DetectFormat(tt.path)
		if err != nil {
			t.Errorf("DetectFormat(%s) failed: %v", tt.path, err)
		}
		if format != tt.expected {
			t.Errorf("DetectFormat(%s) = %s, expected %s", tt.path, format, tt.expected)
		}
	}
}

func TestLoader(t *testing.T) {
	loader := NewLoader()

	// Test JSON
	cfg, err := loader.Parse("testdata/test.json")
	if err != nil {
		t.Fatalf("Failed to load JSON: %v", err)
	}

	if cfg.Tools["go"].Version != "1.21" {
		t.Errorf("Expected go version 1.21, got %s", cfg.Tools["go"].Version)
	}

	// Test YAML
	cfg, err = loader.Parse("testdata/test.yaml")
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	if cfg.Tools["node"].Version != "20" {
		t.Errorf("Expected node version 20, got %s", cfg.Tools["node"].Version)
	}
}

func TestNormalizePath(t *testing.T) {
	// Create temp file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(tmpFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	path, err := NormalizePath(tmpFile)
	if err != nil {
		t.Fatalf("NormalizePath failed: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Errorf("Expected absolute path, got %s", path)
	}
}
