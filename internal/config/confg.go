package config

type Config struct {
	Mode       string `yaml:"mode"`
	LogLevel   string `yaml:"logLevel"`
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	ClientName string `yaml:"clientName"`
}

func Init() (*Config, error) {
	// TODO implement config initialization func
	panic("not implemented")
}
