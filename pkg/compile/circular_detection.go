package compile

import (
	"fmt"
	"strings"

	"github.com/kalo-build/morphe-go/pkg/yaml"
	"github.com/kalo-build/morphe-go/pkg/yamlops"
)

// CircularDependency represents a circular dependency between models
type CircularDependency struct {
	Path []string
}

// String returns a human-readable representation of the circular dependency
func (c CircularDependency) String() string {
	return fmt.Sprintf("Circular dependency detected: %s", strings.Join(c.Path, " -> "))
}

// DetectCircularDependencies checks for circular relationships in models
func DetectCircularDependencies(models map[string]yaml.Model) []CircularDependency {
	var cycles []CircularDependency
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	// Build adjacency list for the dependency graph
	graph := buildDependencyGraph(models)

	// Run DFS from each node
	for modelName := range models {
		if !visited[modelName] {
			if foundCycles := dfsDetectCycles(modelName, graph, visited, recStack, path); len(foundCycles) > 0 {
				cycles = append(cycles, foundCycles...)
			}
		}
	}

	return dedupeCycles(cycles)
}

// buildDependencyGraph creates an adjacency list of model dependencies
func buildDependencyGraph(models map[string]yaml.Model) map[string][]string {
	graph := make(map[string][]string)

	for modelName, model := range models {
		deps := []string{}

		// Check all relationships
		for relName, relation := range model.Related {
			// Get the target model name
			targetName := yamlops.GetRelationTargetName(relName, relation.Aliased)

			// For polymorphic relationships, add all possible targets
			if len(relation.For) > 0 {
				for _, forModel := range relation.For {
					if forModel != modelName && !contains(deps, forModel) {
						deps = append(deps, forModel)
					}
				}
			} else if targetName != "" && targetName != modelName {
				// Regular relationship
				if !contains(deps, targetName) {
					deps = append(deps, targetName)
				}
			}
		}

		graph[modelName] = deps
	}

	return graph
}

// dfsDetectCycles performs DFS to detect cycles
func dfsDetectCycles(node string, graph map[string][]string, visited, recStack map[string]bool, path []string) []CircularDependency {
	visited[node] = true
	recStack[node] = true
	path = append(path, node)

	var cycles []CircularDependency

	// Visit all neighbors
	for _, neighbor := range graph[node] {
		if !visited[neighbor] {
			if foundCycles := dfsDetectCycles(neighbor, graph, visited, recStack, path); len(foundCycles) > 0 {
				cycles = append(cycles, foundCycles...)
			}
		} else if recStack[neighbor] {
			// Found a cycle - extract the cycle path
			cycleStart := -1
			for i, n := range path {
				if n == neighbor {
					cycleStart = i
					break
				}
			}

			if cycleStart >= 0 {
				cyclePath := append([]string{}, path[cycleStart:]...)
				cyclePath = append(cyclePath, neighbor) // Complete the cycle
				cycles = append(cycles, CircularDependency{Path: cyclePath})
			}
		}
	}

	recStack[node] = false
	return cycles
}

// dedupeCycles removes duplicate cycles (same nodes in different order)
func dedupeCycles(cycles []CircularDependency) []CircularDependency {
	seen := make(map[string]bool)
	unique := []CircularDependency{}

	for _, cycle := range cycles {
		// Create a normalized representation
		normalized := normalizeCycle(cycle.Path)
		key := strings.Join(normalized, ",")

		if !seen[key] {
			seen[key] = true
			unique = append(unique, cycle)
		}
	}

	return unique
}

// normalizeCycle rotates the cycle so it starts with the lexicographically smallest node
func normalizeCycle(path []string) []string {
	if len(path) <= 1 {
		return path
	}

	// Remove the duplicate last element if present
	if path[0] == path[len(path)-1] {
		path = path[:len(path)-1]
	}

	// Find the smallest element
	minIdx := 0
	for i, node := range path {
		if node < path[minIdx] {
			minIdx = i
		}
	}

	// Rotate to start with the smallest element
	normalized := make([]string, len(path))
	for i := range path {
		normalized[i] = path[(minIdx+i)%len(path)]
	}

	return normalized
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
