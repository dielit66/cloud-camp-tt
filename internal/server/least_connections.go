package server

import (
	"net/http"
)

func (lb *LoadBalancer) LBLeastConnectionsMethod(w http.ResponseWriter, r *http.Request) {
	b := lb.pool.GetLessLoadedBackend()

	if b == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("all backends are dead"))
		return
	}

	b.AddConnection()
	defer b.ConnectionDone()

	b.Proxy.ServeHTTP(w, r)
}
