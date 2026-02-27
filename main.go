package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/mise-seq/config-loader/config"
	"github.com/mise-seq/config-loader/hooks"
	"github.com/mise-seq/config-loader/mise"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Parse global flags
	configPath := flag.String("c", "tools.yaml", "Path to config file")
	dryRun := flag.Bool("dry-run", false, "Dry run mode")
	forceHooks := flag.Bool("force-hooks", false, "Force hook execution")
	postinstallOnUpdate := flag.Bool("postinstall-on-update", false, "Run postinstall on update")
	verbose := flag.Bool("v", false, "Verbose output")
	showVersion := flag.Bool("version", false, "Show version")
	stateDir := flag.String("state-dir", "", "Custom state directory")

	flag.Parse()

	args := flag.Args()

	// Handle version flag early
	if *showVersion {
		fmt.Printf("mise-seq version %s (commit: %s, date: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Determine subcommand
	subcommand := "install" // default
	if len(args) > 0 {
		subcommand = args[0]
		args = args[1:]
	}

	// Validate subcommand
	validSubcommands := []string{"install", "upgrade", "list", "status", "help"}
	valid := false
	for _, v := range validSubcommands {
		if subcommand == v {
			valid = true
			break
		}
	}
	if !valid {
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand '%s'\n", subcommand)
		fmt.Fprintf(os.Stderr, "Run 'mise-seq help' for usage.\n")
		os.Exit(1)
	}

	// Handle help
	if subcommand == "help" {
		printHelp()
		os.Exit(0)
	}

	// Initialize logger
	config.InitLogger(*verbose)

	// Load runtime config
	runtimeCfg := config.LoadRuntimeConfig()
	runtimeCfg.DryRun = *dryRun
	runtimeCfg.ForceHooks = *forceHooks
	runtimeCfg.RunPostinstallOnUpdate = *postinstallOnUpdate
	if *stateDir != "" {
		runtimeCfg.StateDir = *stateDir
	}
	if *verbose {
		runtimeCfg.Debug = true
	}

	ctx := context.Background()

	// Setup environment
	if err := runtimeCfg.SetupEnvironment(); err != nil {
		config.Error("Failed to setup environment: %v", err)
		os.Exit(1)
	}

	// Bootstrap
	bootstrapper := mise.NewBootstrapper()
	bootstrapper.SetVersion(runtimeCfg.CUEVersion)

	if err := bootstrapper.EnsureMise(ctx); err != nil {
		config.Error("mise is required but not found")
		config.Error("Please install mise: https://github.com/jdx/mise")
		os.Exit(1)
	}

	if err := bootstrapper.EnsureCue(ctx); err != nil {
		config.Warn("cue not available: %v", err)
		// Continue without CUE support
	}

	// Load config
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		config.Error("Config file not found: %s", *configPath)
		os.Exit(1)
	}

	loader := config.NewLoader()
	cfg, err := loader.Parse(*configPath)
	if err != nil {
		config.Error("Failed to load config: %v", err)
		os.Exit(1)
	}

	miseClient := mise.NewClient()

	// Execute subcommand
	switch subcommand {
	case "install":
		err = runInstall(ctx, cfg, miseClient, runtimeCfg, *verbose, *dryRun)
	case "upgrade":
		err = runUpgrade(ctx, cfg, miseClient, runtimeCfg, *verbose, *dryRun)
	case "list":
		err = runList(ctx, cfg, miseClient, *verbose)
	case "status":
		err = runStatus(ctx, cfg, miseClient, *verbose)
	}

	if err != nil {
		config.Error("%v", err)
		os.Exit(1)
	}
}

func runInstall(ctx context.Context, cfg *config.Config, client *mise.Client, runtimeCfg *config.RuntimeConfig, verbose, dryRun bool) error {
	config.Info("=== Installing tools ===")

	// Apply settings
	if verbose {
		config.Info("Applying mise settings...")
	}
	if err := client.ApplyMiseSettings(ctx, cfg); err != nil {
		config.Warn("Failed to apply settings: %v", err)
	}

	// Run defaults preinstall
	if config.HasDefaults(cfg) {
		if verbose {
			config.Info("Running default preinstall hooks...")
		}
		preinstall, _ := config.GetDefaultsHooks(cfg)
		if len(preinstall) > 0 {
			scripts := mise.ExtractHookScripts(preinstall)
			hookRunner := hooks.NewRunnerWithOptions(dryRun, runtimeCfg.StateDir, runtimeCfg.ForceHooks, runtimeCfg.RunPostinstallOnUpdate)
			hookRunner.SetVerbose(verbose)
			_, err := hookRunner.RunDefaultsHook(ctx, hooks.HookTypePreinstall, scripts)
			if err != nil {
				config.Warn("Default preinstall hooks failed: %v", err)
			}
		}
	}

	// Install tools
	if verbose {
		config.Info("Installing tools...")
	}
	if err := client.InstallAllWithHooks(ctx, cfg); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	config.Info("Installation complete!")
	return nil
}

func runUpgrade(ctx context.Context, cfg *config.Config, client *mise.Client, runtimeCfg *config.RuntimeConfig, verbose, dryRun bool) error {
	config.Info("=== Upgrading tools ===")

	// Get tools to upgrade
	tools := config.GetTools(cfg)
	if len(tools) == 0 {
		config.Info("No tools configured")
		return nil
	}

	// Upgrade each tool
	for toolName := range tools {
		if verbose {
			config.Info("Upgrading %s...", toolName)
		}
		_, err := client.UpgradeWithOutput(ctx, toolName)
		if err != nil {
			config.Warn("Failed to upgrade %s: %v", toolName, err)
		}
	}

	config.Info("Upgrade complete!")
	return nil
}

func runList(ctx context.Context, cfg *config.Config, client *mise.Client, verbose bool) error {
	config.Info("=== Configured tools ===")

	tools := config.GetTools(cfg)
	if len(tools) == 0 {
		config.Info("No tools configured")
		return nil
	}

	// Get installation order
	order := config.GetToolOrder(cfg)

	if len(order) > 0 {
		fmt.Println("Installation order:")
		for i, toolName := range order {
			tool, ok := tools[toolName]
			if !ok {
				continue
			}
			fmt.Printf("  %d. %s @ %s\n", i+1, toolName, tool.Version)
		}
	} else {
		fmt.Println("Tools:")
		for toolName, tool := range tools {
			fmt.Printf("  - %s @ %s\n", toolName, tool.Version)
		}
	}

	// List installed tools
	if verbose {
		fmt.Println("\n=== Installed tools ===")
		installed, err := client.ListTools(ctx)
		if err != nil {
			config.Warn("Failed to list installed tools: %v", err)
		} else {
			for _, t := range installed {
				fmt.Printf("  - %s\n", t)
			}
		}
	}

	return nil
}

func runStatus(ctx context.Context, cfg *config.Config, client *mise.Client, verbose bool) error {
	config.Info("=== Status ===")

	tools := config.GetTools(cfg)

	// Check which tools are managed by mise
	installed, err := client.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	installedMap := make(map[string]bool)
	for _, t := range installed {
		installedMap[t.Name] = true
	}

	fmt.Println("Tools:")
	for toolName, tool := range tools {
		version := tool.Version
		if version == "" {
			version = "latest"
		}

		status := "not installed"
		if installedMap[toolName] {
			status = "installed"
		}

		fmt.Printf("  %s @ %s [%s]\n", toolName, version, status)
	}

	// Show state directory info
	stateMgr := hooks.NewStateManager()
	fmt.Printf("\nState directory: %s\n", stateMgr.StateDir)

	return nil
}

func printHelp() {
	fmt.Print(`mise-seq - Tool installer with hooks

Usage:
  mise-seq [global-flags] <command> [command-flags] [args]

Commands:
  install    Install all tools from config (default)
  upgrade    Upgrade installed tools
  list       List installed tools
  status     Show status of configured tools

Global Flags:
  -c <file>     Config file (default: tools.yaml)
  --dry-run     Dry run mode
  --force-hooks Force hook execution
  --postinstall-on-update  Run postinstall on update
  -v            Verbose output
  --version     Show version

Examples:
  mise-seq install -c tools.yaml
  mise-seq upgrade
  mise-seq list
  mise-seq status
`)
}
