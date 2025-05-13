package server

import (
	"encoding/json"
	"net/http"
	"strings"

	ratelimiter "github.com/dielit66/cloud-camp-tt/internal/rate_limiter"
	"github.com/dielit66/cloud-camp-tt/pkg/errors"
)

func (lb *LoadBalancer) handleRateLimitConfig(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/api/ratelimit/config") {
		err := errors.NewAPIError(http.StatusNotFound, "Not found")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(err.Code)
		w.Write(err.ToJSON())
		return
	}

	switch r.Method {
	case http.MethodPost:
		if path != "/api/ratelimit/config" {
			err := errors.NewAPIError(http.StatusNotFound, "Not found")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(err.Code)
			w.Write(err.ToJSON())
			return
		}
		lb.handleAddConfig(w, r)
	case http.MethodGet:
		if !strings.HasPrefix(path, "/api/ratelimit/config/") {
			err := errors.NewAPIError(http.StatusNotFound, "Not found")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(err.Code)
			w.Write(err.ToJSON())
			return
		}
		lb.handleGetConfig(w, r)
	case http.MethodDelete:
		if !strings.HasPrefix(path, "/api/ratelimit/config/") {
			err := errors.NewAPIError(http.StatusNotFound, "Not found")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(err.Code)
			w.Write(err.ToJSON())
			return
		}
		lb.handleDeleteConfig(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (lb *LoadBalancer) handleAddConfig(w http.ResponseWriter, r *http.Request) {
	var req ratelimiter.AddConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lb.logger.Warn("Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		err := errors.NewAPIError(http.StatusBadRequest, "Invalid request body")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(err.Code)
		w.Write(err.ToJSON())
		return
	}

	if err := req.Validate(); err != nil {
		lb.logger.Warn("Invalid config request", map[string]interface{}{
			"ip":    req.IP,
			"error": err.Error(),
		})

		err := errors.NewAPIError(http.StatusBadRequest, err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(err.Code)
		w.Write(err.ToJSON())
		return
	}

	config := ratelimiter.Config{
		MaxTokens:  req.MaxTokens,
		RefillRate: req.RefillRate,
	}

	if err := lb.repo.SetConfig(r.Context(), req.IP, config); err != nil {
		lb.logger.Error("Failed to set config", map[string]interface{}{
			"ip":    req.IP,
			"error": err.Error(),
		})
		err := errors.NewAPIError(http.StatusInternalServerError, "Failed to set config")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(err.Code)
		w.Write(err.ToJSON())
		return
	}

	// Очищаем bucket, чтобы новая конфигурация применилась
	lb.rl.ClearBucket(req.IP)

	lb.logger.Info("Bucket cleared after setting new config", map[string]interface{}{
		"ip": req.IP,
	})

	resp := ratelimiter.ConfigResponse{
		IP:         req.IP,
		MaxTokens:  req.MaxTokens,
		RefillRate: req.RefillRate,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		lb.logger.Error("Failed to encode response", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

func (lb *LoadBalancer) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	ip := strings.TrimPrefix(r.URL.Path, "/api/ratelimit/config/")
	if ip == "" {
		err := errors.NewAPIError(http.StatusBadRequest, "IP is required")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(err.Code)
		w.Write(err.ToJSON())
		return
	}

	config, err := lb.repo.GetConfig(r.Context(), ip)
	if err != nil {
		lb.logger.Error("Failed to get config", map[string]interface{}{
			"ip":    ip,
			"error": err.Error(),
		})
		err := errors.NewAPIError(http.StatusInternalServerError, "Failed to get config")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(err.Code)
		w.Write(err.ToJSON())
		return
	}

	resp := ratelimiter.ConfigResponse{
		IP:         ip,
		MaxTokens:  config.MaxTokens,
		RefillRate: config.RefillRate,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		lb.logger.Error("Failed to encode response", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

func (lb *LoadBalancer) handleDeleteConfig(w http.ResponseWriter, r *http.Request) {

	ip := strings.TrimPrefix(r.URL.Path, "/api/ratelimit/config/")
	if ip == "" {
		err := errors.NewAPIError(http.StatusBadRequest, "IP is required")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(err.Code)
		w.Write(err.ToJSON())
		return
	}

	if err := lb.repo.DeleteConfig(r.Context(), ip); err != nil {
		lb.logger.Error("Failed to delete config", map[string]interface{}{
			"ip":    ip,
			"error": err.Error(),
		})

		err := errors.NewAPIError(http.StatusInternalServerError, "Failed to delete config")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(err.Code)
		w.Write(err.ToJSON())
		return
	}

	lb.rl.ClearBucket(ip)
	w.WriteHeader(http.StatusNoContent)
}
