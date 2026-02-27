package mise

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mise-seq/config-loader/config"
)

// Bootstrapper handles bootstrapping (installing mise and cue if needed)
type Bootstrapper struct {
	misePath string
	cuePath  string
	version  string
}

// NewBootstrapper creates a new bootstrapper
func NewBootstrapper() *Bootstrapper {
	return &Bootstrapper{
		misePath: "",
		cuePath:  "",
		version:  "",
	}
}

// SetVersion sets the version to bootstrap
func (b *Bootstrapper) SetVersion(version string) {
	b.version = version
}

// FindMise checks if mise is available
func (b *Bootstrapper) FindMise() (string, error) {
	// Check if already available
	path, err := exec.LookPath("mise")
	if err == nil {
		b.misePath = path
		return path, nil
	}

	// Check MISE_INSTALLATION env var
	if installDir := os.Getenv("MISE_INSTALLATION"); installDir != "" {
		misePath := filepath.Join(installDir, "bin", "mise")
		if _, err := os.Stat(misePath); err == nil {
			b.misePath = misePath
			return misePath, nil
		}
	}

	return "", fmt.Errorf("mise not found")
}

// FindCue checks if cue is available
func (b *Bootstrapper) FindCue() (string, error) {
	// Check if already available
	path, err := exec.LookPath("cue")
	if err == nil {
		b.cuePath = path
		return path, nil
	}

	return "", fmt.Errorf("cue not found")
}

// InstallCue installs cue using mise
func (b *Bootstrapper) InstallCue(ctx context.Context) error {
	version := b.version
	if version == "" {
		version = os.Getenv("CUE_VERSION")
		if version == "" {
			version = "latest"
		}
	}

	// Use mise to install cue
	toolSpec := fmt.Sprintf("cue@%s", version)

	// First ensure mise is available
	if _, err := b.FindMise(); err != nil {
		return fmt.Errorf("mise not available to install cue: %w", err)
	}

	// Install cue using mise
	cmd := exec.CommandContext(ctx, "mise", "use", "-g", toolSpec)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install cue: %w\noutput: %s", err, string(output))
	}

	// Find cue again
	cuePath, err := b.FindCue()
	if err != nil {
		return fmt.Errorf("cue still not found after installation: %w", err)
	}

	config.Info("Installed cue: %s", cuePath)
	return nil
}

// EnsureCue ensures cue is available, installing if needed
func (b *Bootstrapper) EnsureCue(ctx context.Context) error {
	if _, err := b.FindCue(); err == nil {
		return nil // cue already available
	}

	config.Info("cue not found, installing...")
	return b.InstallCue(ctx)
}

// EnsureMise ensures mise is available, installing if needed
func (b *Bootstrapper) EnsureMise(ctx context.Context) error {
	if _, err := b.FindMise(); err == nil {
		return nil // mise already available
	}

	config.Info("mise not found, please install mise first")
	config.Info("See: https://github.com/jdx/mise")
	return fmt.Errorf("mise not found")
}

// HasRequiredTools checks if required tools are available
func (b *Bootstrapper) HasRequiredTools() (bool, error) {
	_, miseErr := b.FindMise()
	_, cueErr := b.FindCue()

	hasMise := miseErr == nil
	hasCue := cueErr == nil

	return hasMise && hasCue, nil
}

// MissingTools returns a list of missing required tools
func (b *Bootstrapper) MissingTools() []string {
	var missing []string

	if _, findErr := b.FindMise(); findErr != nil {
		missing = append(missing, "mise")
	}

	if _, findErr := b.FindCue(); findErr != nil {
		missing = append(missing, "cue")
	}

	return missing
}

// ParseToolSpec parses a tool specification (e.g., "node@20.0.0")
func ParseToolSpec(spec string) (name, version string, err error) {
	parts := strings.SplitN(spec, "@", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid tool spec: %s", spec)
	}
	return parts[0], parts[1], nil
}

// ParseToolsFromString parses tools from a string (one per line or comma-separated)
func ParseToolsFromString(s string) []string {
	// Split by newlines or commas
	lines := strings.Split(s, "\n")
	var tools []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Handle comma-separated
		parts := strings.Split(line, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				tools = append(tools, p)
			}
		}
	}
	return tools
}
