package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type packageInfo struct {
	Info struct {
		RequiresDist []string `json:"requires_dist"`
	} `json:"info"`
}

// GetDependencies получает прямые зависимости пакета из PyPI.
func GetDependencies(repositoryURL, packageName string) ([]string, error) {
	repositoryURL = strings.TrimRight(repositoryURL, "/")
	url := fmt.Sprintf("%s/pypi/%s/json", repositoryURL, packageName)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос к PyPI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("не удалось получить данные о пакете: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать ответ от PyPI: %w", err)
	}

	return parseResponse(data)
}

func parseResponse(data []byte) ([]string, error) {
	var result packageInfo

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа PyPI: %w", err)
	}

	return result.Info.RequiresDist, nil
}
