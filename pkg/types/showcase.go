package types

import (
	"encoding/json"
	"fmt"
	"os"
)

type ShowcaseExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
}

func (e ShowcaseExample) GetName() string        { return e.Name }
func (e ShowcaseExample) GetDescription() string { return e.Description }

func LoadExamples(path string) ([]ShowcaseExample, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read examples file: %w", err)
	}

	var result struct {
		Examples []ShowcaseExample `json:"examples"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse examples JSON: %w", err)
	}

	return result.Examples, nil
}
