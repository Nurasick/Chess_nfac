//go:build integration

package integration

import (
	"sync/atomic"
	"time"
)

var counter uint64

func uniqueSuffix() int64 {
	return time.Now().UnixNano() + int64(atomic.AddUint64(&counter, 1))
}
