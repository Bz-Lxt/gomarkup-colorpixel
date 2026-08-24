package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func Open(ctx context.Context, url string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 16
	cfg.MaxConnIdleTime = 2 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	// Ping with the *cancellable* startup context so a shutdown signal
	// (SIGTERM/SIGINT) can abort a handshake that is stuck waiting for the
	// PostgreSQL startup reply. pgxpool.Acquire returns ctx.Err() promptly,
	// and pool.Close() below cancels the pool's base context which in turn
	// unblocks any in-flight connection constructor still reading the wire.
	// Using context.WithoutCancel here would defeat cancellation and wedge
	// the process until the platform forcibly sends SIGKILL.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &DB{Pool: pool}, nil
}

func (d *DB) Close() {
	if d != nil && d.Pool != nil {
		d.Pool.Close()
	}
}
