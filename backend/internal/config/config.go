package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	Port string
	Host string
	
	// Database
	DatabaseURL string
	
	// Redis
	RedisURL string
	
	// JWT
	JWTSecret    string
	JWTIssuer    string
	JWTExpiry    time.Duration
	RefreshExpiry time.Duration
	
	// CORS
	CORSOrigins []string
	
	// Rate Limiting
	RateLimitRequests int
	RateLimitWindow   time.Duration
	
	// DSN Host for project DSNs
	DSNHost string
	
	// Email (for future use)
	SMTPHost string
	SMTPPort int
	EmailFrom string
}

func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "8080"),
		Host: getEnv("HOST", "0.0.0.0"),
		
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/minisentry?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		
		JWTSecret:     getEnv("JWT_SECRET", "your-256-bit-secret-change-in-production"),
		JWTIssuer:     getEnv("JWT_ISSUER", "minisentry"),
		JWTExpiry:     getDurationEnv("JWT_EXPIRY", 15*time.Minute),
		RefreshExpiry: getDurationEnv("REFRESH_EXPIRY", 7*24*time.Hour),
		
		CORSOrigins: []string{
			getEnv("FRONTEND_URL", "http://localhost:3000"),
		},
		
		RateLimitRequests: getIntEnv("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getDurationEnv("RATE_LIMIT_WINDOW", time.Minute),
		
		DSNHost: getEnv("DSN_HOST", "api.minisentry.com"),
		
		SMTPHost:  getEnv("SMTP_HOST", ""),
		SMTPPort:  getIntEnv("SMTP_PORT", 587),
		EmailFrom: getEnv("EMAIL_FROM", "noreply@minisentry.local"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}