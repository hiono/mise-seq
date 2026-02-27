package mise

import (
	"testing"

	"github.com/mise-seq/config-loader/config"
)

func TestFilterHooksByWhen(t *testing.T) {
	hooks := []config.Hook{
		{Run: "always", When: []config.When{config.WhenAlways}},
		{Run: "install", When: []config.When{config.WhenInstall}},
		{Run: "update", When: []config.When{config.WhenUpdate}},
		{Run: "empty", When: []config.When{}}, // empty = always runs
		{Run: "multi", When: []config.When{config.WhenInstall, config.WhenUpdate}},
	}

	tests := []struct {
		name     string
		when     config.When
		expected int // count only, since order varies
	}{
		{
			name:     "filter install",
			when:     config.WhenInstall,
			expected: 4, // always + install + empty + multi
		},
		{
			name:     "filter update",
			when:     config.WhenUpdate,
			expected: 4, // always + update + empty + multi
		},
		{
			name:     "filter always",
			when:     config.WhenAlways,
			expected: 2, // always + empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterHooksByWhen(hooks, tt.when)
			if len(result) != tt.expected {
				t.Errorf("Expected %d hooks, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestExtractHookScripts(t *testing.T) {
	hooks := []config.Hook{
		{Run: "echo hello"},
		{Run: "  echo whitespace  "},
		{Run: ""},
		{Run: "echo world"},
	}

	scripts := ExtractHookScripts(hooks)

	expected := []string{"echo hello", "echo whitespace", "echo world"}
	if len(scripts) != len(expected) {
		t.Fatalf("Expected %d scripts, got %d", len(expected), len(scripts))
	}

	for i, e := range expected {
		if scripts[i] != e {
			t.Errorf("Expected script %d to be '%s', got '%s'", i, e, scripts[i])
		}
	}
}

func TestExtractHookScripts_Empty(t *testing.T) {
	scripts := ExtractHookScripts(nil)
	if len(scripts) != 0 {
		t.Errorf("Expected 0 scripts for nil input, got %d", len(scripts))
	}

	scripts = ExtractHookScripts([]config.Hook{})
	if len(scripts) != 0 {
		t.Errorf("Expected 0 scripts for empty input, got %d", len(scripts))
	}
}

func TestGetToolHooks(t *testing.T) {
	cfg := &config.Config{
		Tools: map[string]config.Tool{
			"tool1": {
				Preinstall:  []config.Hook{{Run: "tool1-preinstall"}},
				Postinstall: []config.Hook{{Run: "tool1-postinstall"}},
			},
			"tool2": {
				Preinstall: []config.Hook{{Run: "tool2-preinstall"}},
			},
		},
	}

	preinstall, postinstall := GetToolHooks(cfg, "tool1")
	if len(preinstall) != 1 || preinstall[0].Run != "tool1-preinstall" {
		t.Errorf("Expected tool1 preinstall")
	}
	if len(postinstall) != 1 || postinstall[0].Run != "tool1-postinstall" {
		t.Errorf("Expected tool1 postinstall")
	}

	// Non-existent tool
	preinstall, postinstall = GetToolHooks(cfg, "nonexistent")
	if preinstall != nil || postinstall != nil {
		t.Error("Expected nil for non-existent tool")
	}

	// Nil config
	preinstall, postinstall = GetToolHooks(nil, "tool1")
	if preinstall != nil || postinstall != nil {
		t.Error("Expected nil for nil config")
	}
}
