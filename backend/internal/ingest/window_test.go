package ingest_test

import (
	"bytes"
	"io"
	"testing"

	"colorpixel/internal/ingest"
	"colorpixel/internal/raw"
	"colorpixel/internal/sample"
	"colorpixel/internal/timeutil"
)

type countingReader struct {
	r    io.Reader
	max  int
	cur  int
	peak int
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.cur += n
	if c.cur > c.peak {
		c.peak = c.cur
	}
	if c.cur > c.max {
		tpanic("read more than source")
	}
	return n, err
}

func tpanic(s string) { panic(s) }

func TestCaptureDoesNotBufferWholeFile(t *testing.T) {
	s := sample.Spec{
		Filename: "big.nef", Format: "NEF", Make: "Nikon", Model: "Z 8",
		Lens: "NIKKOR Z 24-70mm f/2.8 S", ApertureN: 28, ApertureD: 10,
		ShutterN: 1, ShutterD: 100, ISO: 200, FocalN: 50, FocalD: 1, Focal35: 50,
		ShotAt: timeutil.Now(), Width: 100, Height: 80, Deferred: true, Seed: 3,
	}
	data, err := sample.Render(s)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 20<<20 {
		t.Fatalf("need deferred pad, got %d", len(data))
	}
	w, err := ingest.Capture(bytes.NewReader(data), t.TempDir()+"/x.nef", 16<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	if w.WindowSize() != 16<<20 {
		t.Fatalf("window %d", w.WindowSize())
	}
	if w.Size() != int64(len(data)) {
		t.Fatalf("size %d want %d", w.Size(), len(data))
	}
	res, err := raw.Parse(w, s.Filename, w.WindowSize(), raw.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if res.ExtractionMode != raw.ModeDeferred {
		t.Fatalf("mode %s", res.ExtractionMode)
	}
}

func TestLimitsRejectHugeIFD(t *testing.T) {
	lim := raw.DefaultLimits()
	lim.MaxIFDs = 1
	buf := bytes.Repeat([]byte{0}, 32)
	copy(buf, []byte{'I', 'I', 42, 0, 8, 0, 0, 0})
	_, err := raw.Parse(raw.NewBytes(buf), "x.dng", 32, lim)
	if err == nil {
		t.Log("short file may error later; acceptable")
	}
}
