package httpserver

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/yarik/unik/internal/auth"
	"github.com/yarik/unik/internal/cache"
	"github.com/yarik/unik/internal/config"
	"github.com/yarik/unik/internal/domain"
	"github.com/yarik/unik/internal/observability"
	"github.com/yarik/unik/internal/repository"
	"github.com/yarik/unik/internal/service"
)

const accessCookieName = "access_token"

type Server struct {
	cfg         config.Config
	logger      *zap.Logger
	repo        *repository.Repository
	authService *service.AuthService
	tokens      *auth.Manager
	cache       *cache.Redis
	metrics     *observability.Metrics
	templates   *template.Template
	mux         *http.ServeMux
}

func New(
	cfg config.Config,
	logger *zap.Logger,
	repo *repository.Repository,
	authService *service.AuthService,
	tokens *auth.Manager,
	redisCache *cache.Redis,
	metrics *observability.Metrics,
) (*Server, error) {
	templateDir := filepath.Join(cfg.WebDir, "templates")
	if stat, err := os.Stat(templateDir); err != nil || !stat.IsDir() {
		return nil, fmt.Errorf("template directory %q is unavailable", templateDir)
	}

	funcs := template.FuncMap{
		"isStaff": func(role string) bool {
			return role == domain.RoleAdmissions || role == domain.RoleAnalyst
		},
		"canView": func(current domain.Applicant, applicantID int64) bool {
			return current.Role == domain.RoleAdmissions || current.Role == domain.RoleAnalyst || current.ID == applicantID
		},
		"orText": func(first, second string) string {
			if first != "" {
				return first
			}
			return second
		},
		"truncate": func(value string, limit int) string {
			runes := []rune(value)
			if len(runes) <= limit {
				return value
			}
			return string(runes[:limit])
		},
		"add":             func(a, b int) int { return a + b },
		"priorityNumbers": func() []int { return []int{1, 2, 3} },
		"first": func(items []domain.PopularDirection, limit int) []domain.PopularDirection {
			if len(items) < limit {
				return items
			}
			return items[:limit]
		},
	}
	templates, err := template.New("").Funcs(funcs).ParseGlob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	s := &Server{
		cfg: cfg, logger: logger, repo: repo, authService: authService,
		tokens: tokens, cache: redisCache, metrics: metrics,
		templates: templates, mux: http.NewServeMux(),
	}
	s.routes()
	return s, nil
}

func (s *Server) Handler() http.Handler {
	return s.requestLogger(s.recoverer(s.securityHeaders(s.sameOrigin(s.mux))))
}

func (s *Server) routes() {
	staticDir := filepath.Join(s.cfg.WebDir, "static")
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	s.mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/favicon.svg", http.StatusMovedPermanently)
	})

	s.mux.HandleFunc("GET /", s.handleRoot)
	s.mux.HandleFunc("GET /health/live", s.handleLiveness)
	s.mux.HandleFunc("GET /health/ready", s.handleReadiness)
	s.mux.HandleFunc("GET /api/openapi.yaml", s.handleOpenAPI)
	s.mux.HandleFunc("GET /api/docs", s.handleDocs)
	s.mux.HandleFunc("GET /api/redoc", s.handleReDoc)
	if s.metrics != nil {
		s.mux.Handle("GET /metrics", s.metrics.Handler())
	}

	s.mux.HandleFunc("GET /auth/register", s.handleRegisterPage)
	s.mux.HandleFunc("POST /auth/register", s.handleRegister)
	s.mux.HandleFunc("GET /auth/enter", s.handleLoginPage)
	s.mux.HandleFunc("GET /auth/login", s.handleLoginRedirect)
	s.mux.HandleFunc("POST /auth/login", s.handleLogin)
	s.mux.HandleFunc("GET /auth/logout", s.handleLogout)

	s.mux.HandleFunc("GET /users/news", s.handleNewsPage)
	s.mux.HandleFunc("GET /users/news/data", s.handleNewsData)
	s.mux.HandleFunc("POST /users/news", s.handleCreateNews)
	s.mux.HandleFunc("DELETE /users/news/{newsID}", s.handleDeleteNews)
	s.mux.HandleFunc("GET /users/list", s.handleListPage)
	s.mux.HandleFunc("GET /users/stats", s.handleStatsPage)
	s.mux.HandleFunc("GET /users/applicants/list", s.handleApplicantList)
	s.mux.HandleFunc("GET /users/applicants/{studentID}", s.handleApplicantPage)
	s.mux.HandleFunc("PATCH /users/applicants/{studentID}/status", s.handleUpdateStatus)
}

func (s *Server) render(w http.ResponseWriter, status int, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
		s.logger.Error("render template", zap.String("template", name), zap.Error(err))
	}
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/users/news", http.StatusFound)
}

func (s *Server) handleLiveness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := contextWithTimeout(r, 2*time.Second)
	defer cancel()
	if err := s.repo.Ping(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database is unavailable")
		return
	}
	if s.cache != nil {
		if err := s.cache.Ping(ctx); err != nil {
			writeError(w, http.StatusServiceUnavailable, "Redis is unavailable")
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("api", "openapi.yaml"))
}

func (s *Server) handleDocs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html><head><title>Unik API</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>SwaggerUIBundle({url:'/api/openapi.yaml',dom_id:'#swagger-ui'})</script></body></html>`))
}

func (s *Server) handleReDoc(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html><head><title>Unik API</title><meta charset="utf-8"></head><body><redoc spec-url="/api/openapi.yaml"></redoc><script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script></body></html>`))
}

func cleanText(value string) string { return strings.TrimSpace(value) }
