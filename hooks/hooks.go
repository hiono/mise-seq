package hooks

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// HookType represents the type of hook
type HookType string

const (
	HookTypePreinstall  HookType = "preinstall"
	HookTypePostinstall HookType = "postinstall"
)

// Runner executes preinstall/postinstall hooks
type Runner struct {
	dryRun     bool
	timeout    time.Duration
	verbose    bool
	stateMgr   *StateManager
}

// SetVerbose sets verbose mode
func (r *Runner) SetVerbose(v bool) {
	r.verbose = v
}

// IsVerbose returns verbose mode
func (r *Runner) IsVerbose() bool {
	return r.verbose
}

// HookResult represents the result of a hook execution
type HookResult struct {
	ToolName   string
	HookType   HookType
	Script     string
	ExitCode   int
	Stdout     string
	Stderr     string
	Error      error
	Duration   time.Duration
	Skipped    bool
	SHA256Hash string
}

// NewRunner creates a new hook runner
func NewRunner(dryRun bool) *Runner {
	return &Runner{
		dryRun:  dryRun,
		timeout:  5 * time.Minute,
		verbose:  false,
		stateMgr: NewStateManager(),
	}
}

// NewRunnerWithOptions creates a runner with custom options
func NewRunnerWithOptions(dryRun bool, stateDir string, forceHooks, runPostinstallOnUpdate bool) *Runner {
	runner := NewRunner(dryRun)
	if stateDir != "" {
		runner.stateMgr.StateDir = stateDir
	}
	runner.stateMgr.ForceHooks = forceHooks
	runner.stateMgr.RunPostinstallOnUpdate = runPostinstallOnUpdate
	return runner
}

// RunDefaultsHook runs a default hook with state management
// The key is used for state tracking (e.g., "defaults.preinstall")
func (r *Runner) RunDefaultsHook(ctx context.Context, hookType HookType, scripts []string) ([]*HookResult, error) {
	toolName := "defaults"
	results := make([]*HookResult, 0, len(scripts))
	var lastError error

	for _, script := range scripts {
		script = strings.TrimSpace(script)
		if script == "" {
			continue
		}

		result, err := r.Run(ctx, toolName, hookType, script)
		results = append(results, result)

		if err != nil {
			lastError = err
		}
	}

	return results, lastError
}

// SetTimeout sets the hook execution timeout
func (r *Runner) SetTimeout(timeout time.Duration) {
	r.timeout = timeout
}

// Run executes a hook script with state management
func (r *Runner) Run(ctx context.Context, toolName string, hookType HookType, script string) (*HookResult, error) {
	result := &HookResult{
		ToolName: toolName,
		HookType: hookType,
		Script:   script,
	}

	// Check if hook should run based on state
	shouldRun, existingHash, err := r.stateMgr.ShouldRunHook(toolName, string(hookType), script)
	if err != nil {
		return result, fmt.Errorf("failed to check hook state: %w", err)
	}

	result.SHA256Hash = existingHash

	if !shouldRun {
		result.Skipped = true
		if r.verbose {
			result.Stdout = fmt.Sprintf("[skip] Hook unchanged (SHA256: %s)", existingHash[:8])
		}
		return result, nil
	}

	if r.dryRun {
		result.Stdout = "[dry-run] Would execute: " + script
		return result, nil
	}

	if script == "" {
		return result, nil
	}

	// Use sh -c to execute the script
	cmd := exec.CommandContext(ctx, "sh", "-c", script)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()

	err = cmd.Run()
	duration := time.Since(startTime)

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	result.Duration = duration

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Error = fmt.Errorf("hook failed with exit code %d: %s", exitErr.ExitCode(), stderr.String())
		} else {
			result.Error = fmt.Errorf("hook execution failed: %w", err)
		}
	}

	// Save state on success (or always save for tracking)
	if result.Error == nil || r.stateMgr.ForceHooks {
		if saveErr := r.stateMgr.SaveHookState(toolName, string(hookType), script); saveErr != nil {
			// Log but don't fail
			if r.verbose {
				result.Stdout += fmt.Sprintf("\n[warn] Failed to save state: %v", saveErr)
			}
		}
	}

	return result, result.Error
}

// RunHooks executes multiple hooks for a tool and returns all results
func (r *Runner) RunHooks(ctx context.Context, toolName string, hookType HookType, scripts []string) ([]*HookResult, error) {
	results := make([]*HookResult, 0, len(scripts))
	var lastError error

	for _, script := range scripts {
		script = strings.TrimSpace(script)
		if script == "" {
			continue
		}

		result, err := r.Run(ctx, toolName, hookType, script)
		results = append(results, result)

		if err != nil {
			lastError = err
			// Continue executing other hooks but return error at the end
		}
	}

	return results, lastError
}

// RunToolHooks runs both preinstall and postinstall hooks for a tool
func (r *Runner) RunToolHooks(ctx context.Context, toolName string, preinstall, postinstall []string) ([]*HookResult, error) {
	var allResults []*HookResult
	var lastError error

	// Run preinstall hooks
	if len(preinstall) > 0 {
		results, err := r.RunHooks(ctx, toolName, HookTypePreinstall, preinstall)
		allResults = append(allResults, results...)
		if err != nil {
			lastError = err
		}
	}

	// Run postinstall hooks
	if len(postinstall) > 0 {
		results, err := r.RunHooks(ctx, toolName, HookTypePostinstall, postinstall)
		allResults = append(allResults, results...)
		if err != nil {
			lastError = err
		}
	}

	return allResults, lastError
}

// ParseHooksFromConfig extracts hook scripts from config structure
func ParseHooksFromConfig(hooks interface{}) []string {
	if hooks == nil {
		return nil
	}

	// Handle []interface{} case (from JSON/YAML parsing)
	if hookList, ok := hooks.([]interface{}); ok {
		scripts := make([]string, 0, len(hookList))
		for _, h := range hookList {
			if hookMap, ok := h.(map[string]interface{}); ok {
				if run, ok := hookMap["run"].(string); ok && run != "" {
					scripts = append(scripts, run)
				}
			}
		}
		return scripts
	}

	return nil
}
