package redis // Client init for standalone or sentinel

import (
	"context"
	"crypto/tls"
	"time"

	redis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"example.com/api-gateway/config"
)

// Client is a thin alias to decouple imports.
type Client = redis.UniversalClient

// NewClient returns a universal client supporting standalone or sentinel.
func NewClient(cfg config.Redis, log *zap.Logger) (Client, error) {
	opt := &redis.UniversalOptions{ // works for single or sentinel
		Addrs:    cfg.Addresses,
		DB:       cfg.DB,
		Username: cfg.Username,
		Password: cfg.Password,
	}
	if cfg.Mode == "sentinel" {
		opt.MasterName = cfg.MasterName
	}
	if cfg.TLS {
		opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	client := redis.NewUniversalClient(opt)

	// Quick health check with timeout (v9 requires you to pass a context)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := client.Ping(ctx).Result(); err != nil {
		log.Error("redis ping failed", zap.Error(err))
		return nil, err
	}

	log.Info("redis connected", zap.Strings("addrs", cfg.Addresses), zap.String("mode", cfg.Mode))
	return client, nil
}
