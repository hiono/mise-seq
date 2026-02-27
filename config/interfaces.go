package config

// ConfigValidator defines the interface for config validation
type ConfigValidator interface {
	Validate(cfg *Config) error
}

// DefaultValidator is the default implementation of ConfigValidator
type DefaultValidator struct{}

// Validate validates a config
func (v *DefaultValidator) Validate(cfg *Config) error {
	return ValidateConfig(cfg)
}
