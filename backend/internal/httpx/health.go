package httpx

import (
	"context"
	"net/http"
	"time"
)

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	st := "ok"
	if err := s.DB.Pool.Ping(ctx); err != nil {
		st = "degraded"
		writeErr(w, http.StatusServiceUnavailable, "db", err.Error())
		return
	}
	js, _ := s.DB.JobStats(context.WithoutCancel(r.Context()))
	writeOK(w, map[string]any{
		"status": st,
		"tz":     "Asia/Shanghai",
		"jobs":   js,
	})
}
