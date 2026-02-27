package config

import (
	_ "embed"
)

// SchemaCue is the embedded CUE schema definition
//
//go:embed schema.cue
var SchemaCue string
