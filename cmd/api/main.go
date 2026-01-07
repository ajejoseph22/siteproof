// Package main is the entry point for the SiteProof API server.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ajejoseph22/siteproof/internal/config"
	"github.com/ajejoseph22/siteproof/internal/handler"
	"github.com/ajejoseph22/siteproof/internal/middleware"
	"github.com/ajejoseph22/siteproof/internal/service"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Version is set at build time.
var Version = "dev"

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// Verify database connection
	if err := pool.Ping(ctx); err != nil {
		logger.Error("failed to ping database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("connected to database")

	// Initialize services
	authService := service.NewAuthService(pool, &cfg.JWT)
	userService := service.NewUserService(pool)
	projectService := service.NewProjectService(pool)
	ticketService := service.NewTicketService(pool)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(pool, Version)
	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService)
	ticketHandler := handler.NewTicketHandler(ticketService)

	// Suppress unused variable warning (userService will be used later)
	_ = userService

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "Idempotency-Key"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health endpoints (no auth required)
	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Ready)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authService))

			// Auth
			r.Post("/auth/logout", authHandler.Logout)
			r.Get("/auth/me", authHandler.Me)

			// Projects
			r.Route("/projects", func(r chi.Router) {
				r.Get("/", projectHandler.List)
				r.Post("/", projectHandler.Create)
				r.Get("/{id}", projectHandler.Get)
				r.Patch("/{id}", projectHandler.Update)
				r.Get("/{project_id}/tickets", ticketHandler.ListByProject)
			})

			// Tickets
			r.Route("/tickets", func(r chi.Router) {
				r.Post("/", ticketHandler.Create)
				r.Get("/{id}", ticketHandler.Get)
				r.Patch("/{id}", ticketHandler.Update)
				r.Post("/{id}/submit", ticketHandler.Submit)
				r.Post("/{id}/reject", ticketHandler.Reject)
			})
		})
	})

	// Create server
	srv := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("starting server",
			slog.String("addr", srv.Addr),
			slog.String("version", Version),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("server stopped")
}
