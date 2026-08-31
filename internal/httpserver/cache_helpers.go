package httpserver

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/yarik/unik/internal/cache"
	"github.com/yarik/unik/internal/domain"
)

func (s *Server) cachedDirections(ctx context.Context) ([]domain.Direction, error) {
	var result []domain.Direction
	if s.readCache(ctx, cache.KeyDirections, &result) {
		return result, nil
	}
	result, err := s.repo.ListDirections(ctx)
	if err != nil {
		return nil, err
	}
	s.writeCache(ctx, cache.KeyDirections, result, s.cfg.DirectionsCacheTTL)
	return result, nil
}

func (s *Server) cachedNews(ctx context.Context) ([]domain.News, error) {
	var result []domain.News
	if s.readCache(ctx, cache.KeyNews, &result) {
		return result, nil
	}
	result, err := s.repo.ListNews(ctx)
	if err != nil {
		return nil, err
	}
	s.writeCache(ctx, cache.KeyNews, result, s.cfg.NewsCacheTTL)
	return result, nil
}

func (s *Server) cachedSummary(ctx context.Context) (domain.Summary, error) {
	var result domain.Summary
	if s.readCache(ctx, cache.KeySummary, &result) {
		return result, nil
	}
	result, err := s.repo.Summary(ctx)
	if err != nil {
		return domain.Summary{}, err
	}
	s.writeCache(ctx, cache.KeySummary, result, s.cfg.SummaryCacheTTL)
	return result, nil
}

func (s *Server) readCache(ctx context.Context, key string, target any) bool {
	if s.cache == nil {
		return false
	}
	err := s.cache.GetJSON(ctx, key, target)
	switch {
	case err == nil:
		s.recordCache(key, "hit")
		return true
	case errors.Is(err, cache.ErrMiss):
		s.recordCache(key, "miss")
	default:
		s.recordCache(key, "error")
		s.logger.Warn("Redis cache read failed", zap.String("key", key), zap.Error(err))
	}
	return false
}

func (s *Server) writeCache(ctx context.Context, key string, value any, ttl time.Duration) {
	if s.cache == nil {
		return
	}
	if err := s.cache.SetJSON(ctx, key, value, ttl); err != nil {
		s.recordCache(key, "error")
		s.logger.Warn("Redis cache write failed", zap.String("key", key), zap.Error(err))
		return
	}
	s.recordCache(key, "write")
}

func (s *Server) invalidateCache(ctx context.Context, keys ...string) {
	if s.cache == nil {
		return
	}
	if err := s.cache.Delete(ctx, keys...); err != nil {
		s.recordCache("multiple", "error")
		s.logger.Warn("Redis cache invalidation failed", zap.Strings("keys", keys), zap.Error(err))
		return
	}
	for _, key := range keys {
		s.recordCache(key, "invalidate")
	}
}

func (s *Server) recordCache(key, result string) {
	if s.metrics != nil {
		s.metrics.CacheOperations.WithLabelValues(key, result).Inc()
	}
}
