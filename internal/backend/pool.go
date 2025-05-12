package backend

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Pool struct {
	Backends []*Backend
	current  uint
	mux      sync.RWMutex
}

func NewPool(urls []string) *Pool {
	var pool Pool

	for _, u := range urls {
		parsedUrl, _ := url.Parse(u)
		pool.Backends = append(pool.Backends, &Backend{
			URL:               parsedUrl,
			Proxy:             httputil.NewSingleHostReverseProxy(parsedUrl),
			alive:             true,
			activeConnections: 0,
		})
	}
	pool.current = 0

	return &pool
}

// RR methods

func (p *Pool) GetBackendsLength() int {
	return len(p.Backends)
}

func (p *Pool) Next() *Backend {
	p.mux.Lock()
	defer p.mux.Unlock()

	var nextIndex uint

	if int(p.current+1) == len(p.Backends) {
		nextIndex = 0
	} else {

		nextIndex = p.current + 1
	}
	p.current = nextIndex
	return p.Backends[nextIndex]
}

// LC methods

func (p *Pool) GetLessLoadedBackend() *Backend {
	var lessLoadedBackend *Backend

	p.mux.RLock()
	defer p.mux.RUnlock()

	for _, b := range p.Backends {
		if !b.IsAlive() {
			continue
		}

		if lessLoadedBackend == nil || b.activeConnections < lessLoadedBackend.activeConnections {
			lessLoadedBackend = b
		}
	}

	return lessLoadedBackend
}
