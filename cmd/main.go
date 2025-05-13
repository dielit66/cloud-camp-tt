package main

import (
	"context"
	"log"
	"time"

	"github.com/dielit66/cloud-camp-tt/internal/backend"
	"github.com/dielit66/cloud-camp-tt/internal/config"
	"github.com/dielit66/cloud-camp-tt/internal/healthcheck"
	"github.com/dielit66/cloud-camp-tt/internal/server"
	"github.com/dielit66/cloud-camp-tt/pkg/logging"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")

	if err != nil {
		log.Fatalf("error while reading config, err: %v", err)
	}

	logger := logging.NewZeroLogger(cfg.LoggerLevel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := backend.NewPool(cfg.BackendPool.URLs, logger)

	healthChecker := healthcheck.NewHealthChecker(
		cfg.BackendPool.HealthCheck.Endpoint,
		time.Duration(cfg.BackendPool.HealthCheck.Timeout)*time.Second,
		logger,
	)

	go healthChecker.Start(
		ctx,
		pool.Backends,
		time.Duration(cfg.BackendPool.HealthCheck.Timeout)*time.Second,
	)

	lb := server.NewLoadBalancer(pool, logger)

	if err = lb.StartServer(&cfg.Server); err != nil {
		logger.Fatal(err.Error(), nil)
	}
}
