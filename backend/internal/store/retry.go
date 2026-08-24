package store

import (
	"context"
	"time"
)

func OpenRetry(ctx context.Context, url string, attempts int) (*DB, error) {
	var last error
	if attempts < 1 {
		attempts = 1
	}
	for i := 0; i < attempts; i++ {
		db, err := Open(ctx, url)
		if err == nil {
			return db, nil
		}
		last = err
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(i+1) * 400 * time.Millisecond):
		}
	}
	return nil, last
}
