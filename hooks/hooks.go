package hooks

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Runner executes preinstall/postinstall hooks
type Runner struct {
	dryRun     bool
	timeout    time.Duration
	verbose    bool
}

// HookResult represents the result of a hook execution
type HookResult struct {
	Script   string
	ExitCode int
	Stdout   string
	Stderr   string
	Error    error
	Duration time.Duration
}

// NewRunner creates a new hook runner
func NewRunner(dryRun bool) *Runner {
	return &Runner{
		dryRun:  dryRun,
		timeout:  5 * time.Minute,
		verbose:  false,
	}
}

// SetTimeout sets the hook execution timeout
func (r *Runner) SetTimeout(timeout time.Duration) {
	r.timeout = timeout
}

// SetVerbose enables verbose output
func (r *Runner) SetVerbose(verbose bool) {
	r.verbose = verbose
}

// Run executes a hook script
func (r *Runner) Run(ctx context.Context, script string) (*HookResult, error) {
	if r.dryRun {
		return &HookResult{
			Script:   script,
			ExitCode: 0,
			Stdout:   "[dry-run] Would execute: " + script,
		}, nil
	}

	if script == "" {
		return &HookResult{Script: script, ExitCode: 0}, nil
	}

	// Use sh -c to execute the script
	cmd := exec.CommandContext(ctx, "sh", "-c", script)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()

	err := cmd.Run()
	duration := time.Since(startTime)

	result := &HookResult{
		Script:   script,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Error = fmt.Errorf("hook failed with exit code %d: %s", exitErr.ExitCode(), stderr.String())
		} else {
			result.Error = fmt.Errorf("hook execution failed: %w", err)
		}
	}

	return result, result.Error
}

// RunHooks executes multiple hooks and returns all results
func (r *Runner) RunHooks(ctx context.Context, scripts []string) ([]*HookResult, error) {
	results := make([]*HookResult, 0, len(scripts))
	var lastError error

	for _, script := range scripts {
		script = strings.TrimSpace(script)
		if script == "" {
			continue
		}

		result, err := r.Run(ctx, script)
		results = append(results, result)

		if err != nil {
			lastError = err
			// Continue executing other hooks but return error at the end
		}
	}

	return results, lastError
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
