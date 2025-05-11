package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Port string `env:"PORT" env-default:"8080" yaml:"port"`
}

func LoadConfig(filename string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(filename, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
