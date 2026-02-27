package mise

import (
	"context"
	"encoding/json"
	"fmt"
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

	for _, t := range result.Tools {
		if t.Name == targetTool {
			return true, nil
		}
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
	return true, result, err
}

// UpgradeWithOutput upgrades a tool and captures output
func (c *Client) UpgradeWithOutput(ctx context.Context, tool string) (*Result, error) {
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "mise", "upgrade", tool)
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
	Tools []ToolInfo `json:"tools"`
}

// ListWithOutput runs mise ls --json and returns structured output
func (c *Client) ListWithOutput(ctx context.Context) (*ListResult, error) {
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "mise", "ls", "--json")
	cmd.Stdout = nil
	cmd.Stderr = nil

	stdout := new(strings.Builder)
	stderr := new(strings.Builder)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			_ = exitErr
			// Try to parse even on error (mise returns error for no tools)
			if stdout.Len() == 0 {
				return &ListResult{Tools: []ToolInfo{}}, nil
			}
			// Return whatever we got
		}
	}

	var result ListResult
	if err := json.Unmarshal([]byte(stdout.String()), &result); err != nil {
		// Return empty result on parse error
		return &ListResult{Tools: []ToolInfo{}}, nil
	}

	return &result, nil
}

// ListTools lists all installed tools
func (c *Client) ListTools(ctx context.Context) ([]ToolInfo, error) {
	result, err := c.ListWithOutput(ctx)
	if err != nil {
		return nil, err
	}
	return result.Tools, nil
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
	for _, tool := range result.Tools {
		fmt.Printf("%s %s (%s)\n", tool.Name, tool.Version, tool.Source)
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
	if tool.Preinstall != nil {
		scripts := hooks.ParseHooksFromConfig(tool.Preinstall)
		_, err := hookRunner.RunHooks(ctx, toolName, hooks.HookTypePreinstall, scripts)
		if err != nil {
			return fmt.Errorf("preinstall hook failed for %s: %w", toolName, err)
		}
	}

	// Install tool
	toolSpec := fmt.Sprintf("%s@%s", toolName, tool.Version)
	_, result, err := c.InstallIfNotInstalled(ctx, toolSpec)
	if err != nil {
		return fmt.Errorf("install failed for %s: %w", toolName, err)
	}
	if result != nil && result.Error != nil {
		return fmt.Errorf("install error for %s: %w", toolName, result.Error)
	}

	// Run postinstall hooks
	if tool.Postinstall != nil {
		scripts := hooks.ParseHooksFromConfig(tool.Postinstall)
		_, err := hookRunner.RunHooks(ctx, toolName, hooks.HookTypePostinstall, scripts)
		if err != nil {
			return fmt.Errorf("postinstall hook failed for %s: %w", toolName, err)
		}
	}

	return nil
}

// InstallAllWithHooks installs all tools from config with hooks, respecting tools_order
func (c *Client) InstallAllWithHooks(ctx context.Context, cfg *config.Config) error {
	toolOrder := config.GetToolOrder(cfg)
	tools := config.GetTools(cfg)

	if len(toolOrder) == 0 {
		// No order specified, use tools map keys
		for name := range tools {
			if err := c.InstallWithHooks(ctx, cfg, name); err != nil {
				return err
			}
		}
		return nil
	}

	// Install in specified order
	for _, name := range toolOrder {
		if err := c.InstallWithHooks(ctx, cfg, name); err != nil {
			return err
		}
	}

	return nil
}
