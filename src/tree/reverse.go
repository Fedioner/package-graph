package tree

import (
	"fmt"
	"slices"
	"strings"
)

// ReverseDependencies индексирует входящие рёбра для обхода через BuildGraph.
func ReverseDependencies(
	names []string,
	get func(string) ([]string, error),
	includeExtra, testMode bool,
	filter string,
) (func(string) ([]string, error), error) {
	packages := reversePackages(names, testMode, filter)
	index := make(map[string][]string, len(packages))
	for _, name := range packages {
		index[name] = nil
	}
	for _, response := range fetchLevel(packages, get) {
		if response.err != nil {
			return nil, fmt.Errorf("пакет %s: %w", response.name, response.err)
		}
		for _, requirement := range response.dependencies {
			if !includeExtra && hasExtraMarker(requirement) {
				continue
			}
			node, err := ParseNodes(requirement)
			if err != nil {
				return nil, fmt.Errorf("пакет %s: %w", response.name, err)
			}
			name := packageKey(node.PackageName, testMode)
			if _, exists := index[name]; exists {
				index[name] = append(index[name], response.name)
			}
		}
	}
	return func(name string) ([]string, error) {
		dependencies, exists := index[name]
		if !exists {
			return nil, fmt.Errorf("пакет %s не найден", name)
		}
		return dependencies, nil
	}, nil
}

func reversePackages(names []string, testMode bool, filter string) []string {
	packages := make([]string, 0, len(names))
	for _, name := range names {
		name = packageKey(name, testMode)
		if filter == "" || !strings.Contains(name, filter) {
			packages = append(packages, name)
		}
	}
	slices.Sort(packages)
	return slices.Compact(packages)
}

func packageKey(name string, testMode bool) string {
	if testMode {
		return name
	}
	return NormalizeName(name)
}
