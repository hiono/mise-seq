package mise

import "context"

// ToolManager defines the interface for tool management
type ToolManager interface {
	Install(ctx context.Context, tool string) error
	Upgrade(ctx context.Context, tool string) error
	List(ctx context.Context) ([]string, error)
	IsInstalled(ctx context.Context, tool string) (bool, error)
}
