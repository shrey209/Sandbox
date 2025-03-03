package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config struct
type Config struct {
	Mount []string `yaml:"mount"`
}

// ReadYML reads YAML file and returns Config struct
func ReadYML(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
