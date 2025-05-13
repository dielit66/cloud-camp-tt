package server

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/dielit66/cloud-camp-tt/internal/config"
	"github.com/dielit66/cloud-camp-tt/internal/healthcheck"
	ratelimiter "github.com/dielit66/cloud-camp-tt/internal/rate_limiter"
	errors_middleware "github.com/dielit66/cloud-camp-tt/pkg/errors/middleware"
	"github.com/dielit66/cloud-camp-tt/pkg/middleware"
)

func (lb *LoadBalancer) StartServer(cfg *config.Server) error {
	// Создаем новый HTTP-мультиплексор для маршрутизации запросов
	mux := http.NewServeMux()

	// Определяем метод балансировки нагрузки на основе конфигурации
	var LBMethod http.HandlerFunc
	switch cfg.LBMethod {

	case "RR":
		lb.logger.Info("Round Robin method for load balancer was choosen", nil)
		LBMethod = lb.LBRoundRobinMethod

	case "LC":
		lb.logger.Info("Least Connections method for load balancer was choosen", nil)
		LBMethod = lb.LBLeastConnectionsMethod

	default:
		lb.logger.Info("Round Robin method for load balancer was choosen", nil)
		LBMethod = lb.LBRoundRobinMethod
	}

	// Основной маршрут для load balancer
	mux.HandleFunc("/", LBMethod)
	// - /healthcheck для проверки состояния load balancer'а
	mux.HandleFunc("/healthcheck", healthcheck.HealthCheckHandler)
	mux.HandleFunc("/api/ratelimit/config", lb.handleRateLimitConfig)
	mux.HandleFunc("/api/ratelimit/config/", lb.handleRateLimitConfig)

	// Оборачиваем в middleware для RequestID (сделал для логгирования и дебага по конкретному запросу), обработки ошибок и rate limiter
	handler := middleware.WithRequestID(mux)
	handler = errors_middleware.ErrorHandler(handler)
	handler = ratelimiter.NewRateLimiterHandler(lb.rl, lb.logger)(handler)

	lb.server = &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	// Создаем контекст для обработки сигналов прерывания (SIGINT, SIGTERM)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Запускаем сервер в отдельной горутине
	go func() {
		lb.logger.Info("starting http server", map[string]interface{}{
			"port": cfg.Port,
		})
		// Слушаем входящие соединения; логируем ошибку, если она не связана с graceful shutdown
		if err := lb.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			lb.logger.Fatal("listen and serve returned error", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Ожидаем сигнала прерывания
	<-ctx.Done()

	lb.logger.Info("got interruption signal", nil)

	// Создаем контекст с таймаутом 10 секунд для graceful shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Завершаем работу сервера, позволяя обработать текущие запросы.
	if err := lb.server.Shutdown(ctx); err != nil {
		lb.logger.Error("server shutdown returned error", map[string]interface{}{
			"error": err.Error(),
		})
		return err

	}

	lb.logger.Info("server was gracefully stopped", nil)
	return nil
}
