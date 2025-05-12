package server

import (
	"net/http"

	"github.com/dielit66/cloud-camp-tt/internal/backend"
	"github.com/dielit66/cloud-camp-tt/pkg/logging"
)

type LoadBalancer struct {
	server *http.Server
	pool   *backend.Pool
	logger logging.ILogger
}

func NewLoadBalancer(pool *backend.Pool, l logging.ILogger) *LoadBalancer {
	return &LoadBalancer{
		pool:   pool,
		logger: l,
	}
}
