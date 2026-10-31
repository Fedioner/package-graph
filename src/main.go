package main

import (
	"fmt"
	"os"

	"package-graph/src/config"
)

func main() {
	cfg, err := config.LoadConfig("config.toml")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("package = %s\n", cfg.Package)
	fmt.Printf("repository = %s\n", cfg.Repository)
	fmt.Printf("test_mode = %t\n", cfg.TestMode)
	fmt.Printf("print_tree = %t\n", cfg.ASCIITree)
	fmt.Printf("filter = %s\n", cfg.Filter)
}
