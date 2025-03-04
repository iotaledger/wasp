package main

import (
	"sort"
	"strings"
)

type TypeGroup struct {
	BaseTypes    []TypeDefinition // Core types (not inputs/outputs)
	DerivedTypes []TypeDefinition // Related input/output types
	Dependencies map[string]bool  // Dependencies on other groups
}

// GetOutput Returns the generated types using a Graph to properly order all types (So types don't include types below them)
func (tg *TypeGenerator) GetOutput() string {
	// 1. Group related types
	groups := tg.groupTypes()

	// 2. Sort groups by dependencies
	sortedGroups := tg.sortGroups(groups)

	// 3. Generate output
	return tg.generateOutput(sortedGroups, groups)
}

func (tg *TypeGenerator) groupTypes() map[string]*TypeGroup {
	groups := make(map[string]*TypeGroup)

	// Helper to get base group name
	getBaseGroup := func(name string) string {
		base := strings.TrimSuffix(strings.TrimSuffix(name, "Inputs"), "Outputs")
		if base != name {
			return base
		}
		return name
	}

	// First pass: create groups and assign types
	for _, typeDef := range tg.output {
		groupName := getBaseGroup(typeDef.Name)

		if groups[groupName] == nil {
			groups[groupName] = &TypeGroup{
				Dependencies: make(map[string]bool),
			}
		}

		// Add to appropriate slice based on whether it's a base type
		if typeDef.Name == groupName {
			groups[groupName].BaseTypes = append(groups[groupName].BaseTypes, typeDef)
		} else {
			groups[groupName].DerivedTypes = append(groups[groupName].DerivedTypes, typeDef)
		}
	}

	// Second pass: collect dependencies
	for groupName, group := range groups {
		for _, typeDef := range append(group.BaseTypes, group.DerivedTypes...) {
			for _, dep := range typeDef.Dependencies {
				depGroup := getBaseGroup(dep)
				if depGroup != groupName {
					group.Dependencies[depGroup] = true
				}
			}
		}
	}

	return groups
}

func (tg *TypeGenerator) sortGroups(groups map[string]*TypeGroup) []string {
	var sorted []string
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var visit func(name string)
	visit = func(name string) {
		if visiting[name] {
			return // Handle cycles
		}
		if visited[name] {
			return
		}

		visiting[name] = true
		for dep := range groups[name].Dependencies {
			visit(dep)
		}

		visited[name] = true
		delete(visiting, name)
		sorted = append(sorted, name)
	}

	// Visit all groups
	for name := range groups {
		if !visited[name] {
			visit(name)
		}
	}

	return sorted
}

func (tg *TypeGenerator) generateOutput(sortedGroups []string, groups map[string]*TypeGroup) string {
	var result []string

	importString := "import { bcs, fromHex, toHex } from '@mysten/bcs';\n\n"
	result = append(result, importString)

	for _, groupName := range sortedGroups {
		group := groups[groupName]

		// Add base types first
		for _, typeDef := range group.BaseTypes {
			result = append(result, typeDef.Definition)
		}

		// Then add derived types (inputs/outputs)
		sort.Slice(group.DerivedTypes, func(i, j int) bool {
			return group.DerivedTypes[i].Name < group.DerivedTypes[j].Name
		})
		for _, typeDef := range group.DerivedTypes {
			result = append(result, typeDef.Definition)
		}
	}

	return strings.Join(result, "\n\n")
}
