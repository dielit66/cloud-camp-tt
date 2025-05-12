package backend

import (
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL               *url.URL
	Proxy             *httputil.ReverseProxy
	alive             bool
	mux               sync.RWMutex
	ActiveConnections int32
}

func (b *Backend) SetAlive(isAlive bool) {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.alive = isAlive
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()

	return b.alive
}

func (b *Backend) AddConnection() {
	atomic.AddInt32(&b.ActiveConnections, 1)
}
func (b *Backend) ConnectionDone() {
	atomic.AddInt32(&b.ActiveConnections, -1)
}
