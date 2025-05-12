package server

import (
	"net/http"
	"time"

	"github.com/dielit66/cloud-camp-tt/pkg/middleware"
)

func (lb *LoadBalancer) LBRoundRobinMethod(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.RequestIDKey).(string)

	lb.logger.Debug("new request", map[string]interface{}{
		"client_ip":  r.RemoteAddr,
		"lb_method":  "RoundRobin",
		"request_id": requestID,
		"time":       time.Now().Format(time.RFC3339),
	})

	for i := 0; i < lb.pool.GetBackendsLength(); i++ {
		b := lb.pool.Next()

		if !b.IsAlive() {
			lb.logger.Warn("backend service is unavailable, trying to find another service", map[string]interface{}{
				"host":       b.URL.String(),
				"lb_method":  "RoundRobin",
				"request_id": requestID,
				"time":       time.Now().Format(time.RFC3339),
			})
			continue
		}

		b.Proxy.ServeHTTP(w, r)

		lb.logger.Debug("request was proxied", map[string]interface{}{
			"client_ip":  r.RemoteAddr,
			"host":       b.URL.String(),
			"lb_method":  "RoundRobin",
			"request_id": requestID,
			"time":       time.Now().Format(time.RFC3339),
		})
		return
	}

	lb.logger.Error("all backends are down", map[string]interface{}{
		"client_ip":  r.RemoteAddr,
		"lb_method":  "RoundRobin",
		"request_id": requestID,
		"time":       time.Now().Format(time.RFC3339),
	})
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte("service unavailable"))
}
