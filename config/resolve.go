package config

import (
	"fmt"
	"strings"
)

// ToolRef represents a tool with optional version
type ToolRef struct {
	Name    string
	Version string // empty = latest
}

// ParseToolRef parses "tool@version" string into ToolRef
func ParseToolRef(s string) ToolRef {
	if strings.Contains(s, "@") {
		parts := strings.SplitN(s, "@", 2)
		return ToolRef{
			Name:    parts[0],
			Version: parts[1],
		}
	}
	// No @version means @latest
	return ToolRef{
		Name:    s,
		Version: "latest",
	}
}

// ToolResolver resolves tool dependencies and determines installation order
type ToolResolver struct {
	tools map[string]Tool
}

// NewToolResolver creates a new tool resolver
func NewToolResolver(tools map[string]Tool) *ToolResolver {
	return &ToolResolver{
		tools: tools,
	}
}

// ResolveOrder returns the installation order based on dependencies
// Uses topological sort (Kahn's algorithm)
func (r *ToolResolver) ResolveOrder() ([]string, error) {
	if r.tools == nil {
		return nil, nil
	}

	// Build dependency graph
	// inDegree: number of dependencies each tool depends on
	// deps: map of tool -> its dependencies
	inDegree := make(map[string]int)
	deps := make(map[string][]string)

	for name := range r.tools {
		inDegree[name] = 0
		deps[name] = nil
	}

	// Process each tool's depends
	for name, tool := range r.tools {
		if len(tool.Depends) == 0 {
			continue
		}

		for _, dep := range tool.Depends {
			ref := ParseToolRef(dep)

			// Check if dependency exists in tools
			if _, exists := r.tools[ref.Name]; !exists {
				return nil, fmt.Errorf("tool '%s' depends on unknown tool '%s'", name, ref.Name)
			}

			// Add edge: ref.Name -> name (ref.Name must be installed before name)
			inDegree[name]++
		}

		// Store dependencies
		var depNames []string
		for _, dep := range tool.Depends {
			ref := ParseToolRef(dep)
			depNames = append(depNames, ref.Name)
		}
		deps[name] = depNames
	}

	// Kahn's algorithm for topological sort
	// Start with tools that have no dependencies
	queue := []string{}
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	result := []string{}
	for len(queue) > 0 {
		// Pop from queue
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Reduce in-degree for tools that depend on current
		for name, dependencies := range deps {
			for _, dep := range dependencies {
				if dep == current {
					inDegree[name]--
					if inDegree[name] == 0 {
						queue = append(queue, name)
					}
				}
			}
		}
	}

	// Check for cycles
	if len(result) != len(r.tools) {
		// Find tools in cycle
		var cycle []string
		for name, degree := range inDegree {
			if degree > 0 {
				cycle = append(cycle, name)
			}
		}
		return nil, fmt.Errorf("dependency cycle detected: %v", cycle)
	}

	return result, nil
}

// GetToolWithVersion returns the tool name with version (tool@version or tool@latest)
func GetToolWithVersion(name string, tool Tool) string {
	version := tool.Version
	if version == "" {
		version = "latest"
	}
	return name + "@" + version
}

// GetDependencies returns the dependencies of a tool as ToolRef slice
func (t *Tool) GetDependencies() []ToolRef {
	if t == nil || len(t.Depends) == 0 {
		return nil
	}

	result := make([]ToolRef, len(t.Depends))
	for i, dep := range t.Depends {
		result[i] = ParseToolRef(dep)
	}
	return result
}

// ValidateDependencies validates that all dependencies exist in tools
func ValidateDependencies(cfg *Config) error {
	if cfg == nil || cfg.Tools == nil {
		return nil
	}

	// Check that all dependencies reference existing tools
	for name, tool := range cfg.Tools {
		for _, dep := range tool.Depends {
			ref := ParseToolRef(dep)
			if _, exists := cfg.Tools[ref.Name]; !exists {
				return fmt.Errorf("tool '%s' depends on unknown tool '%s'", name, ref.Name)
			}
		}
	}

	return nil
}
