package miseseq

#NonEmptyString: string & !=""

#When: "install" | "update" | "always"

#ToolKeyPattern: "^[A-Za-z0-9][A-Za-z0-9._+:/@-]*$"

#Version: #NonEmptyString

#Hook: close({
  run:          #NonEmptyString
  when?:        [...#When]
  description?: #NonEmptyString
})

#HookList: [...#Hook]

#Defaults: close({
  preinstall?:  #HookList
  postinstall?: #HookList
})

#ToolConfig: close({
  version:      #Version
  exe?:         #NonEmptyString
  preinstall?:  #HookList
  postinstall?: #HookList
})

#MiseSeqConfig: close({
  defaults?:    #Defaults
  tools_order?: [...string & =~#ToolKeyPattern]

  tools: {
    [=~#ToolKeyPattern]: #ToolConfig
  }

  if tools_order != _|_ {
    for t in tools_order {
      tools[t]: _,
    }
  }
})
