package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"colorpixel/internal/config"
	"colorpixel/internal/httpx"
	"colorpixel/internal/jobs"
	"colorpixel/internal/logger"
	"colorpixel/internal/seed"
	"colorpixel/internal/store"
)

func main() {
	health := flag.Bool("healthcheck", false, "probe local health and exit")
	flag.Parse()
	cfg := config.Load()
	logger.Init(cfg.LogLevel, os.Stdout)

	if *health {
		os.Exit(probe(cfg.HTTPAddr))
	}

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		logger.L().Error("data dir", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := store.OpenRetry(ctx, cfg.DatabaseURL, 8)
	if err != nil {
		logger.L().Error("db open", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := db.Migrate(ctx); err != nil {
		logger.L().Error("migrate", "err", err)
		os.Exit(1)
	}
	if err := seed.Bootstrap(ctx, cfg, db); err != nil {
		logger.L().Error("seed", "err", err)
		os.Exit(1)
	}

	w := &jobs.Worker{DB: db, DataDir: cfg.DataDir}
	go w.Start(ctx)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpx.New(cfg, db, w),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Minute,
		WriteTimeout:      30 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		logger.L().Info("listen", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L().Error("http", "err", err)
			stop()
		}
	}()
	<-ctx.Done()
	sh, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	_ = srv.Shutdown(sh)
}

func probe(addr string) int {
	if addr == "" {
		addr = ":8080"
	}
	host := "127.0.0.1"
	_, port, err := net.SplitHostPort(addr)
	if err == nil {
		addr = net.JoinHostPort(host, port)
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + addr + "/api/v1/health")
	if err != nil || resp.StatusCode >= 400 {
		return 1
	}
	return 0
}
