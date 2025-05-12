package healthcheck

import (
	"context"
	"net/http"
	"time"

	"github.com/dielit66/cloud-camp-tt/internal/backend"
)

type HealthChecker struct {
	Endpoint   string
	Timeout    time.Duration
	httpClient *http.Client
}

func NewHealthChecker(endpoint string, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		Endpoint: endpoint,
		Timeout:  timeout,

		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (hc *HealthChecker) Check(ctx context.Context, b *backend.Backend) bool {
	url := b.URL.String() + hc.Endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		b.SetAlive(false)
		return false
	}

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		b.SetAlive(false)
		return false
	}
	defer resp.Body.Close()

	isAlive := resp.StatusCode == http.StatusOK
	b.SetAlive(isAlive)
	return isAlive
}

func (hc *HealthChecker) Start(ctx context.Context, backends []*backend.Backend, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, b := range backends {
				hc.Check(ctx, b)
			}
		case <-ctx.Done():
			return
		}
	}
}
