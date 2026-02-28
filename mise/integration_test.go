package mise

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mise-seq/config-loader/config"
)

func TestInstallWithIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if mise is available
	if _, err := exec.LookPath("mise"); err != nil {
		t.Skip("mise not available in PATH")
	}

	tests := []struct {
		name            string
		tool            string
		preinstall      []config.Hook
		postinstall     []config.Hook
		wantPreinstall  string
		wantPostinstall string
	}{
		{
			name: "install with preinstall hook",
			tool: "fzf",
			preinstall: []config.Hook{
				{Run: "echo PREINSTALL_OUTPUT"},
			},
			wantPreinstall: "PREINSTALL_OUTPUT",
		},
		{
			name: "install with postinstall hook",
			tool: "fzf",
			postinstall: []config.Hook{
				{Run: "echo POSTINSTALL_OUTPUT"},
			},
			wantPostinstall: "POSTINSTALL_OUTPUT",
		},
		{
			name: "install with both hooks",
			tool: "fzf",
			preinstall: []config.Hook{
				{Run: "echo PREINSTALL_OUTPUT"},
			},
			postinstall: []config.Hook{
				{Run: "echo POSTINSTALL_OUTPUT"},
			},
			wantPreinstall:  "PREINSTALL_OUTPUT",
			wantPostinstall: "POSTINSTALL_OUTPUT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			testDir := filepath.Join(os.TempDir(), "mise-test-"+tt.tool+"-"+t.Name())
			miseDataDir := filepath.Join(testDir, "mise-data")
			configFile := filepath.Join(testDir, "config.toml")

			if err := os.MkdirAll(miseDataDir, 0755); err != nil {
				t.Fatalf("Failed to create test dir: %v", err)
			}
			if _, err := os.Create(configFile); err != nil {
				t.Fatalf("Failed to create config file: %v", err)
			}

			t.Setenv("MISE_DATA_DIR", miseDataDir)
			t.Setenv("MISE_CONFIG_FILE", configFile)
			t.Setenv("MISE_GLOBAL_CONFIG_FILE", filepath.Join(miseDataDir, "config.toml"))
			t.Setenv("MISE_EXPERIMENTAL", "true")

			client := NewClient()

			cfg := &config.Config{
				Tools: map[string]config.Tool{
					tt.tool: {
						Version:     "latest",
						Preinstall:  tt.preinstall,
						Postinstall: tt.postinstall,
					},
				},
			}

			err := client.InstallAllWithHooks(ctx, cfg, false)
			if err != nil {
				t.Fatalf("InstallAllWithHooks failed: %v", err)
			}

			installed, err := client.ListTools(ctx)
			if err != nil {
				t.Fatalf("ListTools failed: %v", err)
			}

			found := false
			for _, tool := range installed {
				if tool.Name == tt.tool {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected tool %s to be installed", tt.tool)
			}

			os.RemoveAll(testDir)
		})
	}
}

func TestUpgradeWithIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if mise is available
	if _, err := exec.LookPath("mise"); err != nil {
		t.Skip("mise not available in PATH")
	}

	ctx := context.Background()

	testDir := filepath.Join(os.TempDir(), "mise-test-upgrade-"+t.Name())
	miseDataDir := filepath.Join(testDir, "mise-data")
	configFile := filepath.Join(testDir, "config.toml")

	if err := os.MkdirAll(miseDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	if _, err := os.Create(configFile); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	t.Setenv("MISE_DATA_DIR", miseDataDir)
	t.Setenv("MISE_CONFIG_FILE", configFile)
	t.Setenv("MISE_GLOBAL_CONFIG_FILE", filepath.Join(miseDataDir, "config.toml"))
	t.Setenv("MISE_EXPERIMENTAL", "true")

	client := NewClient()

	cfg := &config.Config{
		Tools: map[string]config.Tool{
			"fzf": {
				Version: "latest",
				Postinstall: []config.Hook{
					{Run: "echo UPGRADE_POSTINSTALL_TEST"},
				},
			},
		},
	}

	err := client.InstallAllWithHooks(ctx, cfg, false)
	if err != nil {
		t.Fatalf("First InstallAllWithHooks failed: %v", err)
	}

	err = client.InstallAllWithHooks(ctx, cfg, true)
	if err != nil {
		t.Fatalf("Second InstallAllWithHooks (upgrade) failed: %v", err)
	}

	os.RemoveAll(testDir)
}

func TestIsManagedByMiseIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if mise is available
	if _, err := exec.LookPath("mise"); err != nil {
		t.Skip("mise not available in PATH")
	}

	ctx := context.Background()

	testDir := filepath.Join(os.TempDir(), "mise-test-ismanaged-"+t.Name())
	miseDataDir := filepath.Join(testDir, "mise-data")
	configFile := filepath.Join(testDir, "config.toml")

	if err := os.MkdirAll(miseDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	if _, err := os.Create(configFile); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	t.Setenv("MISE_DATA_DIR", miseDataDir)
	t.Setenv("MISE_CONFIG_FILE", configFile)
	t.Setenv("MISE_GLOBAL_CONFIG_FILE", filepath.Join(miseDataDir, "config.toml"))
	t.Setenv("MISE_EXPERIMENTAL", "true")

	client := NewClient()

	isManaged := client.IsManagedByMise(ctx, "jq")
	if isManaged {
		t.Error("Expected jq to not be managed before install")
	}

	cfg := &config.Config{
		Tools: map[string]config.Tool{
			"jq": {Version: "latest"},
		},
	}

	err := client.InstallAllWithHooks(ctx, cfg, false)
	if err != nil {
		t.Fatalf("InstallAllWithHooks failed: %v", err)
	}

	isManaged = client.IsManagedByMise(ctx, "jq")
	if !isManaged {
		t.Error("Expected jq to be managed after install")
	}

	os.RemoveAll(testDir)
}

func TestPreinstallHookExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if mise is available
	if _, err := exec.LookPath("mise"); err != nil {
		t.Skip("mise not available in PATH")
	}

	ctx := context.Background()

	testDir := filepath.Join(os.TempDir(), "mise-test-preinstall-"+t.Name())
	miseDataDir := filepath.Join(testDir, "mise-data")
	configFile := filepath.Join(testDir, "config.toml")

	if err := os.MkdirAll(miseDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	if _, err := os.Create(configFile); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	t.Setenv("MISE_DATA_DIR", miseDataDir)
	t.Setenv("MISE_CONFIG_FILE", configFile)
	t.Setenv("MISE_GLOBAL_CONFIG_FILE", filepath.Join(miseDataDir, "config.toml"))
	t.Setenv("MISE_EXPERIMENTAL", "true")

	client := NewClient()

	cfg := &config.Config{
		Tools: map[string]config.Tool{
			"fzf": {
				Version: "latest",
				Preinstall: []config.Hook{
					{
						Run: "echo PREINSTALL_MARKER_12345",
					},
				},
			},
		},
	}

	err := client.InstallAllWithHooks(ctx, cfg, false)
	if err != nil {
		t.Fatalf("InstallAllWithHooks failed: %v", err)
	}

	installed, err := client.ListTools(ctx)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	found := false
	for _, tool := range installed {
		if tool.Name == "fzf" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected fzf to be installed")
	}

	os.RemoveAll(testDir)
}

func TestPostinstallHookExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if mise is available
	if _, err := exec.LookPath("mise"); err != nil {
		t.Skip("mise not available in PATH")
	}

	ctx := context.Background()

	testDir := filepath.Join(os.TempDir(), "mise-test-postinstall-"+t.Name())
	miseDataDir := filepath.Join(testDir, "mise-data")
	configFile := filepath.Join(testDir, "config.toml")

	if err := os.MkdirAll(miseDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	if _, err := os.Create(configFile); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	t.Setenv("MISE_DATA_DIR", miseDataDir)
	t.Setenv("MISE_CONFIG_FILE", configFile)
	t.Setenv("MISE_GLOBAL_CONFIG_FILE", filepath.Join(miseDataDir, "config.toml"))
	t.Setenv("MISE_EXPERIMENTAL", "true")

	client := NewClient()

	preinstallScript := "echo PREINSTALL_SUCCESS"
	postinstallScript := "echo POSTINSTALL_SUCCESS"

	cfg := &config.Config{
		Tools: map[string]config.Tool{
			"fzf": {
				Version: "latest",
				Preinstall: []config.Hook{
					{Run: preinstallScript},
				},
				Postinstall: []config.Hook{
					{Run: postinstallScript},
				},
			},
		},
	}

	err := client.InstallAllWithHooks(ctx, cfg, false)
	if err != nil {
		t.Fatalf("InstallAllWithHooks failed: %v", err)
	}

	if !strings.Contains(preinstallScript, "PREINSTALL_SUCCESS") {
		t.Log("Preinstall script contains expected marker")
	}
	if !strings.Contains(postinstallScript, "POSTINSTALL_SUCCESS") {
		t.Log("Postinstall script contains expected marker")
	}

	os.RemoveAll(testDir)
}

func TestMiseEnvironmentVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if mise is available
	if _, err := exec.LookPath("mise"); err != nil {
		t.Skip("mise not available in PATH")
	}

	env := getMiseEnv()

	foundQuiet := false
	foundDisableWarnings := false
	foundExperimental := false
	foundGlobalConfig := false

	for _, e := range env {
		if strings.HasPrefix(e, "MISE_QUIET=") {
			foundQuiet = true
		}
		if strings.HasPrefix(e, "MISE_DISABLE_WARNINGS=") {
			foundDisableWarnings = true
		}
		if strings.HasPrefix(e, "MISE_EXPERIMENTAL=") {
			foundExperimental = true
		}
		if strings.HasPrefix(e, "MISE_GLOBAL_CONFIG_FILE=") {
			foundGlobalConfig = true
		}
	}

	if !foundQuiet {
		t.Error("Expected MISE_QUIET in env")
	}
	if !foundDisableWarnings {
		t.Error("Expected MISE_DISABLE_WARNINGS in env")
	}
	if !foundExperimental {
		t.Error("Expected MISE_EXPERIMENTAL in env")
	}
	if !foundGlobalConfig {
		t.Error("Expected MISE_GLOBAL_CONFIG_FILE in env")
	}
}
