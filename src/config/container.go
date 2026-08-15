package config

import (
	"encoding/json"
	"log"
	"os"
)

type ContainerConfig struct {
	Config ConfigData `json:"config"`
}

type ConfigData struct {
	Env        []string `json:"Env"`
	Entrypoint []string `json:"Entrypoint"`
	Volumes    []string `json:"Volumes"`
	WorkingDir string   `json:"WorkingDir"`
}

func GetContainer(rootfs string) ContainerConfig {
	byteValue, err := os.ReadFile(rootfs + "/config.json")
	if err != nil {
		log.Panicf("Error reading configuration file: %v", err)
	}

	var containerConfig ContainerConfig
	json.Unmarshal(byteValue, &containerConfig)

	return containerConfig
}
