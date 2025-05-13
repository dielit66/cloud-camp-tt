package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server      Server      `yaml:"server"`
	LoggerLevel int8        `yaml:"logger_level"`
	BackendPool BackendPool `yaml:"pool"`
	RateLimiter RateLimiter `yaml:"rate_limiter"`
}

type Server struct {
	Port     string `env:"PORT" env-default:"8080" yaml:"port"`
	LBMethod string `yaml:"lb_method"`
}

type BackendPool struct {
	URLs        []string    `yaml:"urls"`
	HealthCheck HealthCheck `yaml:"healthcheck"`
}

type HealthCheck struct {
	Timeout  time.Duration `yaml:"timeout"`
	Endpoint string        `yaml:"endpoint"`
}

type RateLimiter struct {
	Enabled          bool          `yaml:"enabled"`
	CleanupInterval  time.Duration `yaml:"cleanup_interval"`
	RefillInterval   time.Duration `yaml:"refill_interval"`
	BucketExpiration time.Duration `yaml:"bucket_expiration"`
	Default          struct {
		RequestsPerSecond int `yaml:"requests_per_sec"`
		Burst             int `yaml:"burst"`
	} `yaml:"default"`
}

func LoadConfig(filename string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(filename, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
