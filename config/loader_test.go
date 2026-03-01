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

	if len(cfg.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(cfg.Tools))
	}

	found := false
	for _, tool := range cfg.Tools {
		if tool.Name == "go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'go' tool in config")
	}
}

func TestParseYAML(t *testing.T) {
	cfg, err := ParseYAML("testdata/test.yaml")
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if len(cfg.Tools) != 3 {
		t.Errorf("Expected 3 tools in order, got %d", len(cfg.Tools))
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

	found := false
	for _, tool := range cfg.Tools {
		if tool.Name == "go" && tool.Version == "1.21" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected go version 1.21")
	}

	// Test YAML
	cfg, err = loader.Parse("testdata/test.yaml")
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	found = false
	for _, tool := range cfg.Tools {
		if tool.Name == "node" && tool.Version == "20" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected node version 20")
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
