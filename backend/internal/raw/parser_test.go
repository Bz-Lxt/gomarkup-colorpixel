package raw_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"colorpixel/internal/ingest"
	"colorpixel/internal/raw"
	"colorpixel/internal/sample"
	"colorpixel/internal/timeutil"
)

func TestParseAllFormats(t *testing.T) {
	specs := []sample.Spec{
		baseSpec("A.CR3", "CR3", "Canon", "EOS R5", "RF 50mm F1.2 L USM"),
		baseSpec("B.CR2", "CR2", "Canon", "EOS 5D Mark IV", "EF 24-105mm f/4L IS II USM"),
		baseSpec("C.NEF", "NEF", "Nikon", "Z 8", "NIKKOR Z 24-70mm f/2.8 S"),
		baseSpec("D.ARW", "ARW", "SONY", "ILCE-7RM5", "FE 24-70mm F2.8 GM II"),
		baseSpec("E.DNG", "DNG", "Adobe", "Lightroom DNG", "Sigma 35mm F1.4 DG DN"),
	}
	lim := raw.DefaultLimits()
	for _, s := range specs {
		s := s
		t.Run(s.Format, func(t *testing.T) {
			data, err := sample.Render(s)
			if err != nil {
				t.Fatal(err)
			}
			res, err := raw.Parse(raw.NewBytes(data), s.Filename, int64(len(data)), lim)
			if err != nil {
				t.Fatal(err)
			}
			if res.Make != s.Make {
				t.Fatalf("make %q want %q", res.Make, s.Make)
			}
			if res.Model != s.Model {
				t.Fatalf("model %q want %q", res.Model, s.Model)
			}
			if res.LensModel != s.Lens {
				t.Fatalf("lens %q want %q", res.LensModel, s.Lens)
			}
			if res.ISO != 400 {
				t.Fatalf("iso %d", res.ISO)
			}
			if res.Aperture < 1.3 || res.Aperture > 1.5 {
				t.Fatalf("aperture %v", res.Aperture)
			}
			if res.Preview == nil || res.Preview[0] != 0xFF || res.Preview[1] != 0xD8 {
				t.Fatalf("missing jpeg preview")
			}
			if res.ExtractionMode != raw.ModeStream {
				t.Fatalf("mode %s", res.ExtractionMode)
			}
			if res.DateTimeOriginal.IsZero() {
				t.Fatal("datetime missing")
			}
		})
	}
}

func TestDeferredPreview(t *testing.T) {
	s := baseSpec("DEFERRED.NEF", "NEF", "Nikon", "Z 8", "NIKKOR Z 24-70mm f/2.8 S")
	s.Deferred = true
	data, err := sample.Render(s)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 20<<20 {
		t.Fatalf("expected padded file, got %d", len(data))
	}
	dir := t.TempDir()
	oc, err := ingest.Ingest(bytes.NewReader(data), dir+"/x.nef", s.Filename, 16<<20, raw.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if oc.Result.ExtractionMode != raw.ModeDeferred {
		t.Fatalf("want deferred, got %s", oc.Result.ExtractionMode)
	}
	if oc.Result.Preview == nil {
		t.Fatal("preview missing on deferred path")
	}
}

func TestUnknownFormatDegrades(t *testing.T) {
	res, err := raw.Parse(raw.NewBytes([]byte("not-a-raw-file-at-all")), "note.txt", 16<<20, raw.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if res.Format != raw.FormatUNK || res.ExtractionMode != raw.ModeNone {
		t.Fatalf("%+v", res)
	}
}

func TestMalformedIFDOffset(t *testing.T) {
	buf := []byte{'I', 'I', 42, 0, 0, 0, 0, 80}
	_, err := raw.Parse(raw.NewBytes(buf), "bad.dng", 16<<20, raw.DefaultLimits())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAllEXIFTypes(t *testing.T) {
	w := make([]byte, 128)
	w[0], w[1] = 'I', 'I'
	binary.LittleEndian.PutUint16(w[2:], 42)
	binary.LittleEndian.PutUint32(w[4:], 8)
	binary.LittleEndian.PutUint16(w[8:], 3)
	putEntry(w[10:], 0x010F, 2, 4, 0x6E616143)
	putEntry(w[22:], 0x0112, 3, 1, 1)
	putEntry(w[34:], 0x0100, 4, 1, 640)
	binary.LittleEndian.PutUint32(w[46:], 0)
	res, err := raw.Parse(raw.NewBytes(w), "t.dng", 16<<20, raw.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if res.Orientation != 1 {
		t.Fatalf("orientation %d", res.Orientation)
	}
}

func putEntry(b []byte, tag, typ uint16, count, rawv uint32) {
	binary.LittleEndian.PutUint16(b[0:], tag)
	binary.LittleEndian.PutUint16(b[2:], typ)
	binary.LittleEndian.PutUint32(b[4:], count)
	binary.LittleEndian.PutUint32(b[8:], rawv)
}

func baseSpec(name, format, make, model, lens string) sample.Spec {
	return sample.Spec{
		Filename: name, Format: format, Make: make, Model: model, Lens: lens,
		ApertureN: 14, ApertureD: 10, ShutterN: 1, ShutterD: 200, ISO: 400,
		FocalN: 50, FocalD: 1, Focal35: 50, ShotAt: timeutil.Now(),
		Width: 640, Height: 420, Seed: 9,
	}
}

func FuzzParse(f *testing.F) {
	s := baseSpec("x.dng", "DNG", "Adobe", "DNG", "Lens")
	data, _ := sample.Render(s)
	f.Add(data)
	f.Add([]byte("II*\x00"))
	f.Fuzz(func(t *testing.T, b []byte) {
		_, _ = raw.Parse(raw.NewBytes(b), "fuzz.dng", int64(len(b)), raw.DefaultLimits())
	})
}
