package config

// When defines when a hook should run
type When string

const (
	WhenInstall When = "install"
	WhenUpdate  When = "update"
	WhenAlways  When = "always"
)

// Config represents the unified configuration structure
type Config struct {
	Tools    []Tool    `json:"tools,omitempty" yaml:"tools,omitempty" toml:"tools,omitempty"`
	Defaults *Defaults `json:"defaults,omitempty" yaml:"defaults,omitempty" toml:"defaults,omitempty"`
	Settings *Settings `json:"settings,omitempty" yaml:"settings,omitempty" toml:"settings,omitempty"`
}

// Tool represents a single tool configuration
type Tool struct {
	Name        string   `json:"name" yaml:"name" toml:"name"`
	Version     string   `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`
	Package     string   `json:"package,omitempty" yaml:"package,omitempty" toml:"package,omitempty"`
	Plugin      string   `json:"plugin,omitempty" yaml:"plugin,omitempty" toml:"plugin,omitempty"`
	Exe         string   `json:"exe,omitempty" yaml:"exe,omitempty" toml:"exe,omitempty"`
	Disabled    bool     `json:"disabled,omitempty" yaml:"disabled,omitempty" toml:"disabled,omitempty"`
	Depends     []string `json:"depends,omitempty" yaml:"depends,omitempty" toml:"depends,omitempty"`
	Preinstall  []Hook   `json:"preinstall,omitempty" yaml:"preinstall,omitempty" toml:"preinstall,omitempty"`
	Postinstall []Hook   `json:"postinstall,omitempty" yaml:"postinstall,omitempty" toml:"postinstall,omitempty"`
}

// Hook represents a preinstall or postinstall hook
type Hook struct {
	Run         string `json:"run,omitempty" yaml:"run,omitempty" toml:"run,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty" toml:"description,omitempty"`
	When        []When `json:"when,omitempty" yaml:"when,omitempty" toml:"when,omitempty"`
}

// Defaults holds default hooks
type Defaults struct {
	Preinstall  []Hook `json:"preinstall,omitempty" yaml:"preinstall,omitempty" toml:"preinstall,omitempty"`
	Postinstall []Hook `json:"postinstall,omitempty" yaml:"postinstall,omitempty"`
}

// Settings holds mise settings
type Settings struct {
	NPM          NPM    `json:"npm,omitempty" yaml:"npm,omitempty" toml:"npm,omitempty"`
	Experimental string `json:"experimental,omitempty" yaml:"experimental,omitempty" toml:"experimental,omitempty"`
}

type NPM struct {
	PackageManager string `json:"package_manager,omitempty" yaml:"package_manager,omitempty" toml:"package_manager,omitempty"`
}

// HasDefaults checks if config has default hooks
func HasDefaults(cfg *Config) bool {
	if cfg == nil || cfg.Defaults == nil {
		return false
	}
	return len(cfg.Defaults.Preinstall) > 0 || len(cfg.Defaults.Postinstall) > 0
}

// GetDefaultsHooks returns the default hooks from config
func GetDefaultsHooks(cfg *Config) (preinstall, postinstall []Hook) {
	if cfg == nil || cfg.Defaults == nil {
		return nil, nil
	}
	return cfg.Defaults.Preinstall, cfg.Defaults.Postinstall
}

// MergeDefaults applies default hooks to a tool if not already defined
func (c *Config) MergeDefaults() {
	if c.Defaults == nil {
		return
	}

	if c.Tools == nil {
		return
	}

	for i := range c.Tools {
		tool := &c.Tools[i]
		// Merge preinstall hooks
		if len(tool.Preinstall) == 0 && len(c.Defaults.Preinstall) > 0 {
			tool.Preinstall = make([]Hook, len(c.Defaults.Preinstall))
			copy(tool.Preinstall, c.Defaults.Preinstall)
		}

		// Merge postinstall hooks
		if len(tool.Postinstall) == 0 && len(c.Defaults.Postinstall) > 0 {
			tool.Postinstall = make([]Hook, len(c.Defaults.Postinstall))
			copy(tool.Postinstall, c.Defaults.Postinstall)
		}
	}
}

// ValidateConfig validates the configuration
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return nil
	}

	// Validate dependencies (now in validation.go)
	if err := ValidateDependencies(cfg); err != nil {
		return err
	}

	return nil
}
