package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"path/filepath"

	"github.com/gofiber/fiber/v3"
	"github.com/hra42/krites/benchmark"
	"github.com/hra42/krites/config"
	"github.com/hra42/krites/database"
	"github.com/hra42/krites/handler"
	"github.com/hra42/krites/middleware"
	"github.com/hra42/krites/openrouter"
	"github.com/hra42/krites/pricing"
	"github.com/hra42/krites/service"
)

func setupLogger(env string) {
	var h slog.Handler
	if env == "production" {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	slog.SetDefault(slog.New(h))
}

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	setupLogger(cfg.Server.Env)

	mgr, err := service.NewServiceManager(cfg.Database.ServicesDir, cfg.Database.MaxConnections)
	if err != nil {
		slog.Error("failed to create service manager", "error", err)
		os.Exit(1)
	}

	// Pre-register services from config (idempotent)
	for _, svcCfg := range cfg.Services {
		if _, err := mgr.Create(context.Background(), svcCfg.ID, svcCfg.Name); err != nil {
			if err == service.ErrServiceExists {
				slog.Debug("service already exists, skipping", "id", svcCfg.ID)
				continue
			}
			slog.Error("failed to create service", "id", svcCfg.ID, "error", err)
			os.Exit(1)
		}
		slog.Info("created service", "id", svcCfg.ID)
	}

	// Create OpenRouter client
	orClient := openrouter.NewClient(cfg.OpenRouter.APIKey, cfg.Server.Env == "development")

	// Create conversation logger
	convLogger := handler.NewConversationLogger(256)

	// Create handlers
	chatHandler := handler.NewChatHandler(orClient, convLogger)
	svcHandler := handler.NewServiceHandler(mgr, cfg.Database.ServicesDir)
	convHandler := handler.NewConversationHandler(mgr)
	healthHandler := handler.NewHealthHandler(mgr)
	modelsHandler := handler.NewModelsHandler(orClient)

	// Rate limiter (created before app so we can close it on shutdown)
	var rateLimiter *middleware.RateLimiter
	if cfg.Server.RateLimit.Enabled {
		rateLimiter = middleware.NewRateLimiter(cfg.Server.RateLimit.RequestsPerSecond)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	app.Use(middleware.CORS())
	app.Use(middleware.RequestLogger())

	if rateLimiter != nil {
		app.Use(middleware.RateLimit(rateLimiter))
	}

	// Health & readiness (outside auth)
	app.Get("/health", healthHandler.HandleHealth)
	app.Get("/ready", healthHandler.HandleReady)

	// Service management routes (no service context middleware)
	services := app.Group("/services", middleware.AdminAuth(cfg.Server.AdminAPIKey))
	services.Post("/", svcHandler.HandleCreate)
	services.Get("/", svcHandler.HandleList)
	services.Get("/:id", svcHandler.HandleGet)
	services.Delete("/:id", svcHandler.HandleDelete)
	services.Get("/:id/conversations", convHandler.HandleList)

	// Models endpoint (no auth required)
	app.Get("/v1/models", modelsHandler.HandleList)

	// Pricing cache for cost estimation
	pricingCache := pricing.NewPricingCache(orClient, 1*time.Hour)
	if err := pricingCache.Start(context.Background()); err != nil {
		slog.Warn("failed to load model pricing, costs will be zero", "error", err)
	}

	// Benchmark persistence (DuckDB)
	if err := os.MkdirAll(filepath.Dir(cfg.Database.BenchmarkDBPath), 0o755); err != nil {
		slog.Error("failed to create benchmark db directory", "error", err)
		os.Exit(1)
	}
	benchmarkDB, err := database.OpenDB(cfg.Database.BenchmarkDBPath, cfg.Database.MaxConnections)
	if err != nil {
		slog.Error("failed to open benchmark database", "error", err)
		os.Exit(1)
	}
	benchmarkStore, err := benchmark.NewDuckDBStore(benchmarkDB)
	if err != nil {
		slog.Error("failed to initialize benchmark store", "error", err)
		os.Exit(1)
	}
	slog.Info("benchmark store initialized", "path", cfg.Database.BenchmarkDBPath)

	// Benchmark routes (no auth required)
	benchmarkHandler := benchmark.NewHandler(benchmarkStore, orClient, pricingCache)
	benchmarkHandler.RegisterRoutes(app)

	// Chat completion routes (require X-Service-ID header)
	v1 := app.Group("/v1", middleware.ServiceContext(mgr))
	v1.Post("/chat/completions", chatHandler.HandleChatCompletion)

	// TODO: Initialize WebhookNotifier and inject into ChatHandler (see docs/EXTENDING.md)
	// TODO: Add prompt management endpoints: POST/GET /services/:id/prompts
	// TODO: Add fine-tuning endpoints: POST/GET /services/:id/fine-tune

	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	// Start server in a goroutine
	go func() {
		slog.Info("starting server", "addr", addr, "env", cfg.Server.Env)
		if err := app.Listen(addr); err != nil {
			slog.Error("server error", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	pricingCache.Stop()
	slog.Info("pricing cache stopped")

	convLogger.Close()
	slog.Info("conversation logger closed")

	if rateLimiter != nil {
		rateLimiter.Close()
	}

	if err := benchmarkDB.Close(); err != nil {
		slog.Error("benchmark db close error", "error", err)
	}

	if err := mgr.Close(); err != nil {
		slog.Error("service manager close error", "error", err)
	}
	slog.Info("shutdown complete")
}
