package config // Centralized configuration types + loader

import (
	"fmt"      // For formatted errors
	"os"       // For env overrides
	"strings"  // Utility for parsing lists

	"gopkg.in/yaml.v3" // YAML parsing
)

// Root holds the entire configuration tree in parent→child nesting.
type Root struct {
	Server    Server    `yaml:"server"`    // HTTP server options
	Security  Security  `yaml:"security"`  // Auth and JWT
	RateLimit RateLimit `yaml:"rate_limit"` // Rate limiting configuration
	Redis     Redis     `yaml:"redis"`     // Redis modes + credentials
	Database  Database  `yaml:"database"`  // DB driver + DSN/config
	Logging   Logging   `yaml:"logging"`   // Zap logging
}

// Server groups HTTP listen + CORS + timeouts.
type Server struct {
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	CORS     CORS     `yaml:"cors"`
	Timeouts Timeouts `yaml:"timeouts"`
}

// CORS defines cross-origin allowlist and methods.
type CORS struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
	AllowCredentials bool   `yaml:"allow_credentials"`
}

// Timeouts for HTTP server.
type Timeouts struct {
	ReadMS       int `yaml:"read_ms"`
	ReadHeaderMS int `yaml:"read_header_ms"`
	WriteMS      int `yaml:"write_ms"`
	IdleMS       int `yaml:"idle_ms"`
}

// Security holds JWT settings.
type Security struct {
	JWT JWT `yaml:"jwt"`
}

// JWT settings.
type JWT struct {
	Issuer     string `yaml:"issuer"`
	Audience   string `yaml:"audience"`
	Secret     string `yaml:"secret"`      // May be overridden by env JWT_SECRET
	TTLMinutes int    `yaml:"ttl_minutes"`
	RefreshTTL int    `yaml:"refresh_ttl_minutes"`
}

// RateLimit config supports memory or redis.
type RateLimit struct {
	Enabled            bool   `yaml:"enabled"`
	Strategy           string `yaml:"strategy"` // memory|redis
	RequestsPerMinute  int    `yaml:"requests_per_minute"`
	Burst              int    `yaml:"burst"`
}

// Redis supports standalone or sentinel modes.
type Redis struct {
	Mode       string   `yaml:"mode"` // standalone|sentinel
	Addresses  []string `yaml:"addresses"` // host:port list (or sentinel nodes)
	MasterName string   `yaml:"master_name"` // sentinel master logical name
	DB         int      `yaml:"db"`
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
	TLS        bool     `yaml:"tls"`
}

// Database driver selection + DSN per driver.
type Database struct {
	Driver string `yaml:"driver"` // mysql|postgres|sqlite
	DSN    string `yaml:"dsn"`    // DSN string per driver
}

// Logging config for zap.
type Logging struct {
	Level    string `yaml:"level"` // debug|info|warn|error
	JSON     bool   `yaml:"json"`
	Sampling bool   `yaml:"sampling"`
}

// Load reads config/config.yaml, applies env overrides, and returns Root.
func Load() (Root, error) {
	var cfg Root // Holder

	raw, err := os.ReadFile("config/config.yaml") // Read YAML file
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err) // File missing or unreadable
	}
	if err := yaml.Unmarshal(raw, &cfg); err != nil { // Parse YAML
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	// Env overrides (minimal: JWT_SECRET)
	if s := os.Getenv("JWT_SECRET"); s != "" {
		cfg.Security.JWT.Secret = s // Override secret from environment
	}
	// Optional: allow comma-separated CORS origins via env CORS_ORIGINS
	if o := os.Getenv("CORS_ORIGINS"); o != "" {
		cfg.Server.CORS.AllowedOrigins = strings.Split(o, ",")
	}
	return cfg, nil // Ready to use
}