# mise-seq

[English](./README.md) | [日本語](./README.ja.md) | [中文](./README.zh.md)

---

A Go library and CLI for tool installation using mise.

---

## Features

- Multi-format config loading: JSON, YAML, TOML, CUE
- Unified Loader API: auto-detects format
- mise CLI wrapper: install, upgrade, list, status tools
- Plugin support: adds mise plugins before installation
- Package format: runtime:tool (npm:difit, go:github.com/...)
- Ordered list: list order = installation order
- Disable tools: skip installation with disabled: true
- Hook support: runs scripts before/after installation
- SHA256 state management: skips unchanged hooks
- Defaults: applies hooks to all tools
- Settings: configures mise (npm, experimental)
- CLI subcommands: install, upgrade, list, status

---

## Installation

### Binary

Download from [Releases](https://github.com/mise-seq/config-loader/releases)

### From Source

```bash
go install github.com/mise-seq/config-loader@latest
```

### Library

```bash
go get github.com/mise-seq/config-loader
```

---

## Quick Start

### CLI

```bash
# Install tools from config
mise-seq -c tools.yaml

# Dry run
mise-seq -c tools.yaml --dry-run

# Upgrade installed tools
mise-seq upgrade -c tools.yaml

# List installed tools
mise-seq list

# Show status
mise-seq status -c tools.yaml
```

### Library

```go
package main

import (
	"context"
	"log"

	"github.com/mise-seq/config-loader/config"
	"github.com/mise-seq/config-loader/mise"
)

func main() {
	ctx := context.Background()

	// Load configuration
	loader := config.NewLoader()
	cfg, err := loader.Parse("tools.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Install tools with hooks
	client := mise.NewClient()
	if err := client.InstallAllWithHooks(ctx, cfg); err != nil {
		log.Fatalf("Installation failed: %v", err)
	}
}
```

---

## Configuration Format

### YAML

```yaml
tools:
  - name: jq
    version: latest
  - name: lazygit
    version: latest
  - name: glab
    version: latest
    plugin: glab  # Add mise plugin before installation
  - name: difit
    package: npm:difit  # runtime:tool format
  - name: gemini
    package: npm:@google/gemini-cli
    disabled: true  # Skip installation

defaults:
  preinstall:
    - run: echo "Installing {{.ToolName}}..."
  postinstall:
    - run: echo "Installed {{.ToolName}}"

settings:
  npm:
    package_manager: pnpm
  experimental: true
```

### JSON

```json
{
  "tools": {
    "jq": { "version": "latest" },
    "lazygit": { "version": "latest" }
  },
  "tools_order": ["jq", "lazygig"],
  "defaults": {
    "preinstall": [{ "run": "echo Installing..." }]
  }
}
```

### TOML

```toml
[tools.jq]
version = "latest"

[tools.lazygit]
version = "latest"

tools_order = ["jq", "lazygit"]

[defaults.preinstall]
run = "echo Installing..."

[settings.npm]
package_manager = "pnpm"
```

### CUE

```cue
MiseSeqConfig: {
    tools: {
        jq: version: "latest"
        lazygit: version: "latest"
    }
    tools_order: ["jq", "lazygit"]
    defaults: preinstall: [{run: "echo Installing..."}]
}
```

### Configuration Reference

All fields are optional unless marked as required.

#### Tool Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Tool name (required) |
| `version` | string | No | `"latest"` | Tool version |
| `package` | string | No | - | runtime:tool format (e.g., `npm:difit`, `go:github.com/user/repo`) |
| `plugin` | string | No | - | mise plugin to add before installation |
| `exe` | string | No | `<name>` | Executable name |
| `disabled` | bool | No | `false` | Skip installation if true |
| `depends` | array | No | `[]` | Dependencies (tool names only) |
| `preinstall` | array | No | `[]` | Hooks to run before installation |
| `postinstall` | array | No | `[]` | Hooks to run after installation |

#### Dependency Syntax

```yaml
tools:
  - name: rust
    version: 1.88
    depends:
      - gcc
      - cargo
```
  rust:
    version: 1.88
    depends:
      - gcc    # same as gcc@latest
      - cargo  # same as cargo@latest
```

**Key Points:**
- @version defaults to @latest
- version field defaults to "latest" 
- exe field defaults to tool key name

#### Minimal Configuration (All Omitted)

```yaml
# All defaults: version=latest, exe=tool name, depends=[]
tools:
  jq:
  gcc:
  rust:
```

---

## Hooks

The following hook types are supported:

- `preinstall`: Run before tool installation
- `postinstall`: Run after tool installation

### Hook Timing

Hooks can be configured to run on specific events:

- `install`: Only on first install
- `update`: Only on version upgrade
- `always`: Always run

### Hook Example

```yaml
tools:
  lazygit:
    version: latest
    preinstall:
      - run: |
          echo "Installing lazygit..."
        when: ["install"]
    postinstall:
      - run: |
          mkdir -p "$HOME/.config/lazygit"
        when: ["always"]
```

### Defaults

Apply hooks to all tools:

```yaml
defaults:
  preinstall:
    - run: echo "Installing {{.ToolName}}"
  postinstall:
    - run: echo "Done installing {{.ToolName}}"
```

### State Management

Hooks use SHA256 markers to skip unchanged hooks:

- First run: executes hook, saves SHA256
- Subsequent runs: compares SHA256, skips if unchanged
- `--force-hooks`: force execution even if unchanged
- `--postinstall-on-update`: run postinstall on version change

---

## CLI Reference

### Commands

| Command   | Description                        |
|-----------|-----------------------------------|
| `install` | Install all tools (default)       |
| `upgrade` | Upgrade installed tools           |
| `list`    | List installed tools              |
| `status`  | Show status of configured tools   |

### Global Flags

| Flag                      | Description                        |
|---------------------------|------------------------------------|
| `-c <file>`               | Config file (default: tools.yaml)   |
| `--dry-run`               | Dry run mode                       |
| `--force-hooks`           | Force hook execution               |
| `--postinstall-on-update`| Run postinstall on version change |
| `-v`                      | Verbose output                     |
| `--version`               | Show version                       |
| `--help`                  | Show help                          |

### Environment Variables

| Variable                    | Description                    |
|-----------------------------|-------------------------------|
| `DRY_RUN`                  | Enable dry run mode            |
| `DEBUG`                    | Enable debug output           |
| `FORCE_HOOKS`              | Force hook execution          |
| `RUN_POSTINSTALL_ON_UPDATE`| Run postinstall on update     |
| `STATE_DIR`                | Custom state directory        |
| `CUE_VERSION`              | CUE version for bootstrap     |
| `MISE_SHIMS_DEFAULT`       | Mise shims path               |
| `MISE_DATA_DIR`            | Mise data directory           |

---

## API Reference

### config package

```go
// Create a new loader
loader := config.NewLoader()

// Parse a config file (auto-detects format)
cfg, err := loader.Parse("tools.yaml")

// Get tools from config
tools := config.GetTools(cfg)

// Get installation order
order := config.GetToolOrder(cfg)

// Check for defaults
hasDefaults := config.HasDefaults(cfg)

// Get default hooks
preinstall, postinstall := config.GetDefaultsHooks(cfg)
```

### mise package

```go
client := mise.NewClient()

// Install a single tool
result, err := client.InstallWithOutput(ctx, "jq@latest")

// Check if tool is installed
installed, err := client.IsInstalled(ctx, "jq")

// Check if tool is managed by mise (in config)
managed, err := client.IsManagedByMise(ctx, "jq")

// Install if not already installed
installed, result, err := client.InstallIfNotInstalled(ctx, "jq@latest")

// Set global default (equivalent to mise use -g)
err := client.SetGlobal(ctx, "jq@latest")

// Upgrade a tool
result, err := client.UpgradeWithOutput(ctx, "jq")

// Upgrade if installed
result, err := client.UpgradeIfInstalled(ctx, "jq")

// List installed tools
tools, err := client.ListTools(ctx)

// Install with hooks (respects tools_order, auto-detects install vs upgrade)
err := client.InstallAllWithHooks(ctx, cfg, runPostinstallOnUpdate)

// InstallAllWithHooks: runPostinstallOnUpdate=true runs postinstall hooks on upgrade

// Apply settings
err := client.ApplySettings(ctx, cfg.Settings)

// Bootstrap: ensure mise/cue available
bootstrapper := mise.NewBootstrapper()
err := bootstrapper.EnsureMise(ctx)
err := bootstrapper.EnsureCue(ctx)
```

### hooks package

```go
runner := hooks.NewRunner(false)

// Run with state management
result, err := runner.Run(ctx, "toolname", hooks.HookTypePreinstall, "echo hello")

// Run multiple hooks
results, err := runner.RunHooks(ctx, "toolname", hooks.HookTypePreinstall, []string{
    "echo first",
    "echo second",
})

// Run default hooks
results, err := runner.RunDefaultsHook(ctx, hooks.HookTypePreinstall, []string{
    "echo default hook",
})

// Custom runner with options
runner := hooks.NewRunnerWithOptions(false, "/custom/state", true, false)
```

---

## Prerequisites

Go 1.21+

### Auto-install behavior

If mise is not installed:
- Downloads mise to ~/.local/bin/mise
- Adds to PATH after download
- mise becomes available for subsequent commands

**After installation, restart your shell:**

```bash
exec $SHELL
```

Verify with: `mise --version`

---

## License

MIT
