package ingest

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"colorpixel/internal/raw"
)

type Outcome struct {
	Result   *raw.Result
	DestPath string
	Size     int64
}

func Ingest(r io.Reader, destPath string, filename string, windowBytes int, lim raw.Limits) (*Outcome, error) {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return nil, err
	}
	w, err := Capture(r, destPath, windowBytes)
	if err != nil {
		return nil, fmt.Errorf("capture: %w", err)
	}
	defer w.Close()

	res, err := raw.Parse(w, filename, w.WindowSize(), lim)
	if err != nil {
		return nil, err
	}
	if res.Preview != nil && len(res.Preview) > lim.PreviewMax {
		tmp := filepath.Join(filepath.Dir(destPath), filepath.Base(destPath)+".preview.bin")
		if err := os.WriteFile(tmp, res.Preview, 0o644); err == nil {
			res.Preview = nil
			res.Warnings = append(res.Warnings, "preview spilled to disk")
		} else {
			res.Preview = res.Preview[:lim.PreviewMax]
		}
	}
	return &Outcome{Result: res, DestPath: destPath, Size: w.Size()}, nil
}
