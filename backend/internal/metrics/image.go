package metrics

import (
	"image"
	"image/color"
	"math"
)

type Scores struct {
	Sharpness    float64
	Noise        float64
	ClipShadow   float64
	ClipHighlight float64
	EVDeviation  float64
	HistR        []int
	HistG        []int
	HistB        []int
	HistY        []int
}

func Analyze(img image.Image) Scores {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	hr := make([]int, 256)
	hg := make([]int, 256)
	hb := make([]int, 256)
	hy := make([]int, 256)
	var lapSum, lapN float64
	var darkN float64
	var darkMean float64
	var clipLo, clipHi, total float64
	var lumSum float64
	step := 1
	if w*h > 400_000 {
		step = 2
	}
	gray := func(x, y int) float64 {
		r, g, b, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
		return (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 256.0
	}
	for y := 1; y < h-1; y += step {
		for x := 1; x < w-1; x += step {
			c := img.At(b.Min.X+x, b.Min.Y+y)
			rr, gg, bb, _ := c.RGBA()
			r8 := int(rr >> 8)
			g8 := int(gg >> 8)
			b8 := int(bb >> 8)
			if r8 > 255 {
				r8 = 255
			}
			if g8 > 255 {
				g8 = 255
			}
			if b8 > 255 {
				b8 = 255
			}
			hr[r8]++
			hg[g8]++
			hb[b8]++
			y8 := int((77*r8 + 150*g8 + 29*b8) >> 8)
			if y8 > 255 {
				y8 = 255
			}
			hy[y8]++
			total++
			lum := float64(y8)
			lumSum += lum
			if y8 <= 5 {
				clipLo++
			}
			if y8 >= 250 {
				clipHi++
			}
			c0 := gray(x, y)
			lap := math.Abs(4*c0 - gray(x-1, y) - gray(x+1, y) - gray(x, y-1) - gray(x, y+1))
			lapSum += lap
			lapN++
			if y8 < 40 {
				darkMean += lum
				darkN++
			}
		}
	}
	s := Scores{HistR: hr, HistG: hg, HistB: hb, HistY: hy}
	if lapN > 0 {
		s.Sharpness = lapSum / lapN
	}
	if total > 0 {
		s.ClipShadow = clipLo / total
		s.ClipHighlight = clipHi / total
		mean := lumSum / total
		s.EVDeviation = math.Abs(math.Log2((mean+1)/128.0))
	}
	if darkN > 8 {
		mean := darkMean / darkN
		var ss float64
		n := 0.0
		for y := 1; y < h-1; y += step * 2 {
			for x := 1; x < w-1; x += step * 2 {
				r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
				y8 := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(bl)) / 256.0
				if y8 < 40 {
					d := y8 - mean
					ss += d * d
					n++
				}
			}
		}
		if n > 0 {
			s.Noise = math.Sqrt(ss / n)
		}
	}
	return s
}

func HistogramFromImage(img image.Image) Scores {
	return Analyze(img)
}

func ToNRGBA(c color.Color) color.NRGBA {
	r, g, b, a := c.RGBA()
	return color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}
