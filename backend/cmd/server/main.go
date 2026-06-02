package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"minisentry/internal/config"
	"minisentry/internal/database"
	"minisentry/internal/handlers"
	"minisentry/internal/middleware"
	"minisentry/internal/services"

	"github.com/go-chi/chi/v5"
)

func main() {
	// Structured JSON logger
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg := config.Load()
	slog.Info("starting minisentry", "host", cfg.Host, "port", cfg.Port)

	// Database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Services
	jwtService, err := services.NewJWTService(cfg.JWTIssuer, cfg.JWTExpiry, cfg.RefreshExpiry)
	if err != nil {
		slog.Error("jwt init failed", "error", err)
		os.Exit(1)
	}

	passwordService := services.NewDefaultPasswordService()
	userService := services.NewUserService(db, passwordService)
	organizationService := services.NewOrganizationService(db)
	projectService := services.NewProjectService(db, cfg.DSNHost)
	errorService := services.NewErrorService(db)
	issueService := services.NewIssueService(db.DB)
	statsHandler := handlers.NewStatsHandler(db.DB)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	orgMiddleware := middleware.NewOrganizationMiddleware(organizationService)
	projectMiddleware := middleware.NewProjectMiddleware(projectService)

	// Handlers
	userHandler := handlers.NewUserHandler(userService, jwtService)
	orgHandler := handlers.NewOrganizationHandler(organizationService)
	projectHandler := handlers.NewProjectHandler(projectService)
	errorHandler := handlers.NewErrorHandler(errorService)
	issueHandler := handlers.NewIssueHandler(issueService)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RecoveryMiddleware)
	r.Use(middleware.RequestIDMiddleware)
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.SecurityMiddleware)
	r.Use(middleware.CORSMiddleware(cfg.CORSOrigins))
	r.Use(middleware.ContentTypeMiddleware)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		sqlDB, err := db.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy"})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "healthy",
			"version":  "1.1.0",
			"database": "connected",
		})
	})

	r.Get("/api/version", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"version": "1.1.0",
			"name":    "minisentry-api",
		})
	})

	// Routes
	errorHandler.RegisterRoutes(r, projectMiddleware)

	r.Route("/api/v1", func(r chi.Router) {
		userHandler.RegisterRoutes(r, authMiddleware)
		orgHandler.RegisterRoutes(r, authMiddleware, orgMiddleware)
		projectHandler.RegisterRoutes(r, authMiddleware, orgMiddleware, projectMiddleware)
		issueHandler.RegisterRoutes(r, authMiddleware, projectMiddleware)

		// Stats overview (authenticated)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			statsHandler.RegisterRoutes(r)
		})
	})

	// 404 / 405
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	})

	// Print clean startup table
	printStartupBanner(cfg)

	// Graceful HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Signal channel
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	db.Close()
	slog.Info("server stopped cleanly")
}

func printStartupBanner(cfg *config.Config) {
	div := strings.Repeat("─", 60)
	fmt.Printf("\n%s\n", div)
	fmt.Printf("  MiniSentry v1.1.0 — ready\n")
	fmt.Printf("  HTTP  : %s:%s\n", cfg.Host, cfg.Port)
	fmt.Printf("  Health: /health  |  Version: /api/version\n")
	fmt.Printf("  Stats : /api/v1/stats/overview (auth required)\n")
	fmt.Printf("%s\n\n", div)
}
