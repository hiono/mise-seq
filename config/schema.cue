package miseseq

// Version can be string, float, or empty (defaults to latest)
#Version: string | float | *"" | "latest"

// Package format: "runtime:tool" (no @version allowed)
// Examples: "npm:difit", "go:github.com/user/repo", "npm:@scope/package"
#Package: string

// Hook trigger timing
#When: "install" | "update" | "always"

// Hook definition
#Hook: {
  run:          string
  when?:        [...#When]
  description?: string
}

#HookList: [...#Hook]

// Default hooks applied to all tools
#Defaults: {
  preinstall?:  #HookList
  postinstall?: #HookList
}

// Tool configuration - name is REQUIRED
#ToolConfig: {
  name:        string                      // REQUIRED - tool name
  version?:    #Version                   // defaults to "latest"
  package?:    #Package                   // runtime:tool format (e.g., npm:difit)
  plugin?:     string                      // Mise plugin to use
  exe?:        string | *name             // defaults to name
  disabled?:   bool | *false              // defaults to false
  depends?:    [...string]                // List of tool names (NOT "name@version")
  preinstall?:  #HookList                  // Tool-specific preinstall hooks
  postinstall?: #HookList                  // Tool-specific postinstall hooks
}

// NPM settings
#NPM: {
  package_manager?: string
}

// Settings for mise
#Settings: {
  npm?:         #NPM
  experimental?: string
}

// Main configuration - tools is now an ORDERED LIST
#MiseSeqConfig: {
  defaults?:    #Defaults
  tools: [...#ToolConfig]                  // ORDERED LIST (not map!)
  settings?: #Settings
}
