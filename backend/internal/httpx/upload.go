package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"colorpixel/internal/ingest"
	"colorpixel/internal/raw"
	"colorpixel/internal/store"
	"colorpixel/internal/timeutil"
)

func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, s.Cfg.MaxUploadBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeErr(w, http.StatusBadRequest, "upload", "multipart required")
		return
	}
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		files = r.MultipartForm.File["file"]
	}
	if len(files) == 0 {
		writeErr(w, http.StatusBadRequest, "upload", "no files")
		return
	}
	lim := raw.Limits{
		MaxIFDs: s.Cfg.MaxIFDs, MaxDepth: s.Cfg.MaxIFDDepth,
		MaxAlloc: s.Cfg.MaxAllocCount, PreviewMax: s.Cfg.PreviewMaxBytes,
		WindowBytes: s.Cfg.PreviewWindowBytes,
	}
	var out []map[string]any
	for _, fh := range files {
		name := sanitizeName(fh.Filename)
		f, err := fh.Open()
		if err != nil {
			writeErr(w, http.StatusBadRequest, "upload", err.Error())
			return
		}
		dest := filepath.Join(s.Cfg.DataDir, "raw", fmt.Sprintf("%d_%s", timeutil.Now().UnixNano(), name))
		oc, err := ingest.Ingest(f, dest, name, s.Cfg.PreviewWindowBytes, lim)
		_ = f.Close()
		if err != nil {
			writeErr(w, http.StatusUnprocessableEntity, "parse", err.Error())
			return
		}
		previewPath := ""
		if oc.Result.Preview != nil {
			previewPath = dest + ".preview.jpg"
			if err := os.WriteFile(previewPath, oc.Result.Preview, 0o644); err != nil {
				writeErr(w, http.StatusInternalServerError, "preview", err.Error())
				return
			}
		}
		tagsJSON, _ := json.Marshal(oc.Result.Tags)
		a := &store.Asset{
			Filename:         name,
			Format:           string(oc.Result.Format),
			SizeBytes:        oc.Size,
			StoragePath:      dest,
			PreviewPath:      previewPath,
			ExtractionMode:   oc.Result.ExtractionMode,
			CameraMake:       oc.Result.Make,
			CameraModel:      oc.Result.Model,
			LensModel:        oc.Result.LensModel,
			LensSpec:         oc.Result.LensSpec,
			Aperture:         oc.Result.Aperture,
			ShutterText:      oc.Result.ShutterText,
			ShutterSeconds:   oc.Result.ShutterSeconds,
			ISO:              oc.Result.ISO,
			FocalLength:      oc.Result.FocalLength,
			FocalLength35mm:  oc.Result.FocalLength35mm,
			DateTimeOriginal: oc.Result.DateTimeOriginal,
			Orientation:      oc.Result.Orientation,
			WhiteBalance:     oc.Result.WhiteBalance,
			ExposureBias:     oc.Result.ExposureBias,
			Width:            oc.Result.Width,
			Height:           oc.Result.Height,
			TileStatus:       "pending",
			ExifRaw:          tagsJSON,
		}
		if previewPath == "" {
			a.TileStatus = "none"
		}
		if err := s.DB.InsertAsset(r.Context(), a); err != nil {
			writeErr(w, http.StatusInternalServerError, "db", err.Error())
			return
		}
		if previewPath != "" {
			if _, err := s.DB.EnqueueJob(r.Context(), a.ID, "tiles"); err != nil {
				writeErr(w, http.StatusInternalServerError, "job", err.Error())
				return
			}
		}
		out = append(out, map[string]any{
			"id":               a.ID,
			"filename":         a.Filename,
			"format":           a.Format,
			"extraction_mode":  a.ExtractionMode,
			"camera_make":      a.CameraMake,
			"camera_model":     a.CameraModel,
			"lens_model":       a.LensModel,
			"iso":              a.ISO,
			"aperture":         a.Aperture,
			"shutter_text":     a.ShutterText,
			"focal_length":     a.FocalLength,
			"datetime_original": timeutil.FormatDisplay(a.DateTimeOriginal),
			"warnings":         oc.Result.Warnings,
		})
	}
	writeOK(w, map[string]any{"items": out})
}

func sanitizeName(name string) string {
	name = filepath.Base(name)
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	s := b.String()
	if s == "" || s == "." || strings.Contains(s, "..") {
		return "upload.bin"
	}
	return s
}
