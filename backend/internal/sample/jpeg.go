package sample

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"math"
)

func EncodeScene(w, h int, seed int, hue float64) ([]byte, error) {
	if w < 8 {
		w = 8
	}
	if h < 8 {
		h = 8
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			fx := float64(x) / float64(w)
			fy := float64(y) / float64(h)
			n := noise(x+seed*13, y+seed*7)
			edge := 0.0
			if x%64 < 2 || y%64 < 2 {
				edge = 40
			}
			r := clamp8(80 + 120*fx + 20*math.Sin(hue) + n + edge)
			g := clamp8(70 + 110*fy + 18*math.Cos(hue*1.3) + n*0.6)
			b := clamp8(90 + 90*(1-fx) + 16*math.Sin(fy*6+hue) + n*0.4)
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 86}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func noise(x, y int) float64 {
	v := (x * 374761393) ^ (y * 668265263)
	v = (v ^ (v >> 13)) * 1274126177
	return float64(int(v%21) - 10)
}

func clamp8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
