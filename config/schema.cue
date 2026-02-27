package miseseq

// Version can be string, float, or empty (defaults to latest)
#Version: string | float | *"" | "latest"

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

// Tool configuration
// All fields are optional - defaults are:
//   version: "latest"
//   exe: <tool name>
//   depends: []
#ToolConfig: {
  version?:    #Version
  exe?:        string
  depends?:    [...string]
  preinstall?:  #HookList
  postinstall?: #HookList
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

// Main configuration
#MiseSeqConfig: {
  defaults?:    #Defaults
  tools_order?: [...string]

  tools: {
    [string]: #ToolConfig
  }

  settings?: #Settings
}
