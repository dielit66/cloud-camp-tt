package ratelimiter

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/dielit66/cloud-camp-tt/internal/config"
	"github.com/dielit66/cloud-camp-tt/pkg/logging"
)

// Config represents the rate limit configuration
type Config struct {
	MaxTokens  int `json:"max_tokens"`
	RefillRate int `json:"refill_rate"`
}

// AddConfigRequest represents the API request to add/update a configuration
type AddConfigRequest struct {
	IP         string `json:"ip"`
	MaxTokens  int    `json:"max_tokens"`
	RefillRate int    `json:"refill_rate"`
}

// Validate validates the request
func (r *AddConfigRequest) Validate() error {
	if r.IP == "" {
		return errors.New("ip is required")
	}
	if r.MaxTokens <= 0 {
		return errors.New("max_tokens must be positive")
	}
	if r.RefillRate <= 0 {
		return errors.New("refill_rate must be positive")
	}
	return nil
}

// ConfigResponse represents the API response
type ConfigResponse struct {
	IP         string `json:"ip"`
	MaxTokens  int    `json:"max_tokens"`
	RefillRate int    `json:"refill_rate"`
}

type ISettingsRepository interface {
	GetConfig(ctx context.Context, ip string) (Config, error)
	SetConfig(ctx context.Context, ip string, config Config) error
	DeleteConfig(ctx context.Context, ip string) error // Добавляем метод
}

type TokenBucket struct {
	tokens     float64
	config     Config
	lastRefill time.Time
}

type RateLimiter struct {
	buckets   map[string]*TokenBucket
	mutex     sync.RWMutex
	repo      ISettingsRepository
	logger    logging.ILogger
	ticker    *time.Ticker
	cfg       *config.RateLimiter
	isEnabled bool
}

func NewRateLimiter(ctx context.Context, repo ISettingsRepository, logger logging.ILogger, cfg *config.RateLimiter) *RateLimiter {
	rl := &RateLimiter{
		buckets:   make(map[string]*TokenBucket),
		repo:      repo,
		logger:    logger,
		cfg:       cfg,
		isEnabled: cfg.Enabled,
	}

	if cfg.Enabled {
		rl.ticker = time.NewTicker(cfg.RefillInterval)
		go rl.StartRefillLoop(ctx)
		rl.logger.Info("Rate limiter enabled, starting refill loop", map[string]interface{}{
			"refill_interval":   cfg.RefillInterval.String(),
			"bucket_expiration": cfg.BucketExpiration.String(),
		})
	} else {
		rl.logger.Info("Rate limiter disabled", nil)
	}

	return rl
}

func (rl *RateLimiter) IsEnabled() bool {
	return rl.isEnabled
}

func (rl *RateLimiter) Allow(ctx context.Context, ip string) bool {
	if !rl.isEnabled {
		rl.logger.Debug("Rate limiter disabled, allowing request", map[string]interface{}{
			"ip": ip,
		})
		return true
	}

	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	bucket, exists := rl.buckets[ip]
	if !exists {
		cfg, err := rl.repo.GetConfig(ctx, ip)
		if err != nil || cfg.MaxTokens == 0 {
			rl.logger.Info("No config found for IP, using default", map[string]interface{}{
				"ip":    ip,
				"error": err,
			})
			cfg = Config{
				MaxTokens:  rl.cfg.Default.MaxTokens,
				RefillRate: rl.cfg.Default.RefillRate,
			}
		}
		bucket = &TokenBucket{
			tokens:     float64(cfg.MaxTokens),
			config:     cfg,
			lastRefill: time.Now(),
		}
		rl.buckets[ip] = bucket
		rl.logger.Info("Created new bucket for IP", map[string]interface{}{
			"ip":          ip,
			"max_tokens":  cfg.MaxTokens,
			"refill_rate": cfg.RefillRate,
		})
	}

	rl.logger.Info("Checking rate limit for IP", map[string]interface{}{
		"ip":            ip,
		"tokens_before": bucket.tokens,
		"last_refill":   bucket.lastRefill,
	})

	if bucket.tokens >= 1 {
		bucket.tokens -= 1
		rl.logger.Info("Request allowed, tokens deducted", map[string]interface{}{
			"ip":           ip,
			"tokens_after": bucket.tokens,
		})
		return true
	}

	rl.logger.Warn("Request rate limited", map[string]interface{}{
		"ip":     ip,
		"tokens": bucket.tokens,
	})
	return false
}

func (rl *RateLimiter) ClearBucket(ip string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	delete(rl.buckets, ip)
	rl.logger.Info("Bucket cleared", map[string]interface{}{
		"ip": ip,
	})
}

func (rl *RateLimiter) StartRefillLoop(ctx context.Context) {
	rl.logger.Info("Starting rate limiter refill loop", map[string]interface{}{
		"refill_interval":   rl.cfg.RefillInterval.String(),
		"bucket_expiration": rl.cfg.BucketExpiration.String(),
	})

	for {
		select {
		case <-ctx.Done():
			rl.logger.Info("Stopping rate limiter refill loop", nil)
			rl.ticker.Stop()
			return
		case <-rl.ticker.C:
			rl.mutex.Lock()
			rl.logger.Debug("Running bucket refill cycle", map[string]interface{}{
				"buckets": len(rl.buckets),
			})
			for ip, bucket := range rl.buckets {
				rl.refillBucket(bucket)
				if time.Since(bucket.lastRefill) > rl.cfg.BucketExpiration {
					rl.logger.Info("Removing expired bucket", map[string]interface{}{
						"ip": ip,
					})
					delete(rl.buckets, ip)
				}
			}
			rl.logger.Debug("Completed bucket refill cycle", map[string]interface{}{
				"buckets": len(rl.buckets),
			})
			rl.mutex.Unlock()
		}
	}
}

func (rl *RateLimiter) refillBucket(bucket *TokenBucket) {
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	newTokens := bucket.tokens + elapsed*float64(bucket.config.RefillRate)
	if newTokens > float64(bucket.config.MaxTokens) {
		newTokens = float64(bucket.config.MaxTokens)
	}
	bucket.tokens = newTokens
	bucket.lastRefill = now
	rl.logger.Debug("Refilled bucket", map[string]interface{}{
		"tokens":      bucket.tokens,
		"max_tokens":  bucket.config.MaxTokens,
		"refill_rate": bucket.config.RefillRate,
		"elapsed_sec": elapsed,
	})
}
