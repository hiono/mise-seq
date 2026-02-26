# config-loader

A Go library for loading configuration files (CUE, JSON, YAML, TOML) and wrapping [mise](https://github.com/jdx/mise) CLI for tool management.

---

## Features

- **Multi-format config loading**: JSON, YAML, TOML, CUE
- **Unified Loader API**: Auto-detect format and parse with single call
- **mise CLI wrapper**: Install, upgrade, list tools with Go
- **Hook support**: Run preinstall/postinstall hooks during tool installation
- **Ordered installation**: Respect `tools_order` for sequential installs

---

## Installation

```bash
go get github.com/mise-seq/config-loader
```

---

## Quick Start

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
  jq:
    version: latest
  lazygit:
    version: latest

tools_order:
  - jq
  - lazygit
```

### JSON

```json
{
  "tools": {
    "jq": { "version": "latest" },
    "lazygit": { "version": "latest" }
  },
  "tools_order": ["jq", "lazygit"]
}
```

### TOML

```toml
[tools.jq]
version = "latest"

[tools.lazygit]
version = "latest"

tools_order = ["jq", "lazygit"]
```

---

## Hooks

The following hook types are supported:

- `preinstall`: Run before tool installation
- `postinstall`: Run after tool installation

### Hook Example

```yaml
tools:
  lazygit:
    version: latest
    preinstall:
      - run: |
          echo "Installing lazygit..."
    postinstall:
      - run: |
          mkdir -p "$HOME/.config/lazygit"
```

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
```

### mise package

```go
client := mise.NewClient()

// Install a single tool
result, err := client.InstallWithOutput(ctx, "jq@latest")

// Check if tool is installed
installed, err := client.IsInstalled(ctx, "jq")

// Install if not already installed
installed, result, err := client.InstallIfNotInstalled(ctx, "jq@latest")

// Upgrade a tool
result, err := client.UpgradeWithOutput(ctx, "jq")

// List installed tools
tools, err := client.ListTools(ctx)

// Install with hooks (respects tools_order)
err := client.InstallAllWithHooks(ctx, cfg)
```

### hooks package

```go
runner := hooks.NewRunner(false)

// Run a single hook
result, err := runner.Run(ctx, "echo hello")

// Run multiple hooks
results, err := runner.RunHooks(ctx, []string{
    "echo first",
    "echo second",
})
```

---

## Prerequisites

- Go 1.21+
- [mise](https://github.com/jdx/mise) CLI installed system-wide

---

## License

MIT
