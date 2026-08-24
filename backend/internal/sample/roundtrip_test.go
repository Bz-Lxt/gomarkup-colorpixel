package sample_test

import (
	"testing"

	"colorpixel/internal/raw"
	"colorpixel/internal/sample"
	"colorpixel/internal/timeutil"
)

func TestCatalogCoversYearAndFormats(t *testing.T) {
	cat := sample.BuildCatalog()
	if len(cat) < 200 {
		t.Fatalf("catalog %d", len(cat))
	}
	seen := map[string]int{}
	months := map[string]int{}
	deferred := 0
	for _, s := range cat {
		seen[s.Format]++
		months[s.ShotAt.In(timeutil.Beijing).Format("2006-01")]++
		if s.Deferred {
			deferred++
		}
	}
	for _, f := range []string{"CR3", "CR2", "NEF", "ARW", "DNG"} {
		if seen[f] == 0 {
			t.Fatalf("missing format %s", f)
		}
	}
	if len(months) < 12 {
		t.Fatalf("months %d", len(months))
	}
	if deferred != 1 {
		t.Fatalf("deferred %d", deferred)
	}
}

func TestRenderRoundtripLensFields(t *testing.T) {
	s := sample.Spec{
		Filename: "x.arw", Format: "ARW", Make: "SONY", Model: "ILCE-7RM5",
		Lens: "FE 35mm F1.4 GM", ApertureN: 14, ApertureD: 10, ShutterN: 1, ShutterD: 250,
		ISO: 640, FocalN: 35, FocalD: 1, Focal35: 35, ShotAt: timeutil.Now(),
		Width: 200, Height: 120, Seed: 4,
	}
	data, err := sample.Render(s)
	if err != nil {
		t.Fatal(err)
	}
	res, err := raw.Parse(raw.NewBytes(data), s.Filename, int64(len(data)), raw.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if res.ISO != 640 || res.FocalLength35mm != 35 {
		t.Fatalf("%+v", res)
	}
	if raw.TagName(0xA434) != "LensModel" {
		t.Fatal("tag registry")
	}
}

func TestEncodeSceneNotEmpty(t *testing.T) {
	b, err := sample.EncodeScene(32, 24, 1, 0.5)
	if err != nil || len(b) < 100 || b[0] != 0xFF {
		t.Fatalf("jpeg %d %v", len(b), err)
	}
}
