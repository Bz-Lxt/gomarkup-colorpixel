package tiles

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"math"
	"os"
	"path/filepath"

	"golang.org/x/image/draw"
)

const TileSize = 256

type Result struct {
	MaxZ   int
	Width  int
	Height int
}

func Build(preview []byte, destDir string) (*Result, image.Image, error) {
	img, err := jpeg.Decode(bytes.NewReader(preview))
	if err != nil {
		return nil, nil, fmt.Errorf("decode preview: %w", err)
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return nil, nil, fmt.Errorf("empty image")
	}
	maxZ := int(math.Ceil(math.Log2(math.Max(float64(w), float64(h)) / float64(TileSize))))
	if maxZ < 0 {
		maxZ = 0
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, nil, err
	}
	for z := 0; z <= maxZ; z++ {
		scale := math.Pow(2, float64(maxZ-z))
		lw := int(math.Max(1, math.Round(float64(w)/scale)))
		lh := int(math.Max(1, math.Round(float64(h)/scale)))
		level := image.NewRGBA(image.Rect(0, 0, lw, lh))
		draw.CatmullRom.Scale(level, level.Bounds(), img, b, draw.Over, nil)
		cols := int(math.Ceil(float64(lw) / TileSize))
		rows := int(math.Ceil(float64(lh) / TileSize))
		for y := 0; y < rows; y++ {
			for x := 0; x < cols; x++ {
				tile := image.NewRGBA(image.Rect(0, 0, TileSize, TileSize))
				src := image.Rect(x*TileSize, y*TileSize, min((x+1)*TileSize, lw), min((y+1)*TileSize, lh))
				draw.Draw(tile, image.Rect(0, 0, src.Dx(), src.Dy()), level, src.Min, draw.Src)
				var buf bytes.Buffer
				if err := jpeg.Encode(&buf, tile, &jpeg.Options{Quality: 82}); err != nil {
					return nil, nil, err
				}
				name := filepath.Join(destDir, fmt.Sprintf("%d", z), fmt.Sprintf("%d_%d.jpg", x, y))
				if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
					return nil, nil, err
				}
				if err := os.WriteFile(name, buf.Bytes(), 0o644); err != nil {
					return nil, nil, err
				}
			}
		}
	}
	thumb := image.NewRGBA(image.Rect(0, 0, 320, 210))
	draw.CatmullRom.Scale(thumb, thumb.Bounds(), img, b, draw.Over, nil)
	var tbuf bytes.Buffer
	_ = jpeg.Encode(&tbuf, thumb, &jpeg.Options{Quality: 80})
	_ = os.WriteFile(filepath.Join(destDir, "thumb.jpg"), tbuf.Bytes(), 0o644)
	return &Result{MaxZ: maxZ, Width: w, Height: h}, img, nil
}

func Path(destDir string, z, x, y int) string {
	return filepath.Join(destDir, fmt.Sprintf("%d", z), fmt.Sprintf("%d_%d.jpg", x, y))
}
