package unit

import (
	"context"
	"testing"
	"time"

	"redbook-go/internal/pkg/lock"
)

func TestMemoryLockExclusive(t *testing.T) {
	unlock, err := lock.NewRedisLocker(nil).Lock(context.Background(), "key", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lock.NewRedisLocker(nil).Lock(context.Background(), "key", time.Second); err == nil {
		t.Fatal("expected second lock to fail")
	}
	unlock()
	unlock2, err := lock.NewRedisLocker(nil).Lock(context.Background(), "key", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	unlock2()
}
