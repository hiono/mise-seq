package config

import (
	"testing"
)

func TestParseToolRef(t *testing.T) {
	tests := []struct {
		input    string
		expected ToolRef
	}{
		{"jq@latest", ToolRef{Name: "jq", Version: "latest"}},
		{"go@1.22", ToolRef{Name: "go", Version: "1.22"}},
		{"rust@1.88", ToolRef{Name: "rust", Version: "1.88"}},
		{"nodejs", ToolRef{Name: "nodejs", Version: "latest"}},
		{"pnpm", ToolRef{Name: "pnpm", Version: "latest"}},
		{"cargo-binstall", ToolRef{Name: "cargo-binstall", Version: "latest"}},
	}

	for _, tt := range tests {
		result := ParseToolRef(tt.input)
		if result != tt.expected {
			t.Errorf("ParseToolRef(%s) = %+v, expected %+v", tt.input, result, tt.expected)
		}
	}
}

func TestToolResolver_NoDependencies(t *testing.T) {
	tools := map[string]Tool{
		"jq":   {Version: "latest"},
		"node": {Version: "20"},
	}

	resolver := NewToolResolver(tools)
	order, err := resolver.ResolveOrder()

	if err != nil {
		t.Fatalf("ResolveOrder failed: %v", err)
	}

	if len(order) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(order))
	}
}

func TestToolResolver_SimpleDependency(t *testing.T) {
	tools := map[string]Tool{
		"nodejs": {Version: "20"},
		"pnpm": {
			Version:  "latest",
			Depends:  []string{"nodejs@20"},
		},
	}

	resolver := NewToolResolver(tools)
	order, err := resolver.ResolveOrder()

	if err != nil {
		t.Fatalf("ResolveOrder failed: %v", err)
	}

	// nodejs should come before pnpm
	if len(order) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(order))
	}

	nodeIdx := -1
	pnpmIdx := -1
	for i, name := range order {
		if name == "nodejs" {
			nodeIdx = i
		}
		if name == "pnpm" {
			pnpmIdx = i
		}
	}

	if nodeIdx == -1 || pnpmIdx == -1 {
		t.Error("Both tools should be in order")
	}

	if nodeIdx >= pnpmIdx {
		t.Error("nodejs should be installed before pnpm")
	}
}

func TestToolResolver_ChainDependency(t *testing.T) {
	tools := map[string]Tool{
		"gcc":   {Version: "latest"},
		"go":    {Version: "1.22", Depends: []string{"gcc@latest"}},
		"rust":  {Version: "1.88", Depends: []string{"gcc@latest"}},
		"cargo": {Version: "latest", Depends: []string{"rust@1.88"}},
	}

	resolver := NewToolResolver(tools)
	order, err := resolver.ResolveOrder()

	if err != nil {
		t.Fatalf("ResolveOrder failed: %v", err)
	}

	// gcc -> go,rust -> cargo
	if len(order) != 4 {
		t.Errorf("Expected 4 tools, got %d", len(order))
	}

	// Check order
	gccIdx := -1
	goIdx := -1
	rustIdx := -1
	cargoIdx := -1
	for i, name := range order {
		switch name {
		case "gcc":
			gccIdx = i
		case "go":
			goIdx = i
		case "rust":
			rustIdx = i
		case "cargo":
			cargoIdx = i
		}
	}

	if gccIdx >= goIdx || gccIdx >= rustIdx {
		t.Error("gcc should come before go and rust")
	}

	if rustIdx >= cargoIdx {
		t.Error("rust should come before cargo")
	}
}

func TestToolResolver_CycleDetection(t *testing.T) {
	tools := map[string]Tool{
		"a": {Depends: []string{"b"}},
		"b": {Depends: []string{"c"}},
		"c": {Depends: []string{"a"}}, // cycle!
	}

	resolver := NewToolResolver(tools)
	_, err := resolver.ResolveOrder()

	if err == nil {
		t.Error("Expected error for cycle, got nil")
	}
}

func TestToolResolver_UnknownDependency(t *testing.T) {
	tools := map[string]Tool{
		"jq": {Depends: []string{"unknown@latest"}},
	}

	resolver := NewToolResolver(tools)
	_, err := resolver.ResolveOrder()

	if err == nil {
		t.Error("Expected error for unknown dependency, got nil")
	}
}

func TestValidateDependencies(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Config
		expectErr bool
	}{
		{
			name:      "nil config",
			cfg:       nil,
			expectErr: false,
		},
		{
			name: "valid dependencies",
			cfg: &Config{
				Tools: map[string]Tool{
					"nodejs": {Version: "20"},
					"pnpm":   {Depends: []string{"nodejs@20"}},
				},
			},
			expectErr: false,
		},
		{
			name: "unknown dependency",
			cfg: &Config{
				Tools: map[string]Tool{
					"pnpm": {Depends: []string{"nodejs@20"}},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDependencies(tt.cfg)
			if tt.expectErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestGetToolWithVersion(t *testing.T) {
	tests := []struct {
		name    string
		tool    Tool
		expected string
	}{
		{"jq", Tool{Version: "latest"}, "jq@latest"},
		{"go", Tool{Version: "1.22"}, "go@1.22"},
		{"rust", Tool{Version: "1.88"}, "rust@1.88"},
		{"node", Tool{Version: ""}, "node@latest"}, // empty = latest
	}

	for _, tt := range tests {
		result := GetToolWithVersion(tt.name, tt.tool)
		if result != tt.expected {
			t.Errorf("GetToolWithVersion(%s, %+v) = %s, expected %s", tt.name, tt.tool, result, tt.expected)
		}
	}
}
