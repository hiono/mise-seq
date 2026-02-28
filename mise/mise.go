package mise

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mise-seq/config-loader/config"
	"github.com/mise-seq/config-loader/hooks"
)

// Client wraps mise CLI invocations
type Client struct {
	timeout time.Duration
}

// NewClient creates a new mise client
func NewClient() *Client {
	return &Client{
		timeout: 10 * time.Minute,
	}
}

// SetTimeout sets the command timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// getMiseEnv returns common environment variables for mise commands
func getMiseEnv() []string {
	miseDataDir := os.Getenv("MISE_DATA_DIR")
	if miseDataDir == "" {
		miseDataDir = os.ExpandEnv("$HOME/.local/share/mise")
	}
	return []string{
		"MISE_QUIET=1",
		"MISE_DISABLE_WARNINGS=1",
		"MISE_EXPERIMENTAL=true",
		"MISE_GLOBAL_CONFIG_FILE=" + miseDataDir + "/config.toml",
	}
}

// Result represents the result of a mise command
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

// InstallWithOutput installs a tool and captures output
func (c *Client) InstallWithOutput(ctx context.Context, tool string) (*Result, error) {
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "mise", "install", tool)
	cmd.Env = append(os.Environ(), getMiseEnv()...)
	cmd.Env = append(cmd.Env,
		"GITHUB_TOKEN="+os.Getenv("GH_TOKEN"),
		"GITLAB_TOKEN="+os.Getenv("GITLAB_TOKEN"),
	)
	cmd.Stdout = nil
	cmd.Stderr = nil

	stdout := new(strings.Builder)
	stderr := new(strings.Builder)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()

	result := &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Error = fmt.Errorf("mise install failed: %w", err)
		} else {
			result.Error = err
		}
	}

	return result, nil
}

// IsInstalled checks if a tool is already installed
func (c *Client) IsInstalled(ctx context.Context, tool string) (bool, error) {
	result, err := c.ListWithOutput(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to list tools: %w", err)
	}

	// Parse tool name from "runtime:tool" format
	targetTool := strings.Split(tool, "@")[0]
	if idx := strings.LastIndex(targetTool, ":"); idx != -1 {
		targetTool = targetTool[idx+1:]
	}

	// Check if tool exists in the map (regardless of installed status)
	if _, exists := result.Tools[targetTool]; exists {
		return true, nil
	}

	return false, nil
}

// InstallIfNotInstalled installs a tool only if it's not already installed
func (c *Client) InstallIfNotInstalled(ctx context.Context, tool string) (bool, *Result, error) {
	installed, err := c.IsInstalled(ctx, tool)
	if err != nil {
		return false, nil, err
	}

	if installed {
		return false, nil, nil
	}

	result, err := c.InstallWithOutput(ctx, tool)
	if err != nil {
		// Check if it's a "not found in mise tool registry" error
		if result != nil && strings.Contains(result.Stderr, "not found in mise tool registry") {
			return false, nil, fmt.Errorf("tool %s not found in mise registry", tool)
		}
		return false, nil, err
	}
	return true, result, err
}

// SetGlobal sets a tool as global default (mise use -g)
func (c *Client) SetGlobal(ctx context.Context, tool string) error {
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "mise", "use", "-g", tool)
	cmd.Env = append(os.Environ(), getMiseEnv()...)
	cmd.Env = append(cmd.Env,
		"GITHUB_TOKEN="+os.Getenv("GH_TOKEN"),
		"GITLAB_TOKEN="+os.Getenv("GITLAB_TOKEN"),
	)

	stdout := new(strings.Builder)
	stderr := new(strings.Builder)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("mise use -g failed: %s", stderr.String())
	}
	return nil
}

// IsManagedByMise checks if a tool is already managed by mise (shim exists)
// Uses mise ls --json to check if tool is in the list
func (c *Client) IsManagedByMise(ctx context.Context, tool string) bool {
	result, err := c.ListWithOutput(ctx)
	if err != nil {
		return false
	}

	// Parse tool name from "runtime:tool" format
	targetTool := strings.Split(tool, "@")[0]
	if idx := strings.LastIndex(targetTool, ":"); idx != -1 {
		targetTool = targetTool[idx+1:]
	}

	// Check if tool exists in the map
	if _, exists := result.Tools[targetTool]; exists {
		return true
	}
	return false
}

// UpgradeWithOutput upgrades a tool and captures output
func (c *Client) UpgradeWithOutput(ctx context.Context, tool string) (*Result, error) {
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "mise", "upgrade", tool)
	cmd.Env = append(os.Environ(), getMiseEnv()...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	stdout := new(strings.Builder)
	stderr := new(strings.Builder)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()

	result := &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Error = fmt.Errorf("mise upgrade failed: %w", err)
		} else {
			result.Error = err
		}
	}

	return result, nil
}

// UpgradeIfInstalled upgrades a tool only if it's already installed
func (c *Client) UpgradeIfInstalled(ctx context.Context, tool string) (bool, *Result, error) {
	installed, err := c.IsInstalled(ctx, tool)
	if err != nil {
		return false, nil, err
	}

	if !installed {
		return false, nil, nil
	}

	result, err := c.UpgradeWithOutput(ctx, tool)
	return true, result, err
}

// ToolInfo represents information about an installed tool
type ToolInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Source  string `json:"source"`
}

// ListResult represents the result of mise ls --json
type ListResult struct {
	Tools map[string][]struct {
		Version          string `json:"version"`
		RequestedVersion string `json:"requested_version"`
		InstallPath      string `json:"install_path"`
		Installed        bool   `json:"installed"`
		Active           bool   `json:"active"`
	} `json:"-"`
	Raw string
}

// ListWithOutput runs mise ls --json and returns structured output
func (c *Client) ListWithOutput(ctx context.Context) (*ListResult, error) {
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "mise", "ls", "--json")
	cmd.Env = append(os.Environ(), getMiseEnv()...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	stdout := new(strings.Builder)
	stderr := new(strings.Builder)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()

	result := &ListResult{
		Tools: make(map[string][]struct {
			Version          string `json:"version"`
			RequestedVersion string `json:"requested_version"`
			InstallPath      string `json:"install_path"`
			Installed        bool   `json:"installed"`
			Active           bool   `json:"active"`
		}),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			_ = exitErr
			// Try to parse even on error (mise returns error for no tools)
			if stdout.Len() == 0 {
				return result, nil
			}
		}
	}

	result.Raw = stdout.String()

	// Parse as map[string][]ToolInfo
	var toolsMap map[string][]struct {
		Version          string `json:"version"`
		RequestedVersion string `json:"requested_version"`
		InstallPath      string `json:"install_path"`
		Installed        bool   `json:"installed"`
		Active           bool   `json:"active"`
	}
	if err := json.Unmarshal([]byte(stdout.String()), &toolsMap); err != nil {
		return result, nil
	}

	result.Tools = toolsMap
	return result, nil
}

// ListTools lists all installed tools
func (c *Client) ListTools(ctx context.Context) ([]ToolInfo, error) {
	result, err := c.ListWithOutput(ctx)
	if err != nil {
		return nil, err
	}

	var tools []ToolInfo
	for name, versions := range result.Tools {
		if len(versions) > 0 {
			tools = append(tools, ToolInfo{
				Name:    name,
				Version: versions[0].Version,
			})
		}
	}
	return tools, nil
}

// Install installs a tool via mise (legacy method)
func (c *Client) Install(ctx context.Context, tool string) error {
	result, err := c.InstallWithOutput(ctx, tool)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("install failed with exit code %d: %s", result.ExitCode, result.Stderr)
	}
	return nil
}

// Upgrade upgrades a tool via mise (legacy method)
func (c *Client) Upgrade(ctx context.Context, tool string) error {
	result, err := c.UpgradeWithOutput(ctx, tool)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("upgrade failed with exit code %d: %s", result.ExitCode, result.Stderr)
	}
	return nil
}

// List lists installed tools (legacy method)
func (c *Client) List(ctx context.Context) error {
	result, err := c.ListWithOutput(ctx)
	if err != nil {
		return err
	}
	for name, versions := range result.Tools {
		if len(versions) > 0 {
			version := versions[0].Version
			installed := versions[0].Installed
			fmt.Printf("%s %s (installed: %v)\n", name, version, installed)
		}
	}
	return nil
}

// InstallWithHooks installs a tool with preinstall/postinstall hooks
func (c *Client) InstallWithHooks(ctx context.Context, cfg *config.Config, toolName string) error {
	hookRunner := hooks.NewRunner(false)
	tool, exists := cfg.Tools[toolName]
	if !exists {
		return fmt.Errorf("tool %s not found in config", toolName)
	}

	// Run preinstall hooks
	if len(tool.Preinstall) > 0 {
		scripts := ExtractHookScripts(tool.Preinstall)
		results, err := hookRunner.RunHooks(ctx, toolName, hooks.HookTypePreinstall, scripts)
		if err != nil {
			// Output all hook output on error
			for _, result := range results {
				fmt.Print(result.Stdout)
				fmt.Fprint(os.Stderr, result.Stderr)
			}
			desc := ""
			if len(tool.Preinstall) > 0 && tool.Preinstall[0].Description != "" {
				desc = fmt.Sprintf(" (%s)", tool.Preinstall[0].Description)
			}
			return fmt.Errorf("preinstall hook%s failed for %s: %w", desc, toolName, err)
		}
		// Output on success too
		for _, result := range results {
			if result.Stdout != "" {
				fmt.Print(result.Stdout)
			}
			if result.Stderr != "" {
				fmt.Fprint(os.Stderr, result.Stderr)
			}
		}
	}

	// Install tool
	// Use toolName (full spec) for mise install, exeName is for reference only
	toolSpec := fmt.Sprintf("%s@%s", toolName, tool.Version)
	_, result, err := c.InstallIfNotInstalled(ctx, toolSpec)
	if err != nil {
		// Check if it's a "not found" error - skip this tool
		if strings.Contains(err.Error(), "not found in mise registry") {
			fmt.Printf("[WARN] Tool %s not found in mise registry, skipping\n", toolName)
			return nil
		}
		return fmt.Errorf("install failed for %s: %w", toolName, err)
	}
	if result != nil && result.Error != nil {
		// Check if it's a "not found" error - skip this tool
		if strings.Contains(result.Error.Error(), "not found in mise registry") {
			fmt.Printf("[WARN] Tool %s not found in mise registry, skipping\n", toolName)
			return nil
		}
		return fmt.Errorf("install error for %s: %w", toolName, result.Error)
	}

	// Set as global default (equivalent to mise use -g)
	if err := c.SetGlobal(ctx, toolSpec); err != nil {
		return fmt.Errorf("failed to set global default for %s: %w", toolName, err)
	}

	// Run postinstall hooks
	if len(tool.Postinstall) > 0 {
		scripts := ExtractHookScripts(tool.Postinstall)
		results, err := hookRunner.RunHooks(ctx, toolName, hooks.HookTypePostinstall, scripts)
		if err != nil {
			// Output all hook output on error
			for _, result := range results {
				fmt.Print(result.Stdout)
				fmt.Fprint(os.Stderr, result.Stderr)
			}
			desc := ""
			if len(tool.Postinstall) > 0 && tool.Postinstall[0].Description != "" {
				desc = fmt.Sprintf(" (%s)", tool.Postinstall[0].Description)
			}
			return fmt.Errorf("postinstall hook%s failed for %s: %w", desc, toolName, err)
		}
		// Output on success too
		for _, result := range results {
			if result.Stdout != "" {
				fmt.Print(result.Stdout)
			}
			if result.Stderr != "" {
				fmt.Fprint(os.Stderr, result.Stderr)
			}
		}
	}

	return nil
}

// InstallAllWithHooks installs all tools from config with hooks, respecting tools_order and dependencies
func (c *Client) InstallAllWithHooks(ctx context.Context, cfg *config.Config, runPostinstallOnUpdate bool) error {
	toolOrder := config.GetToolOrder(cfg)
	tools := config.GetTools(cfg)

	// Determine installation order
	var installOrder []string

	if len(toolOrder) > 0 {
		// Use tools_order if specified
		installOrder = toolOrder
	} else {
		// Use dependency resolver
		resolver := config.NewToolResolver(tools)
		var err error
		installOrder, err = resolver.ResolveOrder()
		if err != nil {
			return fmt.Errorf("failed to resolve dependency order: %w", err)
		}
	}

	// Install in determined order
	for _, name := range installOrder {
		tool, exists := tools[name]
		if !exists {
			continue
		}

		// Check if already managed by mise
		if c.IsManagedByMise(ctx, name) {
			// Tool is already managed - run update flow
			fmt.Printf("Upgrading %s (already managed by mise)\n", name)
			_, err := c.UpgradeWithOutput(ctx, name)
			if err != nil {
				return fmt.Errorf("failed to upgrade %s: %w", name, err)
			}

			// Run postinstall hooks (update phase)
			if runPostinstallOnUpdate && len(tool.Postinstall) > 0 {
				hookRunner := hooks.NewRunnerWithOptions(false, "", false, runPostinstallOnUpdate)
				scripts := ExtractHookScripts(tool.Postinstall)
				results, err := hookRunner.RunHooks(ctx, name, hooks.HookTypePostinstall, scripts)
				if err != nil {
					for _, result := range results {
						fmt.Print(result.Stdout)
						fmt.Fprint(os.Stderr, result.Stderr)
					}
					desc := ""
					if len(tool.Postinstall) > 0 && tool.Postinstall[0].Description != "" {
						desc = fmt.Sprintf(" (%s)", tool.Postinstall[0].Description)
					}
					return fmt.Errorf("postinstall hook%s failed for %s: %w", desc, name, err)
				}
				// Output on success too
				for _, result := range results {
					if result.Stdout != "" {
						fmt.Print(result.Stdout)
					}
					if result.Stderr != "" {
						fmt.Fprint(os.Stderr, result.Stderr)
					}
				}
			}
		} else {
			// Tool is not managed - run install flow
			fmt.Printf("Installing %s\n", name)
			if err := c.InstallWithHooks(ctx, cfg, name); err != nil {
				return err
			}
		}
	}

	return nil
}
