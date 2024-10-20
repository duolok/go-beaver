package config

import (
    "fmt"
    "os"
	yaml "gopkg.in/yaml.v3"
)

type Task struct {
	File     string `yaml:"file"`
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Schedule Schedule `yaml:"schedule"`
}

type Schedule struct {
	Every int    `yaml:"every"`
	Unit  string `yaml:"unit"`
}

type TaskConfig struct {
	Tasks []Task `yaml:"tasks"`
}

func LoadConfig(path string) (TaskConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return TaskConfig{}, fmt.Errorf("Unable to parse: %w", err)
	}

	var config TaskConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return TaskConfig{}, fmt.Errorf("Unable to parse config: %w", err)
	}

	return config, nil
}
