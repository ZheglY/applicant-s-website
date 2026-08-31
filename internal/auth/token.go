package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	Subject string `json:"sub"`
	UserID  int64  `json:"user_id"`
	Role    string `json:"role"`
	Issued  int64  `json:"iat"`
	Expires int64  `json:"exp"`
}

type Manager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl, now: time.Now}
}

func (m *Manager) Create(userID int64, role string) (string, error) {
	now := m.now()
	header, err := encodeJSON(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	claims, err := encodeJSON(Claims{
		Subject: strconv.FormatInt(userID, 10),
		UserID:  userID,
		Role:    role,
		Issued:  now.Unix(),
		Expires: now.Add(m.ttl).Unix(),
	})
	if err != nil {
		return "", err
	}
	payload := header + "." + claims
	return payload + "." + m.sign(payload), nil
}

func (m *Manager) Verify(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}
	payload := parts[0] + "." + parts[1]
	expected, err := base64.RawURLEncoding.DecodeString(m.sign(payload))
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(actual, expected) {
		return Claims{}, ErrInvalidToken
	}

	var claims Claims
	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || json.Unmarshal(data, &claims) != nil || claims.UserID <= 0 || claims.Role == "" {
		return Claims{}, ErrInvalidToken
	}
	if m.now().Unix() >= claims.Expires {
		return Claims{}, ErrExpiredToken
	}
	return claims, nil
}

func (m *Manager) sign(payload string) string {
	hash := hmac.New(sha256.New, m.secret)
	_, _ = hash.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
}

func encodeJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}
