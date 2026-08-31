package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"go.uber.org/zap"

	"github.com/yarik/unik/internal/auth"
	"github.com/yarik/unik/internal/cache"
	"github.com/yarik/unik/internal/config"
	"github.com/yarik/unik/internal/database"
	"github.com/yarik/unik/internal/httpserver"
	"github.com/yarik/unik/internal/observability"
	"github.com/yarik/unik/internal/repository"
	"github.com/yarik/unik/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger, err := newLogger(cfg.Environment)
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Open(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("connect to database", zap.Error(err))
	}
	defer pool.Close()
	if err = database.Migrate(ctx, pool, logger); err != nil {
		logger.Fatal("run database migrations", zap.Error(err))
	}
	redisCache, err := cache.Open(ctx, cfg.RedisURL, cfg.RedisKeyPrefix, logger)
	if err != nil {
		logger.Fatal("connect to Redis", zap.Error(err))
	}
	defer func() { _ = redisCache.Close() }()

	repo := repository.New(pool)
	tokens := auth.NewManager(cfg.SecretKey, cfg.AccessTokenTTL)
	authService := service.NewAuthService(repo, tokens, cfg)
	if err = authService.Bootstrap(ctx); err != nil {
		logger.Fatal("bootstrap application data", zap.Error(err))
	}
	if err = redisCache.Delete(ctx, cache.KeyDirections); err != nil {
		logger.Warn("invalidate directions cache", zap.Error(err))
	}
	metrics := observability.NewMetrics()
	handler, err := httpserver.New(cfg, logger, repo, authService, tokens, redisCache, metrics)
	if err != nil {
		logger.Fatal("initialize HTTP server", zap.Error(err))
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("HTTP server started", zap.String("address", cfg.HTTPAddr), zap.String("environment", cfg.Environment))
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err = <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err = server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	} else {
		logger.Info("HTTP server stopped")
	}
}

func newLogger(environment string) (*zap.Logger, error) {
	if environment == "development" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
