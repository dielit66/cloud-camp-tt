package server

import (
	"net/http"
	"time"

	"github.com/dielit66/cloud-camp-tt/pkg/middleware"
)

func (lb *LoadBalancer) LBLeastConnectionsMethod(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.RequestIDKey).(string)

	lb.logger.Debug("new request", map[string]interface{}{
		"client_ip":  r.RemoteAddr,
		"lb_method":  "LeastConnections",
		"request_id": requestID,
		"time":       time.Now().Format(time.RFC3339),
	})

	b := lb.pool.GetLessLoadedBackend()

	if b == nil {
		lb.logger.Error("all backends are down", map[string]interface{}{
			"client_ip":  r.RemoteAddr,
			"lb_method":  "LeastConnections",
			"request_id": requestID,
			"time":       time.Now().Format(time.RFC3339),
		})
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("all backends are down"))
		return
	}

	lb.logger.Debug("least busy backend was found", map[string]interface{}{
		"client_ip":          r.RemoteAddr,
		"host":               b.URL.String(),
		"active_connections": b.ActiveConnections,
		"lb_method":          "LeastConnections",
		"request_id":         requestID,
		"time":               time.Now().Format(time.RFC3339),
	})

	b.AddConnection()
	defer b.ConnectionDone()

	b.Proxy.ServeHTTP(w, r)

	lb.logger.Debug("request was proxied", map[string]interface{}{
		"client_ip":  r.RemoteAddr,
		"host":       b.URL.String(),
		"lb_method":  "LeastConnections",
		"request_id": requestID,
		"time":       time.Now().Format(time.RFC3339),
	})

}
