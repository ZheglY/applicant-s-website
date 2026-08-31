package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestApplicationStack(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("set RUN_INTEGRATION_TESTS=1 to run destructive integration tests against dedicated services")
	}
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	redisURL := os.Getenv("TEST_REDIS_URL")
	if databaseURL == "" || redisURL == "" {
		t.Fatal("TEST_DATABASE_URL and TEST_REDIS_URL are required")
	}

	ctx := context.Background()
	logger := zap.NewNop()
	pool, err := database.Open(ctx, databaseURL, logger)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err = database.Migrate(ctx, pool, logger); err != nil {
		t.Fatal(err)
	}
	repo := repository.New(pool)
	if err = repo.ResetApplicationData(ctx); err != nil {
		t.Fatal(err)
	}

	redisCache, err := cache.Open(ctx, redisURL, "unik:test:", logger)
	if err != nil {
		t.Fatal(err)
	}
	defer redisCache.Close()
	_ = redisCache.Delete(ctx, cache.KeyDirections, cache.KeyNews, cache.KeySummary)

	cfg := config.Config{
		Environment: "test", WebDir: filepath.Join("..", "..", "app"),
		SecretKey: "01234567890123456789012345678901", AccessTokenTTL: time.Hour,
		DirectionsCacheTTL: time.Minute, NewsCacheTTL: time.Minute, SummaryCacheTTL: time.Minute,
		LoginRateLimit: 20, LoginRateWindow: time.Minute,
		RegistrationRateLimit: 20, RegistrationRateWindow: time.Minute, RateLimitFailOpen: false,
		DefaultAdmissionsLogin: "admin@unik.edu", DefaultAdmissionsPassword: "admin",
		DefaultAnalystLogin: "prepod@unik.edu", DefaultAnalystPassword: "123456",
	}
	tokens := auth.NewManager(cfg.SecretKey, cfg.AccessTokenTTL)
	authService := service.NewAuthService(repo, tokens, cfg)
	if err = authService.Bootstrap(ctx); err != nil {
		t.Fatal(err)
	}
	metrics := observability.NewMetrics()
	app, err := httpserver.New(cfg, logger, repo, authService, tokens, redisCache, metrics)
	if err != nil {
		t.Fatal(err)
	}
	httpTestServer := httptest.NewServer(app.Handler())
	defer httpTestServer.Close()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 5 * time.Second}
	assertStatus(t, client, http.MethodGet, httpTestServer.URL+"/health/ready", nil, http.StatusOK)
	assertStatus(t, client, http.MethodGet, httpTestServer.URL+"/auth/register", nil, http.StatusOK)

	registration := map[string]any{
		"fullname": "Тестов Иван Иванович", "password": "student123", "password_confirm": "student123",
		"birthdate": "2007-03-15", "phone": "+7 (900) 000-00-00", "email": "integration.student@example.com",
		"telegram": "@integration", "school": "Школа № 1", "achievements": "Золотая медаль",
		"priorities": []string{"Программная инженерия"}, "agreement": true,
		"ege_scores": map[string]string{"Математика": "90", "Информатика": "88", "Русский язык": "91"},
	}
	response := doJSON(t, client, http.MethodPost, httpTestServer.URL+"/auth/register", registration)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("registration failed: %d %s", response.StatusCode, readBody(response))
	}
	var registered struct {
		ID int64 `json:"id"`
	}
	decodeResponse(t, response, &registered)

	login := doJSON(t, client, http.MethodPost, httpTestServer.URL+"/auth/login", map[string]string{
		"username": "admin@unik.edu", "password": "admin",
	})
	if login.StatusCode != http.StatusOK {
		t.Fatalf("admin login failed: %d %s", login.StatusCode, readBody(login))
	}
	_ = login.Body.Close()

	created := doJSON(t, client, http.MethodPost, httpTestServer.URL+"/users/news", map[string]string{
		"title": "Интеграционный тест", "subtitle": "Redis", "text": "Проверка кеша и API",
	})
	if created.StatusCode != http.StatusOK {
		t.Fatalf("create news failed: %d %s", created.StatusCode, readBody(created))
	}
	_ = created.Body.Close()
	assertStatus(t, client, http.MethodGet, httpTestServer.URL+"/users/news/data", nil, http.StatusOK)
	assertStatus(t, client, http.MethodGet, httpTestServer.URL+"/users/news/data", nil, http.StatusOK)
	assertStatus(t, client, http.MethodGet, httpTestServer.URL+"/users/stats", nil, http.StatusOK)

	directions, err := repo.ListDirections(ctx)
	if err != nil || len(directions) == 0 {
		t.Fatalf("directions unavailable: %v", err)
	}
	status := doJSON(t, client, http.MethodPatch,
		httpTestServer.URL+"/users/applicants/"+jsonNumber(registered.ID)+"/status",
		map[string]any{"direction_id": directions[0].ID, "status": "accepted"},
	)
	if status.StatusCode != http.StatusOK {
		t.Fatalf("status update failed: %d %s", status.StatusCode, readBody(status))
	}
	_ = status.Body.Close()

	metricsResponse, err := client.Get(httpTestServer.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	metricsBody := readBody(metricsResponse)
	if metricsResponse.StatusCode != http.StatusOK || !strings.Contains(metricsBody, `unik_cache_operations_total{key="cache:news",result="hit"}`) {
		t.Fatalf("cache hit metric missing: status=%d", metricsResponse.StatusCode)
	}
}

func assertStatus(t *testing.T, client *http.Client, method, url string, body io.Reader, expected int) {
	t.Helper()
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != expected {
		t.Fatalf("%s %s: expected %d, got %d: %s", method, url, expected, response.StatusCode, readBody(response))
	}
}

func doJSON(t *testing.T, client *http.Client, method, url string, value any) *http.Response {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func decodeResponse(t *testing.T, response *http.Response, target any) {
	t.Helper()
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatal(err)
	}
}

func readBody(response *http.Response) string {
	data, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	return string(data)
}

func jsonNumber(value int64) string {
	data, _ := json.Marshal(value)
	return string(data)
}
