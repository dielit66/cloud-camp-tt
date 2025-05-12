package server

import (
	"net/http"

	"github.com/dielit66/cloud-camp-tt/internal/backend"
)

type LoadBalancer struct {
	server *http.Server
	pool   *backend.Pool
}

func NewLoadBalancer(pool *backend.Pool) *LoadBalancer {
	return &LoadBalancer{
		pool: pool,
	}
}
