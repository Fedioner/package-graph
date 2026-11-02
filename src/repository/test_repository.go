package repository

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var testNamePattern = regexp.MustCompile(`^[A-Z]+$`)

// LoadTestRepository читает и проверяет файл тестового репозитория.
func LoadTestRepository(path string) (map[string][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть тестовый файл: %w", err)
	}
	defer file.Close()

	packages := make(map[string][]string)
	lines := make(map[string]int)
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++

		name, dependencies, err := parseTestLine(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("строка %d: %w", lineNumber, err)
		}

		if _, exists := packages[name]; exists {
			return nil, fmt.Errorf("строка %d: пакет %s уже объявлен", lineNumber, name)
		}

		packages[name] = dependencies
		lines[name] = lineNumber
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения тестового файла: %w", err)
	}
	if len(packages) == 0 {
		return nil, fmt.Errorf("тестовый файл не содержит пакетов")
	}
	if err := checkTestDependencies(packages, lines); err != nil {
		return nil, err
	}

	return packages, nil
}

func parseTestLine(line string) (string, []string, error) {
	if strings.TrimSpace(line) == "" {
		return "", nil, fmt.Errorf("пустые строки запрещены")
	}

	name, rest, found := strings.Cut(line, ":")
	if !found || strings.Contains(rest, ":") {
		return "", nil, fmt.Errorf("ожидался формат: A: B C")
	}

	name = strings.TrimSpace(name)
	if !testNamePattern.MatchString(name) {
		return "", nil, fmt.Errorf("неверное имя пакета: %q", name)
	}

	dependencies := strings.Fields(rest)
	for _, dependency := range dependencies {
		if !testNamePattern.MatchString(dependency) {
			return "", nil, fmt.Errorf("неверное имя зависимости: %q", dependency)
		}
	}

	return name, dependencies, nil
}

func checkTestDependencies(packages map[string][]string, lines map[string]int) error {
	for name, dependencies := range packages {
		for _, dependency := range dependencies {
			if _, exists := packages[dependency]; !exists {
				return fmt.Errorf(
					"строка %d: зависимость %s не объявлена",
					lines[name],
					dependency,
				)
			}
		}
	}

	return nil
}
