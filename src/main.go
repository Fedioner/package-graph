package main

import (
	"fmt"
	"os"

	"package-graph/src/config"
	"package-graph/src/repository"
	"package-graph/src/tree"
)

func main() {
	cfg, err := config.LoadConfig("config.toml")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	getDependencies, err := loadRepository(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	graph, err := tree.BuildGraph(cfg.Package, getDependencies, cfg.IncludeExtra, cfg.TestMode, cfg.Filter)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	root := cfg.Package
	if !cfg.TestMode {
		root = tree.NormalizeName(root)
	}
	if err := graph.Errors[root]; err != nil {
		fmt.Fprintf(os.Stderr, "пакет %s: %v\n", cfg.Package, err)
		return
	}
	if _, exists := graph.Dependencies[root]; !exists {
		fmt.Printf("Пакет %s исключён фильтром\n", cfg.Package)
		return
	}

	printGraph(graph)
}

func printGraph(graph tree.Graph) {
	for packageName, dependencies := range graph.Dependencies {
		for _, dependency := range dependencies {
			if err := graph.Errors[dependency.PackageName]; err != nil {
				fmt.Printf("%s → ошибка пакета %s: %v\n", packageName, dependency.PackageName, err)
				continue
			}
			fmt.Printf("%s → %s%s\n",
				packageName,
				dependency.PackageName,
				dependency.VersionLimit,
			)
		}
	}
}

func loadRepository(cfg config.Config) (func(string) ([]string, error), error) {
	if !cfg.TestMode {
		return func(name string) ([]string, error) {
			return repository.GetDependencies(cfg.Repository, name)
		}, nil
	}

	packages, err := repository.LoadTestRepository(cfg.Repository)
	if err != nil {
		return nil, err
	}
	return func(name string) ([]string, error) {
		dependencies, exists := packages[name]
		if !exists {
			return nil, fmt.Errorf("пакет %s не найден", name)
		}
		return dependencies, nil
	}, nil
}
