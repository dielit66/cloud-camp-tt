package server

import (
	"fmt"
	"net/http"
)

func (lb *LoadBalancer) LBRoundRobinMethod(w http.ResponseWriter, r *http.Request) {
	for i := 0; i < lb.pool.GetBackendsLength(); i++ {
		b := lb.pool.Next()

		fmt.Println(b.URL.String())

		if !b.IsAlive() {
			continue
		}

		b.Proxy.ServeHTTP(w, r)
		return
	}

	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte("service unavailable"))
}
