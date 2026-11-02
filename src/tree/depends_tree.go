package tree

import (
	"fmt"
	"regexp"
	"strings"
)

// Node описывает зависимость пакета и её условия.
type Node struct {
	PackageName  string
	VersionLimit string
	Children     []*Node
}

var packageNamePattern = regexp.MustCompile(
	`^[A-Za-z0-9]([A-Za-z0-9._-]*[A-Za-z0-9])?`,
)

var separatorPattern = regexp.MustCompile(`[-_.]+`)
var quotedPattern = regexp.MustCompile(`"[^"]*"|'[^']*'`)
var extraPattern = regexp.MustCompile(`\bextra\b`)

func NormalizeName(name string) string {
	return separatorPattern.ReplaceAllString(strings.ToLower(name), "-")
}

func hasExtraMarker(dependence string) bool {
	_, marker, found := strings.Cut(dependence, ";")
	if !found {
		return false
	}

	// нужно удалить все в кавычках, чтобы не перепутать с переменной
	return extraPattern.MatchString(quotedPattern.ReplaceAllString(marker, ""))
}

func ParseNodes(dependence string) (Node, error) {
	dependence = strings.TrimSpace(dependence)

	name := packageNamePattern.FindString(dependence)
	if name == "" {
		return Node{}, fmt.Errorf("не удалось определить имя пакета: %q", dependence)
	}

	versionLimit := strings.TrimSpace(dependence[len(name):])

	node := Node{
		PackageName:  name,
		VersionLimit: versionLimit,
		Children:     nil,
	}

	return node, nil
}

// Graph содержит зависимости пакетов и ошибки их получения.
type Graph struct {
	Dependencies map[string][]Node
	Errors       map[string]error
}

// Ограничим число одновременных запросов, банят
const maxConcurrentRequests = 5

type result struct {
	name         string
	dependencies []string
	err          error
}

func fetchLevel(packages []string, getDependencies func(string) ([]string, error)) []result {
	results := make([]result, 0, len(packages))
	for start := 0; start < len(packages); start += maxConcurrentRequests {
		end := min(start+maxConcurrentRequests, len(packages))
		responses := make(chan result, end-start)

		for _, name := range packages[start:end] {
			go func(name string) {
				dependencies, err := getDependencies(name)
				responses <- result{name, dependencies, err}
			}(name)
		}
		for range packages[start:end] {
			results = append(results, <-responses)
		}
	}
	return results
}

type graphOptions struct {
	includeExtra bool
	testMode     bool
	filter       string
}

// BuildGraph строит граф зависимостей обходом в ширину.
func BuildGraph(
	root string,
	getDependencies func(string) ([]string, error),
	includeExtra, testMode bool,
	filter string,
) (Graph, error) {
	// Граф будет списком смежности
	graph := Graph{
		Dependencies: make(map[string][]Node),
		Errors:       make(map[string]error),
	}
	if !testMode {
		root = NormalizeName(root)
	}
	if filter != "" && strings.Contains(root, filter) {
		return graph, nil
	}
	queue := []string{root}
	visited := map[string]bool{
		root: true,
	}
	options := graphOptions{includeExtra, testMode, filter}
	for len(queue) > 0 {
		var nextQueue []string
		for _, response := range fetchLevel(queue, getDependencies) {
			// получили все прямые зависимости очередного пакета
			if response.err != nil {
				graph.Errors[response.name] = response.err
				continue
			}
			found, err := addDependencies(&graph, response, visited, options)
			if err != nil {
				return Graph{}, err
			}
			nextQueue = append(nextQueue, found...)
		}
		queue = nextQueue
	}
	return graph, nil
}

func addDependencies(
	graph *Graph, 
	response result, 
	visited map[string]bool,
	options graphOptions,
) ([]string, error) {

	// инициализируем ноду в графе
	graph.Dependencies[response.name] = nil
	var next []string
	// надо пройтись по всему респонсу и добавить в граф
	for _, dependence := range response.dependencies {
		if !options.includeExtra && hasExtraMarker(dependence) {
			continue
		}
		node, err := ParseNodes(dependence)
		if err != nil {
			return nil, err
		}
		if !options.testMode {
			node.PackageName = NormalizeName(node.PackageName)
		}
		if options.filter != "" && strings.Contains(node.PackageName, options.filter) {
			continue
		}
		graph.Dependencies[response.name] = append(graph.Dependencies[response.name], node)
		// Проверка и доавбление в список обработанных
		if !visited[node.PackageName] {
			visited[node.PackageName] = true
			next = append(next, node.PackageName)
		}
	}
	return next, nil
}
