package httpx

import (
	"encoding/json"
	"net/http"
)

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid", "json body required")
		return
	}
	if req.Username == "" || req.Password == "" {
		writeErr(w, http.StatusBadRequest, "invalid", "username and password required")
		return
	}
	u, err := s.DB.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "auth", "invalid credentials")
		return
	}
	writeOK(w, map[string]any{
		"token":    signToken(s.Cfg.JWTSecret, u.Username),
		"username": u.Username,
	})
}
