package main // Entry point package

import ( // Standard + project imports
	"context" // For timeouts and cancelation
	"fmt"     // Formatting strings (e.g., bind address)
	"net/http" // HTTP server
	"os"       // For env vars
	"time"     // For server timeouts

	"example.com/api-gateway/config" // Config loader
	"example.com/api-gateway/internal/http" // Router
	"example.com/api-gateway/internal/logger" // Zap logger init
	"example.com/api-gateway/internal/redis" // Redis client init
	"example.com/api-gateway/internal/repository" // Repo interfaces + adapters
	"example.com/api-gateway/internal/rate" // Rate limiter impls
	"example.com/api-gateway/internal/service" // Services (auth/user)

	"go.uber.org/zap" // Logger type
)

func main() { // Orchestrates bootstrap and server run
	// 1) Load configuration (env overrides handled in loader)
	cfg, err := config.Load() // Reads config/config.yaml + env overrides
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err)) // Fail fast if config invalid
	}

	// 2) Initialize logger with level/JSON/sampling as per config
	log := logger.New(cfg.Logging) // Build *zap.Logger
	defer log.Sync()               // Flush logs on exit

	// 3) Initialize optional Redis (for rate-limiter) if enabled in config
	var redisClient redis.Client // Interface (nil if not used)
	if cfg.RateLimit.Enabled && cfg.RateLimit.Strategy == "redis" { // Only if redis strategy chosen
		client, rErr := redis.NewClient(cfg.Redis, log) // Create (standalone or sentinel)
		if rErr != nil {
			log.Fatal("redis init failed", zap.Error(rErr)) // Production: no silent fallback here
		}
		redisClient = client // Keep for limiter wiring
	}

	// 4) Initialize repositories based on database driver
	userRepo, dErr := repository.NewUserRepository(cfg.Database, log) // Switchable by config
	if dErr != nil {
		log.Fatal("db init failed", zap.Error(dErr)) // Fail if DB broken
	}

	// 5) Initialize services (business logic)
	userSvc := service.NewUserService(userRepo, log) // CRUD + rules
	authSvc := service.NewAuthService(userRepo, cfg.Security.JWT, log) // Login/JWT

	// 6) Initialize the rate limiter implementation
	var limiter rate.Limiter // Interface: Allow(key) -> (bool, retryAfter)
	if cfg.RateLimit.Enabled { // Only build if enabled
		if cfg.RateLimit.Strategy == "redis" {
			limiter = rate.NewRedisLimiter(redisClient, cfg.RateLimit, log) // Shared across workers
		} else {
			limiter = rate.NewMemoryLimiter(cfg.RateLimit, log) // In-process token bucket
		}
	}

	// 7) Build the HTTP router with all middlewares and handlers
	r := httpx.NewRouter(cfg, log, authSvc, userSvc, limiter) // httpx = internal/http

	// 8) Build server with timeouts
	srv := &http.Server{ // Production timeouts to avoid Slowloris
		Addr:              fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:           r, // Gin router as handler
		ReadTimeout:       time.Duration(cfg.Server.Timeouts.ReadMS) * time.Millisecond,
		ReadHeaderTimeout: time.Duration(cfg.Server.Timeouts.ReadHeaderMS) * time.Millisecond,
		WriteTimeout:      time.Duration(cfg.Server.Timeouts.WriteMS) * time.Millisecond,
		IdleTimeout:       time.Duration(cfg.Server.Timeouts.IdleMS) * time.Millisecond,
	}

	// 9) Start server (blocking)
	log.Info("http server starting", zap.String("addr", srv.Addr))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("server failed", zap.Error(err)) // Crash on unexpected error
	}

	// 10) Graceful shutdown example (not used with plain ListenAndServe). Left for reference.
	_ = os.Setenv("_", "") // no-op to silence unused import if you later switch to graceful shutdown
	_ = context.Background() // placeholder to indicate where shutdown ctx would live
}