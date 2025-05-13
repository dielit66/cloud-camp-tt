package repository

import (
	"context"
	"encoding/json"

	ratelimiter "github.com/dielit66/cloud-camp-tt/internal/rate_limiter"
	"github.com/dielit66/cloud-camp-tt/pkg/logging"
	"github.com/redis/go-redis/v9"
)

type RedisBucketSettingsRepository struct {
	db     *redis.Client
	logger logging.ILogger
}

func NewRedisBucketSettingsRepository(l logging.ILogger) *RedisBucketSettingsRepository {
	return &RedisBucketSettingsRepository{
		db: redis.NewClient(&redis.Options{
			Addr:     "cache:6379",
			Password: "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81",
			DB:       0,
		}),
		logger: l,
	}
}

func (r *RedisBucketSettingsRepository) GetConfig(ctx context.Context, ip string) (ratelimiter.Config, error) {
	key := "rate_limit:config:" + ip
	val, err := r.db.Get(ctx, key).Result()
	if err == redis.Nil {
		r.logger.Info("no config found, using default", map[string]interface{}{
			"ip": ip,
		})
		return ratelimiter.Config{MaxTokens: 10, RefillRate: 1}, nil
	}
	if err != nil {
		r.logger.Error("failed to get config", map[string]interface{}{
			"ip":    ip,
			"error": err.Error(),
		})
		return ratelimiter.Config{}, err
	}

	var config ratelimiter.Config
	if err := json.Unmarshal([]byte(val), &config); err != nil {
		r.logger.Error("failed to unmarshal config", map[string]interface{}{
			"ip":    ip,
			"error": err.Error(),
		})
		return ratelimiter.Config{}, err
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
	r.logger.Info("Config set for IP ", map[string]interface{}{
		"ip": ip,
	})
	return nil
}
