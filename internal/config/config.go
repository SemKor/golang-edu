package config

import ("os"
	"gopkg.in/yaml.v3"
)
type Config struct {
	Mode       string `yaml:"mode"`
	LogLevel   string `yaml:"logLevel"`
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	ClientName string `yaml:"clientName"`
}

func Init(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
