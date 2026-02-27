package config

import (
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/yaml"
)

// ValidateYAMLWithSchema validates YAML data against the embedded CUE schema
func ValidateYAMLWithSchema(yamlData []byte, schemaCue string) (cue.Value, error) {
	ctx := cuecontext.New()

	// Compile the CUE schema from a string
	schemaInst := ctx.CompileString(schemaCue, cue.Filename("#Schema"))
	if schemaInst.Err() != nil {
		return cue.Value{}, fmt.Errorf("failed to compile CUE schema: %w", schemaInst.Err())
	}

	schemaValue := schemaInst.LookupPath(cue.ParsePath("#MiseSeqConfig"))
	if schemaValue.Err() != nil {
		return cue.Value{}, fmt.Errorf("failed to lookup schema: %w", schemaValue.Err())
	}

	// Extract YAML data into CUE format
	yamlFile, err := yaml.Extract("config.yaml", yamlData)
	if err != nil {
		return cue.Value{}, fmt.Errorf("YAML extraction error: %w", err)
	}

	// Build the extracted YAML data into a CUE instance
	yamlInst := ctx.BuildFile(yamlFile)
	if yamlInst.Err() != nil {
		return cue.Value{}, fmt.Errorf("failed to build YAML: %w", yamlInst.Err())
	}

	// Unify the CUE schema and the extracted YAML data
	unifiedValue := schemaValue.Unify(yamlInst)
	if err := unifiedValue.Validate(cue.Concrete(true)); err != nil {
		return cue.Value{}, fmt.Errorf("validation failed: %w", err)
	}

	return unifiedValue, nil
}

// ValidateCUEWithSchema validates CUE data against the embedded CUE schema
func ValidateCUEWithSchema(cueData []byte, schemaCue string) (cue.Value, error) {
	ctx := cuecontext.New()

	// Compile the CUE schema
	schemaInst := ctx.CompileString(schemaCue, cue.Filename("#Schema"))
	if schemaInst.Err() != nil {
		return cue.Value{}, fmt.Errorf("failed to compile CUE schema: %w", schemaInst.Err())
	}

	schemaValue := schemaInst.LookupPath(cue.ParsePath("#MiseSeqConfig"))
	if schemaValue.Err() != nil {
		return cue.Value{}, fmt.Errorf("failed to lookup schema: %w", schemaValue.Err())
	}

	// Compile the input CUE data
	cueInst := ctx.CompileString(string(cueData), cue.Filename("input.cue"))
	if cueInst.Err() != nil {
		return cue.Value{}, fmt.Errorf("failed to compile input CUE: %w", cueInst.Err())
	}

	// Unify and validate
	unifiedValue := schemaValue.Unify(cueInst)
	if err := unifiedValue.Validate(cue.Concrete(true)); err != nil {
		return cue.Value{}, fmt.Errorf("validation failed: %w", err)
	}

	return unifiedValue, nil
}

// ValidateFileWithSchema validates a config file against the embedded CUE schema
func ValidateFileWithSchema(filePath string, schemaCue string) (cue.Value, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return cue.Value{}, fmt.Errorf("failed to read file: %w", err)
	}

	// Detect format
	format, err := DetectFormat(filePath)
	if err != nil {
		return cue.Value{}, err
	}

	// Validate based on format
	return ValidateDataWithSchema(data, format, schemaCue)
}

// ValidateDataWithSchema validates data based on format string
func ValidateDataWithSchema(data []byte, format string, schemaCue string) (cue.Value, error) {
	switch strings.ToLower(format) {
	case "yaml", "yml":
		return ValidateYAMLWithSchema(data, schemaCue)
	case "cue":
		return ValidateCUEWithSchema(data, schemaCue)
	case "json", "toml":
		// JSON/TOML validation requires additional encoding
		// For now, just return nil (validation skipped)
		return cue.Value{}, nil
	default:
		return cue.Value{}, fmt.Errorf("unsupported format for validation: %s", format)
	}
}
