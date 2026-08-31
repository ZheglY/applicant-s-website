package httpserver

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

func (s *Server) enforceRateLimit(
	w http.ResponseWriter,
	r *http.Request,
	scope, identity string,
	limit int,
	window time.Duration,
) bool {
	if s.cache == nil {
		return true
	}
	digest := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(identity))))
	key := scope + ":" + hex.EncodeToString(digest[:12])
	allowed, retryAfter, err := s.cache.Allow(r.Context(), key, limit, window)
	if err != nil {
		s.recordRateLimit(scope, "error")
		s.logger.Warn("Redis rate limiter failed", zap.String("scope", scope), zap.Error(err))
		if s.cfg.RateLimitFailOpen {
			return true
		}
		writeError(w, http.StatusServiceUnavailable, "Rate limiter is unavailable")
		return false
	}
	if allowed {
		s.recordRateLimit(scope, "allow")
		return true
	}
	s.recordRateLimit(scope, "reject")
	seconds := max(1, int(math.Ceil(retryAfter.Seconds())))
	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	writeError(w, http.StatusTooManyRequests, "Too many requests. Try again later")
	return false
}

func (s *Server) recordRateLimit(scope, decision string) {
	if s.metrics != nil {
		s.metrics.RateLimits.WithLabelValues(scope, decision).Inc()
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
