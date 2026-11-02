package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config содержит настройки программы.
type Config struct {
	Package      string `toml:"package"`
	Repository   string `toml:"repository"`
	TestMode     bool   `toml:"test_mode"`
	ASCIITree    bool   `toml:"print_tree"`
	IncludeExtra bool   `toml:"include_extra"`
	Filter       string `toml:"filter"`
}

// LoadConfig читает и проверяет файл настроек.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("не удалось прочитать конфигурацию: %w", err)
	}

	var cfg Config
	decoder := toml.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		var unknown *toml.StrictMissingError
		if errors.As(err, &unknown) {
			key := strings.Join(unknown.Errors[0].Key(), ".")
			return Config{}, fmt.Errorf("неизвестный параметр TOML: %s: %w", key, err)
		}
		return Config{}, fmt.Errorf("ошибка разбора TOML: %w", err)
	}

	if err := ValidateConfig(cfg); err != nil {
		return Config{}, fmt.Errorf("ошибка конфигурации: %w", err)
	}

	return cfg, nil
}

func ValidateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.Package) == "" {
		return fmt.Errorf("package не может быть пустым")
	}

	if strings.TrimSpace(cfg.Repository) == "" {
		return fmt.Errorf("repository не может быть пустым")
	}

	if cfg.TestMode {
		file, err := os.Open(cfg.Repository)
		if err != nil {
			return fmt.Errorf("не удалось открыть тестовый файл: %w", err)
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			return fmt.Errorf("не удалось проверить тестовый файл: %w", err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("repository должен указывать на обычный файл")
		}
	} else {
		u, err := url.Parse(cfg.Repository)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
			return fmt.Errorf("repository должен быть URL с http или https")
		}
	}

	return nil
}
