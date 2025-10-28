package types

import (
	"encoding/json"
	"fmt"
	"os"
)

type Example struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
}

func (e Example) GetName() string        { return e.Name }
func (e Example) GetDescription() string { return e.Description }

func LoadExamples(path string) ([]Example, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read examples file: %w", err)
	}

	var result struct {
		Examples []Example `json:"examples"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse examples JSON: %w", err)
	}

	return result.Examples, nil
}
