package ingest_test

import (
	"bytes"
	"encoding/binary"
	"image/jpeg"
	"path/filepath"
	"testing"
	"time"

	"colorpixel/internal/ingest"
	"colorpixel/internal/raw"
	"colorpixel/internal/sample"
)

func TestIngestCR3LargeSizeUUIDPreservesMetadata(t *testing.T) {
	spec := sample.Spec{
		Filename: "metadata.dng", Format: "DNG", Make: "Canon", Model: "EOS R5",
		Lens: "RF 24-70mm F2.8 L IS USM", ApertureN: 28, ApertureD: 10,
		ShutterN: 1, ShutterD: 250, ISO: 640, FocalN: 50, FocalD: 1, Focal35: 50,
		ShotAt: time.Date(2026, time.March, 14, 10, 30, 0, 0, time.UTC),
		Width: 6000, Height: 4000, Seed: 17,
	}
	tiff, err := sample.Render(spec)
	if err != nil {
		t.Fatal(err)
	}
	data := cr3WithLargeSizeMetadata(tiff)

	outcome, err := ingest.Ingest(
		bytes.NewReader(data),
		filepath.Join(t.TempDir(), "capture.cr3"),
		"capture.cr3",
		len(data),
		raw.DefaultLimits(),
	)
	if err != nil {
		t.Fatal(err)
	}
	got := outcome.Result
	if got.Format != raw.FormatCR3 {
		t.Fatalf("format = %q, want %q", got.Format, raw.FormatCR3)
	}
	if _, err := jpeg.Decode(bytes.NewReader(got.Preview)); err != nil {
		t.Fatalf("preview should remain decodable: %v", err)
	}
	if got.Make != spec.Make || got.Model != spec.Model {
		t.Fatalf("camera = %q %q, want %q %q", got.Make, got.Model, spec.Make, spec.Model)
	}
	if got.LensModel != spec.Lens {
		t.Fatalf("lens = %q, want %q", got.LensModel, spec.Lens)
	}
	if got.ISO != int(spec.ISO) || got.FocalLength35mm != float64(spec.Focal35) {
		t.Fatalf("exposure metadata = ISO %d, %.0fmm; want ISO %d, %dmm", got.ISO, got.FocalLength35mm, spec.ISO, spec.Focal35)
	}
}

func cr3WithLargeSizeMetadata(tiff []byte) []byte {
	ftyp := bmffBox("ftyp", append([]byte("crx "), 0, 0, 0, 0))
	payload := append([]byte("CMT1"), tiff...)
	uuid := make([]byte, 32, 32+len(payload))
	binary.BigEndian.PutUint32(uuid[0:4], 1)
	copy(uuid[4:8], "uuid")
	binary.BigEndian.PutUint64(uuid[8:16], uint64(32+len(payload)))
	copy(uuid[16:32], []byte{
		0x85, 0xc0, 0xb6, 0x87, 0x82, 0x0f, 0x11, 0xe0,
		0x81, 0x11, 0xf4, 0xce, 0x46, 0x2b, 0x6a, 0x48,
	})
	uuid = append(uuid, payload...)
	return append(ftyp, bmffBox("moov", uuid)...)
}

func bmffBox(typ string, payload []byte) []byte {
	b := make([]byte, 8, 8+len(payload))
	binary.BigEndian.PutUint32(b[0:4], uint32(8+len(payload)))
	copy(b[4:8], typ)
	return append(b, payload...)
}
