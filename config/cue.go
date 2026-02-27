package config

import (
	"fmt"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/load"
)

// parseCUEFile parses a CUE configuration file into Config struct
func parseCUEFile(filename string) (*Config, error) {
	// Load the config file
	instances := load.Instances([]string{filename}, nil)
	if len(instances) == 0 {
		return nil, fmt.Errorf("failed to load CUE file: %s", filename)
	}

	bi := instances[0]
	if bi.Err != nil {
		return nil, fmt.Errorf("CUE load error: %v", bi.Err)
	}

	pkgs := cue.Build([]*build.Instance{bi})
	if len(pkgs) == 0 || pkgs[0] == nil {
		return nil, fmt.Errorf("failed to build CUE package")
	}

	cfg := pkgs[0].Value()

	// Decode into Config struct
	config := &Config{}
	if err := cfg.Decode(config); err != nil {
		return nil, fmt.Errorf("CUE decode error: %v", err)
	}

	// Merge defaults
	config.MergeDefaults()

	return config, nil
}

// DetectCUEFormat checks if a file is CUE format
func DetectCUEFormat(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".cue"
}
