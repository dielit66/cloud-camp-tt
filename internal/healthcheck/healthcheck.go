package healthcheck

import (
	"context"
	"net/http"
	"time"

	"github.com/dielit66/cloud-camp-tt/internal/backend"
	"github.com/dielit66/cloud-camp-tt/pkg/logging"
)

type HealthChecker struct {
	Endpoint   string
	Timeout    time.Duration
	httpClient *http.Client
	logger     logging.ILogger
}

func NewHealthChecker(endpoint string, timeout time.Duration, l logging.ILogger) *HealthChecker {
	return &HealthChecker{
		Endpoint: endpoint,
		Timeout:  timeout,

		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: l,
	}
}

func (hc *HealthChecker) Check(ctx context.Context, b *backend.Backend) bool {
	url := b.URL.String() + hc.Endpoint

	hc.logger.Info("Starting health check for backend", map[string]interface{}{
		"url": url,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		hc.logger.Error("Failed to create health check request", map[string]interface{}{
			"url":   url,
			"error": err.Error(),
		})
		b.SetAlive(false)
		return false
	}

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		hc.logger.Error("Health check request failed", map[string]interface{}{
			"url":   url,
			"error": err.Error(),
		})
		b.SetAlive(false)
		return false
	}
	defer resp.Body.Close()

	isAlive := resp.StatusCode == http.StatusOK
	b.SetAlive(isAlive)

	if isAlive {
		hc.logger.Info("Health check succeeded", map[string]interface{}{
			"url":         url,
			"status_code": resp.StatusCode,
		})
	} else {
		hc.logger.Warn("Health check failed", map[string]interface{}{
			"url":         url,
			"status_code": resp.StatusCode,
		})
	}

	return isAlive
}

func (hc *HealthChecker) Start(ctx context.Context, backends []*backend.Backend, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.logger.Debug("Running health check cycle", map[string]interface{}{
				"backends": len(backends),
			})
			for _, b := range backends {
				hc.Check(ctx, b)
			}
			hc.logger.Debug("Completed health check cycle", map[string]interface{}{
				"backends": len(backends),
			})
		case <-ctx.Done():
			return
		}
	}
}
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

}
