package main

import (
	"fmt"
	"os"
	"package-graph/src/config"
	"package-graph/src/repository"
)

func main() {
	cfg, err := config.LoadConfig("config.toml")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if cfg.TestMode {
		fmt.Println("Тестовый режим еще в разработке")
		os.Exit(1)
	}

	dependencies, err := repository.GetDependencies(cfg.Repository, cfg.Package)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Прямые зависимотсти пакета %s:\n", cfg.Package)
	if len(dependencies) == 0 {
		fmt.Println("Прямых зависимостей нет")
	}

	for _, st := range dependencies {
		fmt.Println(st)
	}
}
