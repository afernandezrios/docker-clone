package config

import (
	"encoding/json"
	"log"
	"os"
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

func Load(rootfs string) ImageConfigFile {
	byteValue, err := os.ReadFile(rootfs + "/config.json")
	if err != nil {
		log.Panicf("Error reading configuration file: %v", err)
	}

	var containerConfig ImageConfigFile
	json.Unmarshal(byteValue, &containerConfig)

	return containerConfig
}
