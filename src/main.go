package main

import (
	"fmt"
	"os"
	"slices"

	"package-graph/src/config"
	"package-graph/src/repository"
	"package-graph/src/tree"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadConfig("config.toml")
	if err != nil {
		return err
	}
	getDependencies, err := selectRepository(cfg)
	if err != nil {
		return err
	}

	graph, err := tree.BuildGraph(cfg.Package, getDependencies, cfg.IncludeExtra, cfg.TestMode, cfg.Filter)
	if err != nil {
		return err
	}

	root := cfg.Package
	if !cfg.TestMode {
		root = tree.NormalizeName(root)
	}
	if err := graph.Errors[root]; err != nil {
		return fmt.Errorf("пакет %s: %w", cfg.Package, err)
	}
	if _, exists := graph.Dependencies[root]; !exists {
		fmt.Printf("Пакет %s исключён фильтром\n", cfg.Package)
		return nil
	}

	if cfg.Reverse {
		printReverse(graph, root)
	} else {
		printGraph(graph, root, cfg.ASCIITree)
	}
	return nil
}

func printReverse(graph tree.Graph, root string) {
	var names []string
	for name := range graph.Dependencies {
		if name != root {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	if len(names) == 0 {
		fmt.Printf("У пакета %s нет обратных зависимостей\n", root)
		return
	}
	fmt.Printf("Обратные зависимости пакета %s:\n", root)
	for _, name := range names {
		fmt.Println(name)
	}
}

func printGraph(graph tree.Graph, root string, printTree bool) {
	fmt.Print(tree.PlantUML(graph))
	if printTree {
		fmt.Println()
		fmt.Print(tree.ASCIITree(graph, root))
	}
}

func selectRepository(cfg config.Config) (func(string) ([]string, error), error) {
	if !cfg.TestMode {
		get := func(name string) ([]string, error) {
			return repository.GetDependencies(cfg.Repository, name)
		}
		if !cfg.Reverse {
			return get, nil
		}
		if len(cfg.Candidates) == 0 {
			return nil, fmt.Errorf("для обратного поиска в PyPI укажите reverse_candidates")
		}
		names := append([]string{cfg.Package}, cfg.Candidates...)
		return tree.ReverseDependencies(names, get, cfg.IncludeExtra, false, cfg.Filter)
	}

	packages, err := repository.LoadTestRepository(cfg.Repository)
	if err != nil {
		return nil, err
	}
	get := func(name string) ([]string, error) {
		dependencies, exists := packages[name]
		if !exists {
			return nil, fmt.Errorf("пакет %s не найден", name)
		}
		return dependencies, nil
	}
	if !cfg.Reverse {
		return get, nil
	}
	var names []string
	for name := range packages {
		names = append(names, name)
	}
	return tree.ReverseDependencies(names, get, cfg.IncludeExtra, true, cfg.Filter)
}
