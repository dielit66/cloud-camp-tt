package backend

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Pool struct {
	backends []*Backend
	current  uint
	mux      sync.RWMutex
}

func NewPool(urls []string) *Pool {
	var pool Pool

	for _, u := range urls {
		parsedUrl, _ := url.Parse(u)
		pool.backends = append(pool.backends, &Backend{
			URL:   parsedUrl,
			Proxy: httputil.NewSingleHostReverseProxy(parsedUrl),
			alive: true,
		})
	}
	pool.current = 0

	return &pool
}

func (p *Pool) StartHealthChecker(ctx context.Context, endpoint string, tickerTimer time.Duration) {
	ticker := time.NewTicker(tickerTimer)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, b := range p.backends {
				res, err := http.Get(b.URL.String() + endpoint)
				isAlive := err == nil && res.StatusCode == http.StatusOK
				b.SetAlive(isAlive)
			}
		case <-ctx.Done():
			log.Println("HealthCheck stopped")
			return
		}
	}
}

func (p *Pool) GetBackendsLength() int {
	return len(p.backends)
}

func (p *Pool) Next() *Backend {
	p.mux.Lock()
	defer p.mux.Unlock()

	var nextIndex uint

	if int(p.current+1) == len(p.backends) {
		nextIndex = 0
	} else {

		nextIndex = p.current + 1
	}
	p.current = nextIndex
	return p.backends[nextIndex]
}
