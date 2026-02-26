package config

// Config represents the unified configuration structure
type Config struct {
	ToolsOrder []string          `json:"tools_order,omitempty" yaml:"tools_order,omitempty"`
	Tools      map[string]Tool   `json:"tools,omitempty" yaml:"tools,omitempty"`
	Defaults   *Defaults         `json:"defaults,omitempty" yaml:"defaults,omitempty"`
	Settings   *Settings         `json:"settings,omitempty" yaml:"settings,omitempty"`
}

// Tool represents a single tool configuration
type Tool struct {
	Version     string   `json:"version,omitempty"`
	Exe         string   `json:"exe,omitempty"`
	Preinstall  []Hook  `json:"preinstall,omitempty"`
	Postinstall []Hook  `json:"postinstall,omitempty"`
}

// Hook represents a preinstall or postinstall hook
type Hook struct {
	When        []string `json:"when,omitempty"`
	Description string   `json:"description,omitempty"`
	Run         string   `json:"run,omitempty"`
}

// Defaults holds default hooks
type Defaults struct {
	Preinstall  []Hook `json:"preinstall,omitempty"`
	Postinstall []Hook `json:"postinstall,omitempty"`
}

// Settings holds mise settings
type Settings struct {
	NPM NPM `json:"npm,omitempty"`
}

type NPM struct {
	PackageManager string `json:"package_manager,omitempty"`
}
