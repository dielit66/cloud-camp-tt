package main

import (
	"context"
	"log"
	"time"

	"github.com/dielit66/cloud-camp-tt/internal/backend"
	"github.com/dielit66/cloud-camp-tt/internal/config"
	"github.com/dielit66/cloud-camp-tt/internal/healthcheck"
	"github.com/dielit66/cloud-camp-tt/internal/server"
)

func main() {
	cfg, err := config.LoadConfig("config/config.yaml")

	if err != nil {
		log.Fatalf("error while reading config, err: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := backend.NewPool(cfg.BackendPool.URLs)

	healthChecker := healthcheck.NewHealthChecker(
		cfg.BackendPool.HealthCheck.Endpoint,
		time.Duration(cfg.BackendPool.HealthCheck.Timeout)*time.Second,
	)

	go healthChecker.Start(
		ctx,
		pool.Backends,
		time.Duration(cfg.BackendPool.HealthCheck.Timeout)*time.Second,
	)

	lb := server.NewLoadBalancer(pool)

	if err = lb.StartServer(cfg); err != nil {
		log.Fatal(err)
	}
}
