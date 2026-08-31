package auth

import (
	"errors"
	"testing"
	"time"
)

func TestCreateAndVerify(t *testing.T) {
	m := NewManager("01234567890123456789012345678901", time.Hour)
	now := time.Unix(1_700_000_000, 0)
	m.now = func() time.Time { return now }

	token, err := m.Create(42, "student")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := m.Verify(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 42 || claims.Role != "student" || claims.Subject != "42" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestRejectsTamperedAndExpiredTokens(t *testing.T) {
	m := NewManager("01234567890123456789012345678901", time.Second)
	now := time.Unix(1_700_000_000, 0)
	m.now = func() time.Time { return now }
	token, err := m.Create(1, "student")
	if err != nil {
		t.Fatal(err)
	}

	if _, err = m.Verify(token + "x"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected invalid token, got %v", err)
	}
	m.now = func() time.Time { return now.Add(2 * time.Second) }
	if _, err = m.Verify(token); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("expected expired token, got %v", err)
	}
}
