package metrics_test

import (
	"image"
	"image/color"
	"testing"

	"colorpixel/internal/metrics"
)

func TestAnalyzeSharpVsBlur(t *testing.T) {
	sharp := image.NewRGBA(image.Rect(0, 0, 64, 64))
	blur := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			v := uint8(120)
			if x%8 < 4 {
				v = 220
			}
			sharp.Set(x, y, color.RGBA{v, v, v, 255})
			blur.Set(x, y, color.RGBA{130, 130, 130, 255})
		}
	}
	s1 := metrics.Analyze(sharp)
	s2 := metrics.Analyze(blur)
	if s1.Sharpness <= s2.Sharpness {
		t.Fatalf("sharp %v blur %v", s1.Sharpness, s2.Sharpness)
	}
}
