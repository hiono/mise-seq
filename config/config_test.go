package config

import (
	"testing"
)

func TestConfig_MergeDefaults(t *testing.T) {
	cfg := &Config{
		Defaults: &Defaults{
			Preinstall: []Hook{
				{Run: "echo default-preinstall"},
			},
			Postinstall: []Hook{
				{Run: "echo default-postinstall"},
			},
		},
		Tools: map[string]Tool{
			"tool1": {
				Version: "1.0.0",
			},
			"tool2": {
				Version:     "2.0.0",
				Preinstall: []Hook{{Run: "echo tool2-preinstall"}},
			},
		},
	}

	cfg.MergeDefaults()

	// tool1 should get default hooks
	if len(cfg.Tools["tool1"].Preinstall) != 1 {
		t.Errorf("Expected tool1 to have 1 preinstall hook, got %d", len(cfg.Tools["tool1"].Preinstall))
	}
	if cfg.Tools["tool1"].Preinstall[0].Run != "echo default-preinstall" {
		t.Errorf("Expected tool1 preinstall hook to be default, got %s", cfg.Tools["tool1"].Preinstall[0].Run)
	}

	// tool2 should keep its own hook (not overwritten)
	if len(cfg.Tools["tool2"].Preinstall) != 1 {
		t.Errorf("Expected tool2 to have 1 preinstall hook, got %d", len(cfg.Tools["tool2"].Preinstall))
	}
	if cfg.Tools["tool2"].Preinstall[0].Run != "echo tool2-preinstall" {
		t.Errorf("Expected tool2 preinstall hook to be tool-specific, got %s", cfg.Tools["tool2"].Preinstall[0].Run)
	}
}

func TestHasDefaults(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected bool
	}{
		{
			name:     "nil config",
			cfg:      nil,
			expected: false,
		},
		{
			name:     "nil defaults",
			cfg:      &Config{Defaults: nil},
			expected: false,
		},
		{
			name: "empty defaults",
			cfg: &Config{
				Defaults: &Defaults{
					Preinstall: nil,
					Postinstall: nil,
				},
			},
			expected: false,
		},
		{
			name: "with preinstall",
			cfg: &Config{
				Defaults: &Defaults{
					Preinstall: []Hook{{Run: "echo test"}},
				},
			},
			expected: true,
		},
		{
			name: "with postinstall",
			cfg: &Config{
				Defaults: &Defaults{
					Postinstall: []Hook{{Run: "echo test"}},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasDefaults(tt.cfg)
			if result != tt.expected {
				t.Errorf("HasDefaults() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetDefaultsHooks(t *testing.T) {
	cfg := &Config{
		Defaults: &Defaults{
			Preinstall: []Hook{
				{Run: "echo preinstall", Description: "preinstall hook"},
			},
			Postinstall: []Hook{
				{Run: "echo postinstall", Description: "postinstall hook"},
			},
		},
	}

	preinstall, postinstall := GetDefaultsHooks(cfg)

	if len(preinstall) != 1 {
		t.Fatalf("Expected 1 preinstall hook, got %d", len(preinstall))
	}
	if preinstall[0].Run != "echo preinstall" {
		t.Errorf("Expected preinstall run 'echo preinstall', got '%s'", preinstall[0].Run)
	}

	if len(postinstall) != 1 {
		t.Fatalf("Expected 1 postinstall hook, got %d", len(postinstall))
	}
	if postinstall[0].Run != "echo postinstall" {
		t.Errorf("Expected postinstall run 'echo postinstall', got '%s'", postinstall[0].Run)
	}
}

func TestGetDefaultsHooks_Nil(t *testing.T) {
	preinstall, postinstall := GetDefaultsHooks(nil)
	if preinstall != nil || postinstall != nil {
		t.Error("Expected nil hooks for nil config")
	}

	cfg := &Config{}
	preinstall, postinstall = GetDefaultsHooks(cfg)
	if preinstall != nil || postinstall != nil {
		t.Error("Expected nil hooks for empty config")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Config
		expectErr bool
		errMsg    string
	}{
		{
			name:      "nil config",
			cfg:       nil,
			expectErr: false,
		},
		{
			name:      "empty config",
			cfg:       &Config{},
			expectErr: false,
		},
		{
			name: "valid config with tools_order",
			cfg: &Config{
				Tools: map[string]Tool{
					"jq":   {Version: "latest"},
					"node": {Version: "20"},
				},
				ToolsOrder: []string{"jq", "node"},
			},
			expectErr: false,
		},
		{
			name: "tools_order not in tools",
			cfg: &Config{
				Tools: map[string]Tool{
					"jq": {Version: "latest"},
				},
				ToolsOrder: []string{"jq", "node"},
			},
			expectErr: true,
			errMsg:    "tool_order contains 'node' which is not in tools",
		},
		{
			name: "empty tools_order with tools",
			cfg: &Config{
				Tools:      map[string]Tool{"jq": {Version: "latest"}},
				ToolsOrder: []string{},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.cfg)
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}
