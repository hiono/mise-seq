package hooks

import "context"

// HookExecutor defines the interface for hook execution
type HookExecutor interface {
	Run(ctx context.Context, toolName string, hookType HookType, script string) (*HookResult, error)
	RunHooks(ctx context.Context, toolName string, hookType HookType, scripts []string) ([]*HookResult, error)
}

// StateReader defines the interface for reading hook state
type StateReader interface {
	ReadMarker(toolName, hookType string) (string, error)
}

// StateWriter defines the interface for writing hook state
type StateWriter interface {
	WriteMarker(toolName, hookType, hash string) error
}
