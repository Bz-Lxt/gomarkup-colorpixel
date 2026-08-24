package jobs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"colorpixel/internal/logger"
	"colorpixel/internal/metrics"
	"colorpixel/internal/store"
	"colorpixel/internal/tiles"
)

type Worker struct {
	DB      *store.DB
	DataDir string
	once    sync.Once
}

func (w *Worker) Start(ctx context.Context) {
	w.once.Do(func() {
		if err := w.DB.RecoverJobs(ctx); err != nil {
			logger.L().Error("job recover failed", "err", err)
		}
	})
	t := time.NewTicker(800 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	items, err := w.DB.ClaimJobs(ctx, 2)
	if err != nil {
		logger.L().Error("claim jobs", "err", err)
		return
	}
	for _, j := range items {
		err := w.run(ctx, j)
		msg := ""
		ok := err == nil
		if err != nil {
			msg = err.Error()
			logger.L().Warn("job failed", "id", j.ID, "asset", j.AssetID, "err", err)
		}
		_ = w.DB.FinishJob(ctx, j.ID, ok, msg)
	}
}

func (w *Worker) run(ctx context.Context, j store.Job) error {
	a, err := w.DB.GetAsset(ctx, j.AssetID)
	if err != nil {
		return err
	}
	preview, err := os.ReadFile(a.PreviewPath)
	if err != nil {
		return fmt.Errorf("preview: %w", err)
	}
	dest := filepath.Join(w.DataDir, "tiles", fmt.Sprintf("%d", a.ID))
	res, img, err := tiles.Build(preview, dest)
	if err != nil {
		_ = w.DB.UpdateTiles(ctx, a.ID, "failed", 0, 0, 0, nil, nil, nil, nil, nil)
		return err
	}
	sc := metrics.Analyze(img)
	sh, no, cs, ch, ev := sc.Sharpness, sc.Noise, sc.ClipShadow, sc.ClipHighlight, sc.EVDeviation
	return w.DB.UpdateTiles(ctx, a.ID, "ready", res.MaxZ, res.Width, res.Height, &sh, &no, &cs, &ch, &ev)
}
