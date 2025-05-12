package backend

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	URL   *url.URL
	Proxy *httputil.ReverseProxy
	alive bool
	mux   sync.RWMutex
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
