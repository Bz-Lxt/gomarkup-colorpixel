package ingest_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"colorpixel/internal/ingest"
	"colorpixel/internal/raw"
	"colorpixel/internal/sample"
	"colorpixel/internal/timeutil"
)

const concurrentUploadSize = 32 * 1024

type stagedUpload struct {
	data    []byte
	copied  chan<- struct{}
	release <-chan struct{}
	used    bool
}

func (r *stagedUpload) Read(p []byte) (int, error) {
	if r.used {
		return 0, io.EOF
	}
	r.used = true
	n := copy(p, r.data)
	close(r.copied)
	if r.release != nil {
		<-r.release
	}
	return n, io.EOF
}

type ingestCall struct {
	out *ingest.Outcome
	err error
}

func TestConcurrentIngestKeepsUploadStreamsIsolated(t *testing.T) {
	oldProcs := runtime.GOMAXPROCS(1)
	t.Cleanup(func() { runtime.GOMAXPROCS(oldProcs) })
	oldGC := debug.SetGCPercent(-1)
	t.Cleanup(func() { debug.SetGCPercent(oldGC) })

	specA := concurrentSpec("A.DNG", "Alpha Camera", 125, 11)
	specB := concurrentSpec("B.DNG", "Beta Camera", 1600, 29)
	dataA := renderPaddedUpload(t, specA)
	dataB := renderPaddedUpload(t, specB)

	for attempt := 0; attempt < 16; attempt++ {
		dir := t.TempDir()
		pathA := filepath.Join(dir, fmt.Sprintf("a-%d.dng", attempt))
		pathB := filepath.Join(dir, fmt.Sprintf("b-%d.dng", attempt))
		aCopied := make(chan struct{})
		bCopied := make(chan struct{})
		aDone := make(chan ingestCall, 1)
		bDone := make(chan ingestCall, 1)

		go func() {
			out, err := ingest.Ingest(
				&stagedUpload{data: dataA, copied: aCopied, release: bCopied},
				pathA,
				specA.Filename,
				concurrentUploadSize,
				raw.DefaultLimits(),
			)
			aDone <- ingestCall{out: out, err: err}
		}()

		waitForSignal(t, aCopied, "first upload to fill its read buffer")

		go func() {
			out, err := ingest.Ingest(
				&stagedUpload{data: dataB, copied: bCopied},
				pathB,
				specB.Filename,
				concurrentUploadSize,
				raw.DefaultLimits(),
			)
			bDone <- ingestCall{out: out, err: err}
		}()

		gotA := waitForIngest(t, aDone, "first upload")
		gotB := waitForIngest(t, bDone, "second upload")
		if gotA.err != nil {
			t.Fatalf("attempt %d: first upload: %v", attempt, gotA.err)
		}
		if gotB.err != nil {
			t.Fatalf("attempt %d: second upload: %v", attempt, gotB.err)
		}
		if gotA.out.Result.Model != specA.Model || gotA.out.Result.ISO != int(specA.ISO) {
			t.Fatalf("attempt %d: first upload metadata = model %q ISO %d, want model %q ISO %d",
				attempt, gotA.out.Result.Model, gotA.out.Result.ISO, specA.Model, specA.ISO)
		}
		if gotB.out.Result.Model != specB.Model || gotB.out.Result.ISO != int(specB.ISO) {
			t.Fatalf("attempt %d: second upload metadata = model %q ISO %d, want model %q ISO %d",
				attempt, gotB.out.Result.Model, gotB.out.Result.ISO, specB.Model, specB.ISO)
		}
		assertStoredUpload(t, pathA, dataA)
		assertStoredUpload(t, pathB, dataB)
	}
}

func concurrentSpec(filename, model string, iso uint16, seed int) sample.Spec {
	return sample.Spec{
		Filename:  filename,
		Format:    "DNG",
		Make:      "ColorPixel Test",
		Model:     model,
		Lens:      "35mm Test Lens",
		ApertureN: 28,
		ApertureD: 10,
		ShutterN:  1,
		ShutterD:  250,
		ISO:       iso,
		FocalN:    35,
		FocalD:    1,
		Focal35:   35,
		ShotAt:    time.Date(2026, 8, 25, 10, seed%60, 0, 0, timeutil.Beijing),
		Width:     640,
		Height:    480,
		Seed:      seed,
	}
}

func renderPaddedUpload(t *testing.T, spec sample.Spec) []byte {
	t.Helper()
	data, err := sample.Render(spec)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > concurrentUploadSize {
		t.Fatalf("sample %s is %d bytes, exceeds staged upload size", spec.Filename, len(data))
	}
	padded := make([]byte, concurrentUploadSize)
	copy(padded, data)
	return padded
}

func waitForSignal(t *testing.T, ch <-chan struct{}, stage string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", stage)
	}
}

func waitForIngest(t *testing.T, ch <-chan ingestCall, stage string) ingestCall {
	t.Helper()
	select {
	case got := <-ch:
		return got
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", stage)
		return ingestCall{}
	}
}

func assertStoredUpload(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("stored upload %s differs from its request body", filepath.Base(path))
	}
}
