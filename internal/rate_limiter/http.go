package ratelimiter

import (
	"net/http"

	"github.com/dielit66/cloud-camp-tt/pkg/errors"
	"github.com/dielit66/cloud-camp-tt/pkg/logging"
)

func NewRateLimiterHandler(rl *RateLimiter, logger logging.ILogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			rl.logger.Info("Client IP extracted", map[string]interface{}{
				"ip": ip,
			})
			ctx := r.Context()

			if !rl.Allow(ctx, ip) {
				logger.Warn("Rate limit exceeded", map[string]interface{}{
					"ip": ip,
				})
				err := errors.NewAPIError(http.StatusTooManyRequests, "Sorry, too many request from this ip address, try again later")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(err.Code)
				w.Write(err.ToJSON())
				return
			}

			logger.Info("Request allowed ", map[string]interface{}{
				"ip": ip,
			})
			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP извлекает IP-адрес клиента
func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}
