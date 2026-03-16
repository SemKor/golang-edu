package config

import (
	"github.com/go-yaml/yaml"
	"os"
)

type Config struct {
	N int `yaml:"n"`
}

func Init(path string) (*Config, error) {

	config := &Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
