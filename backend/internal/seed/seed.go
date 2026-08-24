package seed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"colorpixel/internal/config"
	"colorpixel/internal/ingest"
	"colorpixel/internal/logger"
	"colorpixel/internal/raw"
	"colorpixel/internal/sample"
	"colorpixel/internal/store"
)

func Bootstrap(ctx context.Context, cfg config.Config, db *store.DB) error {
	if err := db.EnsureUser(ctx, "photographer", "colorpixel"); err != nil {
		return err
	}
	if !cfg.SampleMode {
		return nil
	}
	n, err := db.CountAlive(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	logger.L().Info("seeding sample RAW catalog")
	lim := raw.Limits{
		MaxIFDs: cfg.MaxIFDs, MaxDepth: cfg.MaxIFDDepth, MaxAlloc: cfg.MaxAllocCount,
		PreviewMax: cfg.PreviewMaxBytes, WindowBytes: cfg.PreviewWindowBytes,
	}
	specs := sample.BuildCatalog()
	for i, spec := range specs {
		data, err := sample.Render(spec)
		if err != nil {
			return fmt.Errorf("render %s: %w", spec.Filename, err)
		}
		dest := filepath.Join(cfg.DataDir, "raw", spec.Filename)
		oc, err := ingest.Ingest(bytes.NewReader(data), dest, spec.Filename, cfg.PreviewWindowBytes, lim)
		if err != nil {
			return fmt.Errorf("ingest %s: %w", spec.Filename, err)
		}
		previewPath := ""
		if oc.Result.Preview != nil {
			previewPath = dest + ".preview.jpg"
			if err := os.WriteFile(previewPath, oc.Result.Preview, 0o644); err != nil {
				return err
			}
		}
		tags, _ := json.Marshal(oc.Result.Tags)
		a := &store.Asset{
			Filename: spec.Filename, Format: string(oc.Result.Format), SizeBytes: oc.Size,
			StoragePath: dest, PreviewPath: previewPath, ExtractionMode: oc.Result.ExtractionMode,
			CameraMake: oc.Result.Make, CameraModel: oc.Result.Model, LensModel: oc.Result.LensModel,
			LensSpec: oc.Result.LensSpec, Aperture: oc.Result.Aperture, ShutterText: oc.Result.ShutterText,
			ShutterSeconds: oc.Result.ShutterSeconds, ISO: oc.Result.ISO, FocalLength: oc.Result.FocalLength,
			FocalLength35mm: oc.Result.FocalLength35mm, DateTimeOriginal: oc.Result.DateTimeOriginal,
			Orientation: oc.Result.Orientation, WhiteBalance: oc.Result.WhiteBalance, ExposureBias: oc.Result.ExposureBias,
			Width: oc.Result.Width, Height: oc.Result.Height, TileStatus: "pending", ExifRaw: tags,
			Rating: (i % 5) + 1,
		}
		if previewPath == "" {
			a.TileStatus = "none"
		}
		if err := db.InsertAsset(ctx, a); err != nil {
			return err
		}
		if previewPath != "" {
			if _, err := db.EnqueueJob(ctx, a.ID, "tiles"); err != nil {
				return err
			}
		}
	}
	logger.L().Info("seed complete", "count", len(specs))
	return nil
}
