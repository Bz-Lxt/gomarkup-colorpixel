package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"colorpixel/internal/store"
	"colorpixel/internal/timeutil"
)

func (s *Server) listAssets(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := store.AssetFilter{
		Q: q.Get("q"), Camera: q.Get("camera"), Lens: q.Get("lens"), Sort: q.Get("sort"),
	}
	f.Page, _ = strconv.Atoi(q.Get("page"))
	f.PageSize, _ = strconv.Atoi(q.Get("page_size"))
	f.ISOMin, _ = strconv.Atoi(q.Get("iso_min"))
	f.ISOMax, _ = strconv.Atoi(q.Get("iso_max"))
	f.FocalMin, _ = strconv.ParseFloat(q.Get("focal_min"), 64)
	f.FocalMax, _ = strconv.ParseFloat(q.Get("focal_max"), 64)
	f.ApertureMin, _ = strconv.ParseFloat(q.Get("aperture_min"), 64)
	f.ApertureMax, _ = strconv.ParseFloat(q.Get("aperture_max"), 64)
	if v := q.Get("from"); v != "" {
		if t, err := time.ParseInLocation("2006-01-02", v, timeutil.Beijing); err == nil {
			f.From = t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.ParseInLocation("2006-01-02", v, timeutil.Beijing); err == nil {
			f.To = t.Add(24*time.Hour - time.Second)
		}
	}
	items, total, err := s.DB.ListAssets(r.Context(), f)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db", err.Error())
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, a := range items {
		out = append(out, assetDTO(a, false))
	}
	writeOK(w, map[string]any{"items": out, "total": total, "page": max(f.Page, 1), "page_size": f.PageSize})
}

func (s *Server) getAsset(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id", "invalid id")
		return
	}
	a, err := s.DB.GetAsset(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, "not_found", "asset not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db", err.Error())
		return
	}
	writeOK(w, assetDTO(*a, true))
}

func (s *Server) patchAsset(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id", "invalid id")
		return
	}
	var req struct {
		Rating *int     `json:"rating"`
		Tags   []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid", "json required")
		return
	}
	a, err := s.DB.GetAsset(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, "not_found", "asset not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db", err.Error())
		return
	}
	rating := a.Rating
	if req.Rating != nil {
		if *req.Rating < 0 || *req.Rating > 5 {
			writeErr(w, http.StatusBadRequest, "invalid", "rating must be 0-5")
			return
		}
		rating = *req.Rating
	}
	tags := a.Tags
	if req.Tags != nil {
		tags = req.Tags
	}
	if err := s.DB.UpdateRating(r.Context(), id, rating, tags); err != nil {
		writeErr(w, http.StatusInternalServerError, "db", err.Error())
		return
	}
	a, _ = s.DB.GetAsset(r.Context(), id)
	writeOK(w, assetDTO(*a, true))
}

func (s *Server) deleteAsset(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id", "invalid id")
		return
	}
	if err := s.DB.SoftDelete(r.Context(), id); errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, "not_found", "asset not found")
		return
	} else if err != nil {
		writeErr(w, http.StatusInternalServerError, "db", err.Error())
		return
	}
	writeOK(w, map[string]any{"deleted": id})
}

func assetDTO(a store.Asset, full bool) map[string]any {
	m := map[string]any{
		"id":                a.ID,
		"filename":          a.Filename,
		"format":            a.Format,
		"size_bytes":        a.SizeBytes,
		"extraction_mode":   a.ExtractionMode,
		"camera_make":       a.CameraMake,
		"camera_model":      a.CameraModel,
		"lens_model":        a.LensModel,
		"lens_spec":         a.LensSpec,
		"aperture":          a.Aperture,
		"shutter_text":      a.ShutterText,
		"shutter_seconds":   a.ShutterSeconds,
		"iso":               a.ISO,
		"focal_length":      a.FocalLength,
		"focal_length_35mm": a.FocalLength35mm,
		"datetime_original": timeutil.FormatDisplay(a.DateTimeOriginal),
		"orientation":       a.Orientation,
		"white_balance":     a.WhiteBalance,
		"exposure_bias":     a.ExposureBias,
		"rating":            a.Rating,
		"tags":              a.Tags,
		"sharpness":         a.Sharpness,
		"noise":             a.Noise,
		"clip_shadow":       a.ClipShadow,
		"clip_highlight":    a.ClipHighlight,
		"ev_deviation":      a.EVDeviation,
		"tile_status":       a.TileStatus,
		"tile_max_z":        a.TileMaxZ,
		"width":             a.Width,
		"height":            a.Height,
		"fidelity":          "preview_jpeg",
		"fidelity_label":    "预览级 (Embedded JPEG)",
		"thumb_url":         "/api/v1/assets/" + strconv.FormatInt(a.ID, 10) + "/thumb.jpg",
		"preview_url":       "/api/v1/assets/" + strconv.FormatInt(a.ID, 10) + "/preview",
	}
	if full {
		var raw any
		_ = json.Unmarshal(a.ExifRaw, &raw)
		if raw == nil {
			raw = map[string]any{}
		}
		m["exif_raw"] = raw
	}
	return m
}

func (s *Server) preview(w http.ResponseWriter, r *http.Request) {
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
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, a.PreviewPath)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
