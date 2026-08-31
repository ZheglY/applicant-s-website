package cache

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"go.uber.org/zap"
)

func TestJSONCacheLifecycle(t *testing.T) {
	server := miniredis.RunT(t)
	store, err := Open(context.Background(), fmt.Sprintf("redis://%s/0", server.Addr()), "test:", zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	type payload struct {
		Name string
		ID   int
	}
	var value payload
	if err = store.GetJSON(context.Background(), "item", &value); !errors.Is(err, ErrMiss) {
		t.Fatalf("expected cache miss, got %v", err)
	}
	if err = store.SetJSON(context.Background(), "item", payload{Name: "Unik", ID: 7}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err = store.GetJSON(context.Background(), "item", &value); err != nil {
		t.Fatal(err)
	}
	if value.Name != "Unik" || value.ID != 7 {
		t.Fatalf("unexpected cached value: %+v", value)
	}
	if err = store.Delete(context.Background(), "item"); err != nil {
		t.Fatal(err)
	}
	if err = store.GetJSON(context.Background(), "item", &value); !errors.Is(err, ErrMiss) {
		t.Fatalf("expected cache miss after delete, got %v", err)
	}
}

func TestRateLimiter(t *testing.T) {
	server := miniredis.RunT(t)
	store, err := Open(context.Background(), fmt.Sprintf("redis://%s/0", server.Addr()), "test:", zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	for attempt := 1; attempt <= 2; attempt++ {
		allowed, _, allowErr := store.Allow(context.Background(), "login:client", 2, time.Minute)
		if allowErr != nil || !allowed {
			t.Fatalf("attempt %d should be allowed: allowed=%v err=%v", attempt, allowed, allowErr)
		}
	}
	allowed, retryAfter, err := store.Allow(context.Background(), "login:client", 2, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if allowed || retryAfter <= 0 {
		t.Fatalf("third attempt must be rejected: allowed=%v retry=%v", allowed, retryAfter)
	}
}
