package httpx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"colorpixel/internal/logger"
	"colorpixel/internal/timeutil"
)

func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, code: 200}
		next.ServeHTTP(ww, r)
		logger.L().Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.code,
			"ms", time.Since(start).Milliseconds(),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(c int) {
	w.code = c
	w.ResponseWriter.WriteHeader(c)
}

func cors(origins []string) func(http.Handler) http.Handler {
	allow := map[string]struct{}{}
	for _, o := range origins {
		allow[o] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}
			if _, ok := allow[origin]; ok || sameOrigin(origin, r.Host) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func sameOrigin(origin, host string) bool {
	origin = strings.TrimPrefix(origin, "http://")
	origin = strings.TrimPrefix(origin, "https://")
	return strings.EqualFold(origin, host)
}

type claims struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
}

func signToken(secret, user string) string {
	c := claims{Sub: user, Exp: timeutil.Now().Add(24 * time.Hour).Unix()}
	body, _ := json.Marshal(c)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(body) + "~" + base64.RawURLEncoding.EncodeToString(sig)
}

func parseToken(secret, tok string) (string, bool) {
	i := strings.LastIndex(tok, "~")
	if i < 0 {
		return "", false
	}
	body, err := base64.RawURLEncoding.DecodeString(tok[:i])
	if err != nil {
		return "", false
	}
	sig, err := base64.RawURLEncoding.DecodeString(tok[i+1:])
	if err != nil {
		return "", false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	if !hmac.Equal(mac.Sum(nil), sig) {
		return "", false
	}
	var c claims
	if json.Unmarshal(body, &c) != nil {
		return "", false
	}
	if timeutil.Now().Unix() > c.Exp {
		return "", false
	}
	return c.Sub, true
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			writeErr(w, http.StatusUnauthorized, "auth", "missing bearer token")
			return
		}
		user, ok := parseToken(s.Cfg.JWTSecret, strings.TrimPrefix(h, "Bearer "))
		if !ok {
			writeErr(w, http.StatusUnauthorized, "auth", "invalid token")
			return
		}
		r.Header.Set("X-User", user)
		next.ServeHTTP(w, r)
	})
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
