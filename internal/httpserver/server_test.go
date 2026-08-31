package httpserver

import (
	"path/filepath"
	"runtime"
	"testing"

	"go.uber.org/zap"

	"github.com/yarik/unik/internal/auth"
	"github.com/yarik/unik/internal/config"
	"github.com/yarik/unik/internal/observability"
	"github.com/yarik/unik/internal/repository"
	"github.com/yarik/unik/internal/service"
)

func TestTemplatesParse(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test file")
	}
	cfg := config.Config{WebDir: filepath.Join(filepath.Dir(filename), "..", "..", "app")}
	repo := repository.New(nil)
	tokens := auth.NewManager("01234567890123456789012345678901", 1)
	authService := service.NewAuthService(repo, tokens, cfg)
	if _, err := New(cfg, zap.NewNop(), repo, authService, tokens, nil, observability.NewMetrics()); err != nil {
		t.Fatalf("templates must parse: %v", err)
	}
}
