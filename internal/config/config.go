package config

import (
	"flag"
	"fmt"

	"github.com/spf13/viper"
)

// nolint: gochecknoglobals
var config *Config

type Config struct {
	Mode     string     `yaml:"mode"`
	Server   AuthServer `yaml:"server"`
	Log      Log        `yaml:"log"`
	TestUser TestUser   `yaml:"testUser"`
}

type TestUser struct {
	Allowed  bool   `yaml:"allowed"`
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
}

type AuthServer struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Log struct {
	Title  string `yaml:"title"`
	Format string `yaml:"format"`
	Level  string `yaml:"level"`
}

func Gist() *Config {
	if config == nil {
		config = &Config{}
	}
	return config
}

func Init() *Config {
	fileNamePtr := flag.String("c", "", "config file name")
	mode := flag.String("mode", "", "Application mode")
	flag.Parse()
	config = &Config{}
	vip := viper.New()
	fileName := *fileNamePtr
	if fileName != "" {
		vip.SetConfigName(fileName)
		vip.SetConfigType("yaml")
		vip.AddConfigPath(".")
		vip.AddConfigPath("./etc")
		if err := vip.ReadInConfig(); err != nil {
			panic(fmt.Errorf("cannot read config file: %w", err))
		}
	} else {
		panic("config name not set")
	}
	err := vip.Unmarshal(config)
	if err != nil {
		panic(fmt.Errorf("cannot unmarshal config: %w", err))
	}
	if *mode != "" {
		config.Mode = *mode
	}
	return config
}
