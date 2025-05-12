package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Port        string `env:"PORT" env-default:"8080" yaml:"port"`
	LBMethod    string `yaml:"lb_method"`
	LoggerLevel int8   `yaml:"logger_level"`
	BackendPool struct {
		URLs        []string `yaml:"urls"`
		HealthCheck struct {
			Timeout  int    `yaml:"timeout"`
			Endpoint string `yaml:"endpoint"`
		} `yaml:"healthcheck"`
	} `yaml:"pool"`
}

func LoadConfig(filename string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(filename, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
