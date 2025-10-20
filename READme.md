# API Gateway (Go 1.22+, Gin, JWT) — Routing · Rate Limit · AuthN/Z


Production‑minded starter that fronts a User Management domain with clean layering.


## Features
- **Routing** with Gin, request ID, access logs, CORS, health, metrics
- **AuthN/Z** via JWT (HS256), RBAC (admin/user), self‑access enforcement
- **Rate Limiting** per‑IP / per‑user (in‑memory token bucket or Redis)
- **DB‑agnostic repos** via GORM (sqlite/mysql/postgres) selected from config
- **Logging** with Zap (JSON, levels, sampling)


## Run (Local)
```bash
# 1) Set JWT secret (override config)
export JWT_SECRET=dev-change-me


# 2) Start API
go run ./cmd/gateway