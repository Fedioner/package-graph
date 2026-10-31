package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"package-graph/src/config"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{"правильный конфиг", "package = 'requests'\nrepository = 'https://pypi.org'", ""},
		{"пустое имя", "package = ''\nrepository = 'https://pypi.org'", "package"},
		{"неверный адрес", "package = 'requests'\nrepository = 'не адрес'", "repository"},
		{"неверный test_mode", "package = 'A'\nrepository = 'https://pypi.org'\ntest_mode = 'yes'", "TestMode"},
		{"неверный print_tree", "package = 'A'\nrepository = 'https://pypi.org'\nprint_tree = 123", "ASCIITree"},
		{"неверный filter", "package = 'A'\nrepository = 'https://pypi.org'\nfilter = 123", "Filter"},
		{"неизвестный параметр", "package = 'A'\nrepository = 'https://pypi.org'\nprint_tre = true", "print_tre"},
		{"отсутствует тестовый файл", "package = 'A'\nrepository = 'missing.txt'\ntest_mode = true", "тестовый файл"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.toml")
			if err := os.WriteFile(path, []byte(test.content), 0600); err != nil {
				t.Fatal(err)
			}

			_, err := config.LoadConfig(path)
			if test.wantErr == "" && err != nil {
				t.Errorf("неожиданная ошибка: %v", err)
			}
			if test.wantErr != "" && (err == nil || !strings.Contains(err.Error(), test.wantErr)) {
				t.Errorf("ожидалась ошибка с %q, получено: %v", test.wantErr, err)
			}
		})
	}
}

func TestMissingConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.toml")
	if _, err := config.LoadConfig(path); err == nil {
		t.Error("ожидалась ошибка отсутствующего файла")
	}
}

func TestDefaultsAndTestFile(t *testing.T) {
	dir := t.TempDir()
	repository := filepath.Join(dir, "repository.txt")
	if err := os.WriteFile(repository, []byte("A: B\n"), 0600); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "config.toml")
	content := "package = 'A'\nrepository = '" + repository + "'\ntest_mode = true"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.TestMode || cfg.ASCIITree || cfg.Filter != "" {
		t.Errorf("неверные значения параметров: %+v", cfg)
	}
}
