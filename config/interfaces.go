package config

// ConfigLoader defines the interface for config loading
type ConfigLoader interface {
	Parse(path string) (*Config, error)
}

// FormatParser defines the interface for format-specific parsing
type FormatParser interface {
	Parse(path string) (*Config, error)
}

// JSONParser implements FormatParser for JSON
type JSONParser struct{}

// Parse implements FormatParser
func (p *JSONParser) Parse(path string) (*Config, error) {
	return ParseJSON(path)
}

// YAMLParser implements FormatParser for YAML
type YAMLParser struct{}

// Parse implements FormatParser
func (p *YAMLParser) Parse(path string) (*Config, error) {
	return ParseYAML(path)
}

// TOMLParser implements FormatParser for TOML
type TOMLParser struct{}

// Parse implements FormatParser
func (p *TOMLParser) Parse(path string) (*Config, error) {
	return ParseTOML(path)
}

// CUEParser implements FormatParser for CUE
type CUEParser struct{}

// Parse implements FormatParser
func (p *CUEParser) Parse(path string) (*Config, error) {
	return ParseCUE(path)
}

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
