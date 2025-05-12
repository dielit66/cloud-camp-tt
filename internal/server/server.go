package server

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/dielit66/cloud-camp-tt/internal/config"
	"github.com/dielit66/cloud-camp-tt/pkg/middleware"
)

func (lb *LoadBalancer) StartServer(cfg *config.Config) error {
	mux := http.NewServeMux()

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

	mux.HandleFunc("/", LBMethod)

	lb.server = &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: middleware.WithRequestID(mux),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		lb.logger.Info("starting http server", map[string]interface{}{
			"port": cfg.Port,
		})
		if err := lb.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			lb.logger.Fatal("listen and serve returned error", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	<-ctx.Done()

	lb.logger.Info("got interruption signal", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := lb.server.Shutdown(ctx); err != nil {
		lb.logger.Error("server shutdown returned error", map[string]interface{}{
			"error": err.Error(),
		})
		return err

	}

	lb.logger.Info("server was gracefully stopped", nil)
	return nil
}
