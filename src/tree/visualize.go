package tree

import (
	"fmt"
	"slices"
	"strings"
)

// PlantUML описывает пакеты и рёбра графа, включая циклы.
func PlantUML(graph Graph) string {
	var out strings.Builder
	out.WriteString("@startuml\n")
	names := graphNames(graph)
	aliases := make(map[string]string, len(names))
	for i, name := range names {
		alias := fmt.Sprintf("p%d", i)
		aliases[name] = alias
		label := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;").Replace(name)
		status := ""
		if graph.Errors[name] != nil {
			status = " <<unavailable>>"
		}
		fmt.Fprintf(&out, "component \"%s\" as %s%s\n", label, alias, status)
	}
	for _, name := range names {
		for _, node := range graph.Dependencies[name] {
			fmt.Fprintf(&out, "%s --> %s\n", aliases[name], aliases[node.PackageName])
		}
	}
	out.WriteString("@enduml\n")
	return out.String()
}

func graphNames(graph Graph) []string {
	names := make(map[string]bool)
	for name, dependencies := range graph.Dependencies {
		names[name] = true
		for _, node := range dependencies {
			names[node.PackageName] = true
		}
	}
	for name := range graph.Errors {
		names[name] = true
	}
	var sorted []string
	for name := range names {
		sorted = append(sorted, name)
	}
	slices.Sort(sorted)
	return sorted
}

type treeFrame struct {
	name   string
	prefix string
	index  int
}

// ASCIITree печатает дерево с пометками циклов и общих ветвей.
func ASCIITree(graph Graph, root string) string {
	var out strings.Builder
	fmt.Fprintln(&out, root)
	stack := []treeFrame{{name: root}}
	seen := map[string]bool{root: true}
	active := map[string]bool{root: true}
	for len(stack) > 0 {
		top := &stack[len(stack)-1]
		children := graph.Dependencies[top.name]
		if top.index == len(children) {
			delete(active, top.name)
			stack = stack[:len(stack)-1]
			continue
		}
		node := children[top.index]
		top.index++
		branch, indent := "|-- ", "|   "
		if top.index == len(children) {
			branch, indent = "`-- ", "    "
		}
		name := node.PackageName
		note := treeNote(graph.Errors[name], active[name], seen[name])
		fmt.Fprintf(&out, "%s%s%s%s%s\n", top.prefix, branch, name, node.VersionLimit, note)
		if note == "" {
			seen[name], active[name] = true, true
			stack = append(stack, treeFrame{name: name, prefix: top.prefix + indent})
		}
	}
	return out.String()
}

func treeNote(err error, active, seen bool) string {
	switch {
	case err != nil:
		return " [ошибка: " + err.Error() + "]"
	case active:
		return " [цикл]"
	case seen:
		return " [уже показан]"
	default:
		return ""
	}
}
