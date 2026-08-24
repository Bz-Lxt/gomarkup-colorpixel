package lens_test

import (
	"testing"
	"time"

	"colorpixel/internal/lens"
	"colorpixel/internal/store"
	"colorpixel/internal/timeutil"
)

func ptr(v float64) *float64 { return &v }

func TestGoldenLensUsesRealMetrics(t *testing.T) {
	now := timeutil.Now()
	var assets []store.Asset
	for i := 0; i < 40; i++ {
		sh := 20.0 + float64(i%5)
		n := 4.0
		cs, ch := 0.01, 0.02
		assets = append(assets, store.Asset{
			LensModel: "RF 50mm F1.2 L USM",
			DateTimeOriginal: now.AddDate(0, 0, -i*3),
			FocalLength35mm: 50, Aperture: 1.4, ISO: 100, Rating: 5,
			Sharpness: &sh, Noise: &n, ClipShadow: &cs, ClipHighlight: &ch,
		})
	}
	for i := 0; i < 12; i++ {
		sh := 8.0
		n := 18.0
		cs, ch := 0.2, 0.2
		assets = append(assets, store.Asset{
			LensModel: "kit zoom",
			DateTimeOriginal: now.AddDate(0, 0, -i*8),
			FocalLength35mm: 24, Aperture: 5.6, ISO: 1600, Rating: 2,
			Sharpness: &sh, Noise: &n, ClipShadow: &cs, ClipHighlight: &ch,
		})
	}
	rep := lens.Build(assets, now, lens.DefaultWeights())
	if rep.GoldenLens != "RF 50mm F1.2 L USM" {
		t.Fatalf("golden %q", rep.GoldenLens)
	}
	if len(rep.Lenses) != 2 {
		t.Fatalf("lenses %d", len(rep.Lenses))
	}
	if rep.Lenses[0].Factors["S"].Excluded != 0 {
		t.Fatal("sharpness should not be excluded")
	}
	if rep.Total < 50 {
		t.Fatalf("total %d", rep.Total)
	}
}

func TestInsufficientData(t *testing.T) {
	now := time.Now().In(timeutil.Beijing)
	var assets []store.Asset
	for i := 0; i < 5; i++ {
		assets = append(assets, store.Asset{
			LensModel: "rare", DateTimeOriginal: now.AddDate(0, 0, -i),
			FocalLength35mm: 35, Aperture: 2,
		})
	}
	rep := lens.Build(assets, now, lens.DefaultWeights())
	if !rep.Lenses[0].Insufficient {
		t.Fatal("expected insufficient")
	}
}

func TestExcludeMissingMetrics(t *testing.T) {
	now := timeutil.Now()
	var assets []store.Asset
	for i := 0; i < 35; i++ {
		a := store.Asset{
			LensModel: "X", DateTimeOriginal: now.AddDate(0, 0, -i),
			FocalLength35mm: 35, Aperture: 2,
		}
		if i%2 == 0 {
			a.Sharpness = ptr(12)
		}
		assets = append(assets, a)
	}
	rep := lens.Build(assets, now, lens.DefaultWeights())
	if rep.Lenses[0].Factors["S"].Excluded == 0 {
		t.Fatal("expected excluded samples without sharpness")
	}
}
