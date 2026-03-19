package config

import (
	"os"

	"gopkg.in/yaml.v3"

)

type DbConfig struct {
	Host        string	`yaml:"host"`
	Port        int	   	`yaml:"port"`
	User		string	`yaml:"user"`
	Password	string	`yaml:"password"`
	Name		string	`yaml:"name"`
}

type Config struct {
	DB DbConfig `yaml:"db"`
}

func LoadConfig(path string) (*Config, error) {
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