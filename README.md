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
    plugin: glab
  - name: difit
    package: npm:difit
  - name: gemini
    package: npm:@google/gemini-cli
    disabled: true

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
  "tools": [
    { "name": "jq", "version": "latest" },
    { "name": "lazygit", "version": "latest" }
  ],
  "defaults": {
    "preinstall": [{ "run": "echo Installing..." }]
  }
}
```

### TOML

```toml
[[tools]]
name = "jq"
version = "latest"

[[tools]]
name = "lazygit"
version = "latest"

[defaults.preinstall]
run = "echo Installing..."

[settings.npm]
package_manager = "pnpm"
```

### CUE

```cue
MiseSeqConfig: {
    tools: [
        {name: "jq", version: "latest"},
        {name: "lazygit", version: "latest"}
    ]
    defaults: preinstall: [{run: "echo Installing..."}]
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

#### Tool Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| name | string | required | Tool name |
| version | string | "latest" | Tool version |
| package | string | - | runtime:tool format |
| plugin | string | - | mise plugin to add |
| exe | string | name | Executable name |
| disabled | bool | false | Skip installation |
| depends | array | [] | Dependencies |
| preinstall | array | [] | Hooks before install |
| postinstall | array | [] | Hooks after install |

#### Dependency Syntax

```yaml
tools:
  - name: rust
    version: 1.88
    depends:
      - gcc
      - cargo
```

**Key Points:**
- version defaults to "latest"
- exe defaults to tool name

---

## Hooks

Hook types:
- preinstall: before tool installation
- postinstall: after tool installation

### Timing

- install: first install only
- update: version upgrade only
- always: every run

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

SHA256 markers track hook changes:
- First run: executes hook, saves SHA256
- Next runs: compares SHA256, skips if unchanged
- --force-hooks: force execution
- --postinstall-on-update: run postinstall on version change

---

## CLI Reference

### Commands

| Command  | Description |
|----------|-------------|
| install  | Install all tools |
| upgrade  | Upgrade tools |
| list     | List tools |
| status   | Show tool status |

### Global Flags

| Flag | Description |
|------|-------------|
| -c file | Config file |
| --dry-run | Dry run |
| --force-hooks | Force hook execution |
| --postinstall-on-update | Run postinstall on update |
| -v | Verbose |
| --version | Show version |
| --help | Show help |

### Environment Variables

| Variable | Description |
|----------|-------------|
| DRY_RUN | Dry run mode |
| DEBUG | Debug output |
| FORCE_HOOKS | Force hook execution |
| RUN_POSTINSTALL_ON_UPDATE | Run postinstall on update |
| STATE_DIR | Custom state directory |
| CUE_VERSION | CUE version |
| MISE_SHIMS_DEFAULT | Mise shims path |
| MISE_DATA_DIR | Mise data directory |

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

### Auto-install

If mise is not installed:
- Downloads mise to ~/.local/bin/mise
- Adds to PATH
- mise becomes available

After installation:

```bash
exec $SHELL
```

Verify: `mise --version`

---

## License

MIT
