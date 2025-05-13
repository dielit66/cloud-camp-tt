package main

import (
	"context"
	"log"

	"github.com/dielit66/cloud-camp-tt/internal/backend"
	"github.com/dielit66/cloud-camp-tt/internal/config"
	"github.com/dielit66/cloud-camp-tt/internal/healthcheck"
	ratelimiter "github.com/dielit66/cloud-camp-tt/internal/rate_limiter"
	repository "github.com/dielit66/cloud-camp-tt/internal/repository/bucket"
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
		cfg.BackendPool.HealthCheck.Timeout,
		logger,
	)

	repo := repository.NewRedisBucketSettingsRepository(logger, &cfg.RateLimiter)
	rl := ratelimiter.NewRateLimiter(ctx, repo, logger, &cfg.RateLimiter)

	go healthChecker.Start(
		ctx,
		pool.Backends,
		cfg.BackendPool.HealthCheck.Timeout,
	)

	lb := server.NewLoadBalancer(pool, logger, rl, repo)

	if err = lb.StartServer(&cfg.Server); err != nil {
		logger.Fatal(err.Error(), nil)
	}
}
