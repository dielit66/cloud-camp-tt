package server

import (
	"net/http"

	"github.com/dielit66/cloud-camp-tt/internal/backend"
	ratelimiter "github.com/dielit66/cloud-camp-tt/internal/rate_limiter"
	"github.com/dielit66/cloud-camp-tt/pkg/logging"
)

type LoadBalancer struct {
	server *http.Server
	pool   *backend.Pool
	logger logging.ILogger
	rl     *ratelimiter.RateLimiter
}

func NewLoadBalancer(pool *backend.Pool, l logging.ILogger, rl *ratelimiter.RateLimiter) *LoadBalancer {
	return &LoadBalancer{
		pool:   pool,
		logger: l,
		rl:     rl,
	}
}
