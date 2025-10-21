// cmd/gateway/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"example.com/api-gateway/config"
	"example.com/api-gateway/internal/http"
	"example.com/api-gateway/internal/logger"
	"example.com/api-gateway/internal/rate"
	rds "example.com/api-gateway/internal/redis"
	"example.com/api-gateway/internal/repository"
	"example.com/api-gateway/internal/service"
	"go.uber.org/zap"
)

func main() {
	// 1) Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	// 2) Redis client (for rate limiting or other features) + async log worker
	rclient, err := rds.NewClient(cfg.Redis, zap.NewNop())
	if err != nil {
		panic(fmt.Errorf("redis connect: %w", err))
	}
	asyncRedis := rds.NewAsyncLogger(rclient, 2048)
	asyncRedis.Start()
	defer asyncRedis.Stop()

	// 3) Logger (console + rotating file) + Redis hook (forward each entry to async)
	log, err := logger.New(cfg.Logging, asyncRedis)
	if err != nil {
		panic(fmt.Errorf("init logger: %w", err))
	}
	defer log.Sync()

	// 4) Rate limiter
	var limiter rate.Limiter
	if cfg.RateLimit.Enabled {
		switch cfg.RateLimit.Strategy {
		case "memory":
			limiter = rate.NewMemoryLimiter(cfg.RateLimit, log)
		case "redis":
			limiter = rate.NewRedisLimiter(rclient, cfg.RateLimit, log)
		default:
			limiter = rate.Noop{}
		}
	} else {
		limiter = rate.Noop{}
	}

	// 5) Repository + services (GORM-based repo constructed from cfg.Database)
	userRepo, err := repository.NewUserRepository(cfg.Database, log)
	if err != nil {
		log.Fatal("user repo init failed", zap.Error(err))
	}

	authSvc := service.NewAuthService(userRepo, cfg.Security.JWT, log)
	userSvc := service.NewUserService(userRepo, log)

	// 6) Router
	engine := httpx.NewRouter(cfg, log, authSvc, userSvc, limiter, asyncRedis, rclient)

	// 7) HTTP Server with timeouts from config
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:           engine,
		ReadTimeout:       time.Duration(cfg.Server.Timeouts.ReadMS) * time.Millisecond,
		ReadHeaderTimeout: time.Duration(cfg.Server.Timeouts.ReadHeaderMS) * time.Millisecond,
		WriteTimeout:      time.Duration(cfg.Server.Timeouts.WriteMS) * time.Millisecond,
		IdleTimeout:       time.Duration(cfg.Server.Timeouts.IdleMS) * time.Millisecond,
	}

	log.Info("http server starting",
		zap.String("addr", srv.Addr),
	)

	// 8) Start server (basic blocking; add graceful shutdown as needed)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("listen and serve failed", zap.Error(err))
	}

	// 9) On exit, best-effort context cancel for Redis pings (example)
	_ = rclient.Close()
	_ = context.Canceled
}
