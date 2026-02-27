package mise

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mise-seq/config-loader/config"
)

// ApplyMiseSettings applies mise settings from config
func (c *Client) ApplyMiseSettings(ctx context.Context, cfg *config.Config) error {
	if cfg.Settings == nil {
		return nil
	}

	// Apply npm.package_manager
	if cfg.Settings.NPM.PackageManager != "" {
		key := "npm.package_manager"
		value := cfg.Settings.NPM.PackageManager
		if err := c.runMiseSettings(ctx, key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	// Apply experimental
	if cfg.Settings.Experimental != "" {
		key := "experimental"
		value := cfg.Settings.Experimental
		if err := c.runMiseSettings(ctx, key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	return nil
}

// runMiseSettings runs mise settings set command
func (c *Client) runMiseSettings(ctx context.Context, key, value string) error {
	cmd := exec.CommandContext(ctx, "mise", "settings", "set", key, value)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mise settings set failed: %w, output: %s", err, string(output))
	}
	return nil
}

// GetDefaultsHooks returns the default hooks from config
func GetDefaultsHooks(cfg *config.Config) (preinstall, postinstall []config.Hook) {
	if cfg == nil || cfg.Defaults == nil {
		return nil, nil
	}
	return cfg.Defaults.Preinstall, cfg.Defaults.Postinstall
}

// GetToolHooks returns the hooks for a specific tool
func GetToolHooks(cfg *config.Config, toolName string) (preinstall, postinstall []config.Hook) {
	if cfg == nil || cfg.Tools == nil {
		return nil, nil
	}

	tool, exists := cfg.Tools[toolName]
	if !exists {
		return nil, nil
	}

	return tool.Preinstall, tool.Postinstall
}

// HasDefaults checks if config has default hooks
func HasDefaults(cfg *config.Config) bool {
	if cfg == nil || cfg.Defaults == nil {
		return false
	}
	return len(cfg.Defaults.Preinstall) > 0 || len(cfg.Defaults.Postinstall) > 0
}

// FilterHooksByWhen filters hooks by their "when" condition
// If hook.When is empty, it's treated as always run
func FilterHooksByWhen(hooks []config.Hook, when config.When) []config.Hook {
	var filtered []config.Hook
	for _, hook := range hooks {
		if len(hook.When) == 0 {
			filtered = append(filtered, hook)
			continue
		}
		for _, w := range hook.When {
			if w == when || w == config.WhenAlways {
				filtered = append(filtered, hook)
				break
			}
		}
	}
	return filtered
}

// ExtractHookScripts extracts the "run" field from hooks
func ExtractHookScripts(hooks []config.Hook) []string {
	scripts := make([]string, 0, len(hooks))
	for _, hook := range hooks {
		script := strings.TrimSpace(hook.Run)
		if script != "" {
			scripts = append(scripts, script)
		}
	}
	return scripts
}
