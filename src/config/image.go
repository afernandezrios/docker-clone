package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ImageConfigFile struct {
	Config ImageConfig `json:"config"`
}

type ImageConfig struct {
	Env        []string            `json:"Env"`
	Entrypoint []string            `json:"Entrypoint"`
	Volumes    map[string]struct{} `json:"Volumes"`
	WorkingDir string              `json:"WorkingDir"`
}

func Load(rootfs string) (*ImageConfig, error) {
	configPath := filepath.Join(rootfs, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration file in %s: %w", configPath, err)
	}

	var raw ImageConfigFile
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("Error parsing image config: %w", err)
	}

	return &raw.Config, nil
}
