package auth

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

const (
	CurrentUserKey = "currentUser"
	blacklistKey   = "redbook:jwt:blacklist:"
)

var memoryBlacklist sync.Map

type Claims struct {
	UserID int64 `json:"userId"`
	TV     int32 `json:"tv"`
	jwt.RegisteredClaims
}

type TokenService struct {
	secret        []byte
	expireSeconds int64
	redis         *redis.Client
}

func NewTokenService(secret string, expireSeconds int64, redis *redis.Client) *TokenService {
	return &TokenService{secret: []byte(secret), expireSeconds: expireSeconds, redis: redis}
}

func (t *TokenService) CreateToken(userID int64, tokenVersion int32) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		TV:     tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			ID:        randomHex(16),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(t.expireSeconds) * time.Second)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}

func (t *TokenService) ParseToken(authorization string) (*Claims, error) {
	if authorization == "" {
		return nil, jwt.ErrTokenRequiredClaimMissing
	}
	token := strings.TrimSpace(authorization)
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	if token == "" {
		return nil, jwt.ErrTokenRequiredClaimMissing
	}
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (any, error) {
		return t.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

func (t *TokenService) IsBlacklisted(ctx context.Context, jti string) bool {
	if jti == "" {
		return false
	}
	if t.redis == nil {
		_, ok := memoryBlacklist.Load(jti)
		return ok
	}
	n, err := t.redis.Exists(ctx, blacklistKey+jti).Result()
	return err == nil && n > 0
}

func (t *TokenService) Blacklist(ctx context.Context, jti string, ttl time.Duration) {
	if t.redis == nil {
		memoryBlacklist.Store(jti, time.Now().Add(ttl))
		return
	}
	_ = t.redis.Set(ctx, blacklistKey+jti, "1", ttl).Err()
}

func EncodePassword(password string) string {
	salt := randomHex(10)
	return encodeWithSalt(password, salt)
}

func MatchesPassword(encoded, raw string) bool {
	if encoded == "" || raw == "" {
		return false
	}
	idx := strings.Index(encoded, "@")
	if idx <= 0 {
		return false
	}
	salt := encoded[:idx]
	return subtleEqual(encoded, encodeWithSalt(raw, salt))
}

func encodeWithSalt(password, salt string) string {
	sum := md5.Sum([]byte(password + salt))
	return salt + "@" + hex.EncodeToString(sum[:])
}

func subtleEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
