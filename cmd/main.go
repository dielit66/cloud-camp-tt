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
	// Подгружаем конфиг из config.yaml
	cfg, err := config.LoadConfig("config.yaml")

	if err != nil {
		log.Fatalf("error while reading config, err: %v", err)
	}

	// Инициализируем логгер
	logger := logging.NewZeroLogger(cfg.LoggerLevel)

	// Создаем контекст с возможностью отмены для управления жизненным циклом сервисов
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := backend.NewPool(cfg.BackendPool.URLs, logger)

	// Инициализируем health checker для периодической проверки состояния бэкендов
	healthChecker := healthcheck.NewHealthChecker(
		cfg.BackendPool.HealthCheck.Endpoint,
		cfg.BackendPool.HealthCheck.Timeout,
		logger,
	)

	// Инициализируем rate limiter для ограничения частоты запросов клиентов
	repo := repository.NewRedisBucketSettingsRepository(logger, &cfg.RateLimiter)
	rl := ratelimiter.NewRateLimiter(ctx, repo, logger, &cfg.RateLimiter)

	// Запускаем health checker в отдельной горутине для периодической проверки бэкендов
	go healthChecker.Start(
		ctx,
		pool.Backends,
		cfg.BackendPool.HealthCheck.Timeout,
	)

	lb := server.NewLoadBalancer(pool, logger, rl, repo)

	// Запускаем HTTP-сервер load balancer'а
	if err = lb.StartServer(&cfg.Server); err != nil {
		logger.Fatal(err.Error(), nil)
	}
}
