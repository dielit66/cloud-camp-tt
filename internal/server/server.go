package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/dielit66/cloud-camp-tt/internal/config"
)

func (lb *LoadBalancer) StartServer(cfg *config.Config) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", lb.LBRoundRobinMethod)

	lb.server = &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := lb.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen and serve returned err: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("got interruption signal")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := lb.server.Shutdown(ctx); err != nil {
		log.Printf("server shutdown returned err: %v", err)
		return err

	}

	log.Println("server was gracefully stopped")
	return nil
}
