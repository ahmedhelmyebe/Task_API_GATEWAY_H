// package httpx // Router wiring (Gin)

// import (
// 	"net/http"
// 	"time"

// 	"github.com/gin-gonic/gin"

// 	"example.com/api-gateway/config"
// 	// "example.com/api-gateway/internal/auth"
// 	// "example.com/api-gateway/internal/domain"
// 	"example.com/api-gateway/internal/handlers"
// 	"example.com/api-gateway/internal/http/middleware"
// 	"example.com/api-gateway/internal/rate"
// 	"example.com/api-gateway/internal/service"
// 	"github.com/prometheus/client_golang/prometheus/promhttp"
// 	"go.uber.org/zap"
// )

// // NewRouter builds the full HTTP router with routes and middleware.
// func NewRouter(cfg config.Root, log *zap.Logger, authSvc *service.AuthService, userSvc *service.UserService, limiter rate.Limiter) *gin.Engine {
// 	r := gin.New() // no default middleware
// 	r.Use(gin.Recovery()) // recover from panics
// 	r.Use(middleware.RequestID()) // add X-Request-Id
// 	r.Use(requestLogger(log)) // simple structured request log
// 	// CORS (minimal allow-all example; adapt for production)
// 	r.Use(func(c *gin.Context) {
// 		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
// 		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-Id")
// 		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
// 		if c.Request.Method == http.MethodOptions { c.AbortWithStatus(204); return }
// 		c.Next()
// 	})

// 	// Health & metrics
// 	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
// 	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

// 	// Middlewares that depend on config
// 	authRequired := middleware.Authenticated(cfg.Security.JWT)
// 	rlmw := middleware.RateLimit(limiter, cfg.RateLimit.RequestsPerMinute)

// // r.POST("/seed/admin", func(c *gin.Context) {
// //     req := struct {
// //         Name, Email, Password string
// //     }{}
// //     if err := c.ShouldBindJSON(&req); err != nil {
// //         c.JSON(400, gin.H{"error": err.Error()})
// //         return
// //     }

// //     hash, _ := auth.Hash(req.Password)
// //     u := &domain.User{
// //         Name: req.Name, Email: req.Email, PasswordHash: hash,
// //         Role: "admin", Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
// //     }

// //     if err := userSvc.Create(u); err != nil {
// //         c.JSON(500, gin.H{"error": err.Error()})
// //         return
// //     }
// //     c.JSON(201, gin.H{"message": "admin created", "id": u.ID})
// // })




// 	// Auth routes
// 	{
// 		h := handlersFrom(authSvc, userSvc) // internal helper builds handlers
// 		r.POST("/auth/login", rlmw, h.Auth.Login)
// 	}

// 	// Protected routes
// 	{
// 		h := handlersFrom(authSvc, userSvc)
// 		grp := r.Group("/")
// 		grp.Use(rlmw, authRequired)
// 		// Users (admin only)
// 		grp.GET("/users", middleware.RequireAdmin(), h.Users.List)
// 		grp.POST("/users", middleware.RequireAdmin(), h.Users.Create)
// 		grp.GET("/users/:id", middleware.RequireSelfOrAdmin(), h.Users.Get)
// 		grp.PATCH("/users/:id", middleware.RequireSelfOrAdmin(), h.Users.Patch)
// 		grp.DELETE("/users/:id", middleware.RequireAdmin(), h.Users.Delete)
// 		// Self utilities
// 		grp.GET("/users/me", h.Users.Me)
// 		grp.PATCH("/users/me", h.Users.PatchMe)
// 	}

// 	return r
// }

// // small aggregate to pass both handlers
// type pair struct { Auth *handlers.AuthHandler; Users *handlers.UserHandler }

// func handlersFrom(a *service.AuthService, u *service.UserService) pair {
// 	return pair{Auth: handlers.NewAuthHandler(a), Users: handlers.NewUserHandler(u)}
// }
// // requestLogger is a minimal structured access log middleware.
// func requestLogger(log *zap.Logger) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		start := time.Now()
// 		c.Next()
// 		log.Info("http",
// 			zap.String("id", c.GetString("req.id")),
// 			zap.String("method", c.Request.Method),
// 			zap.String("path", c.Request.URL.Path),
// 			zap.Int("status", c.Writer.Status()),
// 			zap.Duration("dur", time.Since(start)),
// 		)
// 	}
// }


// internal/http/router.go
package httpx // Router wiring (Gin)

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"example.com/api-gateway/config"
	"example.com/api-gateway/internal/handlers"
	"example.com/api-gateway/internal/http/middleware"
	"example.com/api-gateway/internal/rate"
	rlog "example.com/api-gateway/internal/redis"
	"example.com/api-gateway/internal/service"
)

// NewRouter builds the full HTTP router with routes and middleware.
// It now accepts an Async Redis Logger to persist enriched HTTP access logs.
func NewRouter(
	cfg config.Root,
	log *zap.Logger,
	authSvc *service.AuthService,
	userSvc *service.UserService,
	limiter rate.Limiter,
	redisAsync *rlog.AsyncLogger,
	rclient rlog.Client,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(requestLogger(log, redisAsync)) // file/console + async Redis
	// CORS (allow-all example; adapt for prod)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-Id")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health & metrics
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Middlewares that depend on config
	authRequired := middleware.Authenticated(cfg.Security.JWT)
	rlmw := middleware.RateLimit(limiter, cfg.RateLimit.RequestsPerMinute)

	// Handlers
	pair := handlersFrom(authSvc, userSvc)
	logsHandler := handlers.NewLogsHandler(rclient)

	// Auth routes
	r.POST("/auth/login", rlmw, pair.Auth.Login)

	// Protected routes
	grp := r.Group("/")
	grp.Use(rlmw, authRequired)

	// Users
	grp.GET("/users", middleware.RequireAdmin(), pair.Users.List)
	grp.POST("/users", middleware.RequireAdmin(), pair.Users.Create)
	grp.GET("/users/:id", middleware.RequireSelfOrAdmin(), pair.Users.Get)
	grp.PATCH("/users/:id", middleware.RequireSelfOrAdmin(), pair.Users.Patch)
	grp.DELETE("/users/:id", middleware.RequireAdmin(), pair.Users.Delete)
	grp.GET("/users/me", pair.Users.Me)
	grp.PATCH("/users/me", pair.Users.PatchMe)

	// Admin logs endpoint
	grp.GET("/api/logs", middleware.RequireAdmin(), logsHandler.ListRecent)

	return r
}

// pair holds both handlers for convenience.
type pair struct{ Auth *handlers.AuthHandler; Users *handlers.UserHandler }

// handlersFrom constructs both core handlers.
func handlersFrom(a *service.AuthService, u *service.UserService) pair {
	return pair{Auth: handlers.NewAuthHandler(a), Users: handlers.NewUserHandler(u)}
}

// requestLogger writes a structured access log (zap) and also enqueues a Redis LogEntry.
// It runs AFTER the handler (c.Next) to capture status + duration.
func requestLogger(log *zap.Logger, async *rlog.AsyncLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		dur := time.Since(start)
		reqID := c.GetString("req.id")
		ip := c.ClientIP()
		uid, _ := c.Get("auth.sub")
		role, _ := c.Get("auth.role")

		// 1) Primary log to file/console
		log.Info("http",
			zap.String("id", reqID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("dur", dur),
			zap.String("ip", ip),
			zap.Stringer("user", toStringer(uid)),
			zap.Stringer("role", toStringer(role)),
		)

		// 2) Async to Redis
		if async != nil {
			async.Enqueue(rlog.LogEntry{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Level:     "INFO",
				Message:   "http",
				Context: map[string]any{
					"requestId": reqID,
					"method":    c.Request.Method,
					"path":      c.Request.URL.Path,
					"status":    c.Writer.Status(),
					"duration":  dur.String(),
					"ip":        ip,
					"userID":    uid,
					"role":      role,
				},
			})
		}
	}
}

// toStringer is a small helper to safely render interface values as strings in logs.
type stringer struct{ v any }

func (s stringer) String() string {
	if s.v == nil {
		return ""
	}
	if x, ok := s.v.(string); ok {
		return x
	}
	return ""
}

func toStringer(v any) stringer { return stringer{v: v} }
