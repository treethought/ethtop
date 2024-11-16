package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
  RPC struct {
    HTTP string `yaml:"http"`
    WS   string `yaml:"ws"`
  }
	Log      struct {
		Path string `yaml:"path"`
	} `yaml:"log"`
}

func ReadConfig(path string) (*Config, error) {
	var c Config
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal([]byte(data), &c); err != nil {
		return nil, err
	}
	return &c, nil
}
