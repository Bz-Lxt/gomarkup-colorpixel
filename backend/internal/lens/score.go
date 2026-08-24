package lens

import (
	"fmt"
	"math"
	"sort"
	"time"

	"colorpixel/internal/store"
	"colorpixel/internal/timeutil"
)

type Factor struct {
	Value      float64 `json:"value"`
	Confidence float64 `json:"confidence"`
	Samples    int     `json:"samples"`
	Excluded   int     `json:"excluded_count"`
}

type LensScore struct {
	Lens          string             `json:"lens"`
	Count         int                `json:"count"`
	Insufficient  bool               `json:"insufficient_data"`
	Score         float64            `json:"score"`
	Factors       map[string]Factor  `json:"factors"`
	Weights       Weights            `json:"weights"`
	Derivation    []string           `json:"derivation"`
	FocalHist     map[string]int     `json:"focal_hist"`
	ApertureHist  map[string]int     `json:"aperture_hist"`
	ISOHist       map[string]int     `json:"iso_hist"`
	MonthHist     map[string]int     `json:"month_hist"`
	HourHist      map[string]int     `json:"hour_hist"`
}

type Report struct {
	WindowFrom   string      `json:"window_from"`
	WindowTo     string      `json:"window_to"`
	Total        int         `json:"total"`
	GoldenLens   string      `json:"golden_lens"`
	Recommended  string      `json:"recommended_combo"`
	Lenses       []LensScore `json:"lenses"`
	FocalGlobal  map[string]int `json:"focal_global"`
	ApertureGlobal map[string]int `json:"aperture_global"`
	Weights      Weights     `json:"weights"`
}

func Build(assets []store.Asset, now time.Time, w Weights) Report {
	w = w.Normalize()
	to := timeutil.ToBeijing(now)
	from := to.AddDate(-1, 0, 0)
	var in []store.Asset
	for _, a := range assets {
		if a.DateTimeOriginal.IsZero() {
			continue
		}
		t := timeutil.ToBeijing(a.DateTimeOriginal)
		if !t.Before(from) && t.Before(to) {
			in = append(in, a)
		}
	}
	by := map[string][]store.Asset{}
	focalG := map[string]int{}
	apG := map[string]int{}
	for _, a := range in {
		key := a.LensModel
		if key == "" {
			key = "(unknown lens)"
		}
		by[key] = append(by[key], a)
		focalG[focalBucket(a.FocalLength35mm)]++
		apG[apertureBucket(a.Aperture)]++
	}
	var scores []LensScore
	for lens, rows := range by {
		scores = append(scores, scoreLens(lens, rows, in, w))
	}
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Insufficient != scores[j].Insufficient {
			return !scores[i].Insufficient && scores[j].Insufficient
		}
		return scores[i].Score > scores[j].Score
	})
	golden := ""
	combo := ""
	for _, s := range scores {
		if !s.Insufficient {
			golden = s.Lens
			bestF, bestA := topKey(s.FocalHist), topKey(s.ApertureHist)
			combo = s.Lens + " · " + bestF + " · " + bestA
			break
		}
	}
	if scores == nil {
		scores = []LensScore{}
	}
	return Report{
		WindowFrom:     timeutil.FormatDisplay(from),
		WindowTo:       timeutil.FormatDisplay(to),
		Total:          len(in),
		GoldenLens:     golden,
		Recommended:    combo,
		Lenses:         scores,
		FocalGlobal:    focalG,
		ApertureGlobal: apG,
		Weights:        w,
	}
}

func scoreLens(name string, rows, all []store.Asset, w Weights) LensScore {
	ls := LensScore{
		Lens: name, Count: len(rows), Insufficient: len(rows) < 30,
		Factors: map[string]Factor{}, Weights: w,
		FocalHist: map[string]int{}, ApertureHist: map[string]int{},
		ISOHist: map[string]int{}, MonthHist: map[string]int{}, HourHist: map[string]int{},
	}
	U := float64(len(rows)) / math.Max(1, float64(len(all)))
	focalBuckets := map[string]int{}
	for _, a := range rows {
		fb := focalBucket(a.FocalLength35mm)
		focalBuckets[fb]++
		ls.FocalHist[fb]++
		ls.ApertureHist[apertureBucket(a.Aperture)]++
		ls.ISOHist[isoBucket(a.ISO)]++
		if !a.DateTimeOriginal.IsZero() {
			t := timeutil.ToBeijing(a.DateTimeOriginal)
			ls.MonthHist[t.Format("2006-01")]++
			ls.HourHist[t.Format("15")]++
		}
	}
	V := shannon(focalBuckets)
	A := apertureUtil(rows)
	S, sN, sX := medianProxy(rows, func(a store.Asset) (float64, bool) {
		if a.Sharpness == nil {
			return 0, false
		}
		return *a.Sharpness, true
	})
	Nraw, nN, nX := medianProxy(rows, func(a store.Asset) (float64, bool) {
		if a.Noise == nil {
			return 0, false
		}
		return *a.Noise, true
	})
	Eraw, eN, eX := medianProxy(rows, func(a store.Asset) (float64, bool) {
		if a.ClipShadow == nil || a.ClipHighlight == nil {
			return 0, false
		}
		return 1 - math.Min(1, *a.ClipShadow+*a.ClipHighlight), true
	})
	R, rN, rX := medianProxy(rows, func(a store.Asset) (float64, bool) {
		if a.Rating > 0 {
			return float64(a.Rating) / 5.0, true
		}
		return 0, false
	})
	if rN == 0 {
		O := 0.4*normSharp(S) + 0.3*normNoise(Nraw) + 0.3*clamp01(Eraw)
		R = O
		rN = sN
	}
	Sn := normSharp(S)
	Nn := normNoise(Nraw)
	En := clamp01(Eraw)
	ls.Factors["U"] = Factor{U, conf(len(rows)), len(rows), 0}
	ls.Factors["V"] = Factor{V, conf(len(rows)), len(rows), 0}
	ls.Factors["A"] = Factor{A, conf(len(rows)), len(rows), 0}
	ls.Factors["S"] = Factor{Sn, conf(sN), sN, sX}
	ls.Factors["N"] = Factor{Nn, conf(nN), nN, nX}
	ls.Factors["E"] = Factor{En, conf(eN), eN, eX}
	ls.Factors["R"] = Factor{R, conf(rN), rN, rX}
	ls.Score = 100 * (w.U*U + w.V*V + w.A*A + w.S*Sn + w.N*Nn + w.E*En + w.R*R)
	ls.Derivation = []string{
		sprintf("U=%.3f (share of %d/%d)", U, len(rows), len(all)),
		sprintf("V=%.3f (focal entropy)", V),
		sprintf("A=%.3f (aperture utilization)", A),
		sprintf("S=%.3f from %d samples (%d excluded)", Sn, sN, sX),
		sprintf("N=%.3f from %d samples (%d excluded)", Nn, nN, nX),
		sprintf("E=%.3f from %d samples (%d excluded)", En, eN, eX),
		sprintf("R=%.3f", R),
		sprintf("Score=100*(%.2fU+%.2fV+%.2fA+%.2fS+%.2fN+%.2fE+%.2fR)=%.2f", w.U, w.V, w.A, w.S, w.N, w.E, w.R, ls.Score),
	}
	return ls
}

func medianProxy(rows []store.Asset, fn func(store.Asset) (float64, bool)) (float64, int, int) {
	var vs []float64
	ex := 0
	for _, a := range rows {
		v, ok := fn(a)
		if !ok {
			ex++
			continue
		}
		vs = append(vs, v)
	}
	if len(vs) == 0 {
		return 0, 0, ex
	}
	sort.Float64s(vs)
	return vs[len(vs)/2], len(vs), ex
}

func shannon(h map[string]int) float64 {
	total := 0
	for _, n := range h {
		total += n
	}
	if total == 0 || len(h) <= 1 {
		return 0
	}
	ent := 0.0
	for _, n := range h {
		if n == 0 {
			continue
		}
		p := float64(n) / float64(total)
		ent -= p * math.Log(p)
	}
	return ent / math.Log(float64(len(h)))
}

func apertureUtil(rows []store.Asset) float64 {
	if len(rows) == 0 {
		return 0
	}
	minA := 99.0
	for _, a := range rows {
		if a.Aperture > 0 && a.Aperture < minA {
			minA = a.Aperture
		}
	}
	if minA >= 99 {
		return 0
	}
	var s float64
	n := 0
	for _, a := range rows {
		if a.Aperture <= 0 {
			continue
		}
		s += clamp01(minA / a.Aperture)
		n++
	}
	if n == 0 {
		return 0
	}
	return s / float64(n)
}

func focalBucket(f float64) string {
	switch {
	case f <= 0:
		return "unknown"
	case f < 24:
		return "14-23"
	case f < 35:
		return "24-34"
	case f < 50:
		return "35-49"
	case f < 85:
		return "50-84"
	case f < 135:
		return "85-134"
	default:
		return "135+"
	}
}

func apertureBucket(a float64) string {
	if a <= 0 {
		return "unknown"
	}
	stops := []float64{1.0, 1.2, 1.4, 1.8, 2.0, 2.2, 2.5, 2.8, 3.2, 3.5, 4.0, 4.5, 5.0, 5.6, 6.3, 7.1, 8.0, 11, 16}
	best := stops[0]
	for _, s := range stops {
		if math.Abs(s-a) < math.Abs(best-a) {
			best = s
		}
	}
	return sprintf("f/%.1f", best)
}

func isoBucket(iso int) string {
	switch {
	case iso <= 0:
		return "unknown"
	case iso <= 200:
		return "100-200"
	case iso <= 400:
		return "400"
	case iso <= 800:
		return "800"
	case iso <= 1600:
		return "1600"
	default:
		return "3200+"
	}
}

func conf(n int) float64 {
	return math.Min(1, float64(n)/100)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func normSharp(v float64) float64 {
	return clamp01(v / 40.0)
}

func normNoise(v float64) float64 {
	return clamp01(1 - v/25.0)
}

func topKey(m map[string]int) string {
	bestK := ""
	best := -1
	for k, n := range m {
		if n > best {
			best = n
			bestK = k
		}
	}
	return bestK
}

func sprintf(f string, a ...any) string {
	return fmt.Sprintf(f, a...)
}
