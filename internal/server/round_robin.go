package server

import (
	"net/http"
	"time"

	"github.com/dielit66/cloud-camp-tt/pkg/errors"
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

	err := errors.NewAPIError(http.StatusServiceUnavailable, "Sorry, the service is currently unavailable. Please try again later.")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)
	w.Write(err.ToJSON())
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
