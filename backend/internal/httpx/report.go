package httpx

import (
	"net/http"

	"colorpixel/internal/lens"
	"colorpixel/internal/timeutil"
)

func (s *Server) goldenLens(w http.ResponseWriter, r *http.Request) {
	now := timeutil.Now()
	from := now.AddDate(-1, 0, 0)
	rows, err := s.DB.LensRows(r.Context(), from, now)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db", err.Error())
		return
	}
	rep := lens.Build(rows, now, lens.DefaultWeights())
	writeOK(w, rep)
}
