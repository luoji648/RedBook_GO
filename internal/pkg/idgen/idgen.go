package idgen

import (
	"sync/atomic"
	"time"
)

var seq int64

// NextID returns a non-negative, time-ordered identifier suitable for tables
// that do not use MySQL auto-increment.
func NextID() int64 {
	ts := time.Now().UnixMilli()
	n := atomic.AddInt64(&seq, 1) & 0x3FF
	return ts<<10 | n
}
