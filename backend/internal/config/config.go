package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr           string
	DatabaseURL        string
	DataDir            string
	PreviewWindowBytes int
	JWTSecret          string
	SampleMode         bool
	LogLevel           string
	CORSOrigins        []string
	PreviewMaxBytes    int
	MaxUploadBytes     int64
	MaxIFDs            int
	MaxIFDDepth        int
	MaxAllocCount      int
}

func Load() Config {
	return Config{
		HTTPAddr:           env("HTTP_ADDR", ":8080"),
		DatabaseURL:        env("DATABASE_URL", "postgres://colorpixel:colorpixel@127.0.0.1:28383/colorpixel?sslmode=disable"),
		DataDir:            env("DATA_DIR", "./var"),
		PreviewWindowBytes: envInt("PREVIEW_WINDOW_BYTES", 16<<20),
		JWTSecret:          env("JWT_SECRET", "colorpixel-dev-secret"),
		SampleMode:         env("SAMPLE_MODE", "1") != "0",
		LogLevel:           env("LOG_LEVEL", "info"),
		CORSOrigins:        splitCSV(env("CORS_ORIGINS", "http://localhost:28381")),
		PreviewMaxBytes:    envInt("PREVIEW_MAX_BYTES", 8<<20),
		MaxUploadBytes:     int64(envInt("MAX_UPLOAD_BYTES", 2<<30)),
		MaxIFDs:            envInt("MAX_IFDS", 64),
		MaxIFDDepth:        envInt("MAX_IFD_DEPTH", 8),
		MaxAllocCount:      envInt("MAX_ALLOC_COUNT", 1_000_000),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
