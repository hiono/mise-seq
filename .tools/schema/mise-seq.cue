package miseseq

// Non-empty string constraint
#NonEmptyString: string & !""

// Hook trigger timing
#When: "install" | "update" | "always"

// Tool key pattern (tool name)
#ToolKeyPattern: "^[A-Za-z0-9][A-Za-z0-9._+:/@-]*$"

// Version can be string or float, empty = latest
#Version: string | float | *"" | "latest"

// Hook definition
#Hook: close({
  run:          #NonEmptyString
  when?:        [...#When]
  description?: #NonEmptyString
})

#HookList: [...#Hook]

// Default hooks applied to all tools
#Defaults: close({
  preinstall?:  #HookList
  postinstall?: #HookList
})

// Tool configuration
// All fields are optional - defaults are:
//   version: "latest"
//   exe: <tool name>
//   depends: []
#ToolConfig: close({
  version?:     #Version
  exe?:         #NonEmptyString
  depends?:     [...#NonEmptyString]
  preinstall?:  #HookList
  postinstall?: #HookList
})

// Main configuration
#MiseSeqConfig: {
  defaults?:    #Defaults
  tools_order?: [...string]

  tools: {
    [string]: #ToolConfig
  }
}
}
