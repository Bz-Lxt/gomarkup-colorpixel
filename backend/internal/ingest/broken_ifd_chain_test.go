package ingest_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"path/filepath"
	"testing"

	"colorpixel/internal/ingest"
	"colorpixel/internal/raw"
)

func TestIngestRejectsBrokenIFDChain(t *testing.T) {
	data := make([]byte, 26)
	copy(data[0:2], "II")
	binary.LittleEndian.PutUint16(data[2:4], 42)
	binary.LittleEndian.PutUint32(data[4:8], 8)
	binary.LittleEndian.PutUint16(data[8:10], 1)
	binary.LittleEndian.PutUint16(data[10:12], 0x010f)
	binary.LittleEndian.PutUint16(data[12:14], 2)
	binary.LittleEndian.PutUint32(data[14:18], 4)
	copy(data[18:22], []byte{'N', 'I', 'K', 0})
	binary.LittleEndian.PutUint32(data[22:26], 1<<20)

	lim := raw.DefaultLimits()
	outcome, err := ingest.Ingest(
		bytes.NewReader(data),
		filepath.Join(t.TempDir(), "recovered.nef"),
		"recovered.nef",
		lim.WindowBytes,
		lim,
	)
	if err == nil {
		t.Fatalf("Ingest() accepted a broken IFD chain: %+v", outcome)
	}
	var parseErr *raw.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("Ingest() error = %T %v, want *raw.ParseError", err, err)
	}
	if outcome != nil {
		t.Fatalf("Ingest() returned an outcome after structural failure: %+v", outcome)
	}
}
