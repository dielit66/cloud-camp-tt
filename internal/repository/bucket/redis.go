package repository

import (
	"context"
	"encoding/json"

	"github.com/dielit66/cloud-camp-tt/internal/config"
	ratelimiter "github.com/dielit66/cloud-camp-tt/internal/rate_limiter"
	"github.com/dielit66/cloud-camp-tt/pkg/logging"
	"github.com/redis/go-redis/v9"
)

type RedisBucketSettingsRepository struct {
	db     *redis.Client
	logger logging.ILogger
	cfg    *config.RateLimiter
}

func NewRedisBucketSettingsRepository(l logging.ILogger, cfg *config.RateLimiter) *RedisBucketSettingsRepository {
	return &RedisBucketSettingsRepository{
		db: redis.NewClient(&redis.Options{
			Addr:     cfg.RateLimiterDb.Host + ":" + cfg.RateLimiterDb.Port,
			Password: cfg.RateLimiterDb.Password,
			DB:       0,
		}),
		logger: l,
		cfg:    cfg,
	}
}

func (r *RedisBucketSettingsRepository) GetConfig(ctx context.Context, ip string) (ratelimiter.Config, error) {
	key := "rate_limit:config:" + ip
	val, err := r.db.Get(ctx, key).Result()
	if err == redis.Nil {
		r.logger.Info("No config found, using default", map[string]interface{}{
			"ip": ip,
		})
		return ratelimiter.Config{
			MaxTokens:  r.cfg.Default.MaxTokens,
			RefillRate: r.cfg.Default.RefillRate,
		}, nil
	}
	if err != nil {
		r.logger.Error("Failed to get config", map[string]interface{}{
			"ip":    ip,
			"error": err.Error(),
		})
		return ratelimiter.Config{}, err
	}

	var config ratelimiter.Config
	if err := json.Unmarshal([]byte(val), &config); err != nil {
		r.logger.Error("Failed to unmarshal config", map[string]interface{}{
			"ip":    ip,
			"error": err.Error(),
		})
		return ratelimiter.Config{}, err
	}

	if config.MaxTokens <= 0 || config.RefillRate <= 0 {
		r.logger.Warn("Invalid config, using default", map[string]interface{}{
			"ip":          ip,
			"max_tokens":  config.MaxTokens,
			"refill_rate": config.RefillRate,
		})
		return ratelimiter.Config{
			MaxTokens:  r.cfg.Default.MaxTokens,
			RefillRate: r.cfg.Default.RefillRate,
		}, nil
	}

	return config, nil
}

func (r *RedisBucketSettingsRepository) SetConfig(ctx context.Context, ip string, config ratelimiter.Config) error {
	key := "rate_limit:config:" + ip
	data, err := json.Marshal(config)
	if err != nil {
		r.logger.Error("Failed to marshal config", map[string]interface{}{
			"ip":    ip,
			"error": err.Error(),
		})
		return err
	}

	if err := r.db.Set(ctx, key, data, 0).Err(); err != nil {
		r.logger.Error("Failed to set config", map[string]interface{}{
			"ip":    ip,
			"error": err.Error(),
		})
		return err
	}

	r.logger.Info("Config set for IP", map[string]interface{}{
		"ip":          ip,
		"max_tokens":  config.MaxTokens,
		"refill_rate": config.RefillRate,
	})
	return nil
}

func (r *RedisBucketSettingsRepository) DeleteConfig(ctx context.Context, ip string) error {
	key := "rate_limit:config:" + ip
	if err := r.db.Del(ctx, key).Err(); err != nil {
		r.logger.Error("Failed to delete config", map[string]interface{}{
			"ip":    ip,
			"error": err.Error(),
		})
		return err
	}
	r.logger.Info("Config deleted from Redis", map[string]interface{}{
		"ip": ip,
	})
	return nil
}
