package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/yarik/unik/internal/auth"
	"github.com/yarik/unik/internal/cache"
	"github.com/yarik/unik/internal/config"
	"github.com/yarik/unik/internal/database"
	"github.com/yarik/unik/internal/repository"
	"github.com/yarik/unik/internal/seed"
	"github.com/yarik/unik/internal/service"
)

func main() {
	confirmed := flag.Bool("confirm", false, "confirm destructive replacement of application data")
	flag.Parse()
	if !*confirmed {
		fmt.Fprintln(os.Stderr, "refusing to reset data without --confirm")
		os.Exit(2)
	}
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()
	ctx := context.Background()
	pool, err := database.Open(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("connect to database", zap.Error(err))
	}
	defer pool.Close()
	if err = database.Migrate(ctx, pool, logger); err != nil {
		logger.Fatal("run migrations", zap.Error(err))
	}
	redisCache, err := cache.Open(ctx, cfg.RedisURL, cfg.RedisKeyPrefix, logger)
	if err != nil {
		logger.Fatal("connect to Redis", zap.Error(err))
	}
	defer func() { _ = redisCache.Close() }()
	repo := repository.New(pool)
	authService := service.NewAuthService(repo, auth.NewManager(cfg.SecretKey, cfg.AccessTokenTTL), cfg)
	if err = seed.Run(ctx, repo, authService, logger); err != nil {
		logger.Fatal("seed demo data", zap.Error(err))
	}
	if err = redisCache.Delete(ctx, cache.KeyDirections, cache.KeyNews, cache.KeySummary); err != nil {
		logger.Fatal("invalidate Redis cache", zap.Error(err))
	}
	logger.Info("seed completed", zap.String("summary", seed.SummaryLine()), zap.String("student_password", seed.StudentPassword))
}
