package lock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

func randomToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

var memoryLocks sync.Map

type memoryGuard struct {
	mu sync.Mutex
}

type Locker interface {
	Lock(ctx context.Context, key string, ttl time.Duration) (Unlock, error)
}

type Unlock func()

type RedisLocker struct {
	client *redis.Client
}

func NewRedisLocker(client *redis.Client) *RedisLocker {
	return &RedisLocker{client: client}
}

func (l *RedisLocker) Lock(ctx context.Context, key string, ttl time.Duration) (Unlock, error) {
	if l == nil || l.client == nil {
		return memoryLock(key, ttl)
	}
	token := randomToken()
	ok, err := l.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("lock already held")
	}
	return func() {
		releaseScript.Run(ctx, l.client, []string{key}, token)
	}, nil
}

func memoryLock(key string, ttl time.Duration) (Unlock, error) {
	actual, _ := memoryLocks.LoadOrStore(key, &memoryGuard{})
	guard := actual.(*memoryGuard)
	if !guard.mu.TryLock() {
		return nil, errors.New("lock already held")
	}
	return func() { guard.mu.Unlock() }, nil
}

var releaseScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
end
return 0
`)
