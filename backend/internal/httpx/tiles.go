package httpx

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"image/jpeg"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"

	"colorpixel/internal/metrics"
	"colorpixel/internal/tiles"
)

func (s *Server) tile(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id", "invalid id")
		return
	}
	z, _ := strconv.Atoi(chi.URLParam(r, "z"))
	x, _ := strconv.Atoi(chi.URLParam(r, "x"))
	y, _ := strconv.Atoi(chi.URLParam(r, "y"))
	if z < 0 || x < 0 || y < 0 || z > 16 {
		writeErr(w, http.StatusBadRequest, "tile", "invalid tile coord")
		return
	}
	a, err := s.DB.GetAsset(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not_found", "asset not found")
		return
	}
	p := tiles.Path(s.tileDir(a.ID), z, x, y)
	b, err := os.ReadFile(p)
	if err != nil {
		writeErr(w, http.StatusNotFound, "tile", "tile not ready")
		return
	}
	sum := sha1.Sum(b)
	etag := `"` + hex.EncodeToString(sum[:8]) + `"`
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	if match := r.Header.Get("If-None-Match"); match == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func (s *Server) thumb(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id", "invalid id")
		return
	}
	a, err := s.DB.GetAsset(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not_found", "asset not found")
		return
	}
	p := s.tileDir(a.ID) + "/thumb.jpg"
	if _, err := os.Stat(p); err != nil {
		if a.PreviewPath != "" {
			http.ServeFile(w, r, a.PreviewPath)
			return
		}
		writeErr(w, http.StatusNotFound, "thumb", "not ready")
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, p)
}

func (s *Server) histogram(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id", "invalid id")
		return
	}
	a, err := s.DB.GetAsset(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not_found", "asset not found")
		return
	}
	if a.PreviewPath == "" {
		writeErr(w, http.StatusNotFound, "preview", "no preview")
		return
	}
	b, err := os.ReadFile(a.PreviewPath)
	if err != nil {
		writeErr(w, http.StatusNotFound, "preview", err.Error())
		return
	}
	img, err := jpeg.Decode(bytes.NewReader(b))
	if err != nil {
		writeErr(w, http.StatusUnprocessableEntity, "jpeg", err.Error())
		return
	}
	sc := metrics.Analyze(img)
	writeOK(w, map[string]any{
		"r": sc.HistR, "g": sc.HistG, "b": sc.HistB, "y": sc.HistY,
		"clip_shadow": sc.ClipShadow, "clip_highlight": sc.ClipHighlight,
	})
}

func (s *Server) tileDir(id int64) string {
	return s.Cfg.DataDir + "/tiles/" + strconv.FormatInt(id, 10)
}
