package raw

import (
	"fmt"
	"time"
)

const (
	ModeStream   = "stream"
	ModeDeferred = "deferred"
	ModeNone     = "none"
)

type Format string

const (
	FormatCR3 Format = "CR3"
	FormatCR2 Format = "CR2"
	FormatNEF Format = "NEF"
	FormatARW Format = "ARW"
	FormatDNG Format = "DNG"
	FormatUNK Format = "UNKNOWN"
)

type Limits struct {
	MaxIFDs       int
	MaxDepth      int
	MaxAlloc      int
	PreviewMax    int
	WindowBytes   int
}

func DefaultLimits() Limits {
	return Limits{
		MaxIFDs:     64,
		MaxDepth:    8,
		MaxAlloc:    1_000_000,
		PreviewMax:  8 << 20,
		WindowBytes: 16 << 20,
	}
}

type Result struct {
	Format          Format
	ExtractionMode  string
	Make            string
	Model           string
	LensModel       string
	LensSpec        string
	Aperture        float64
	ShutterText     string
	ShutterSeconds  float64
	ISO             int
	FocalLength     float64
	FocalLength35mm float64
	DateTimeOriginal time.Time
	Orientation     int
	WhiteBalance    string
	ExposureBias    float64
	Width           int
	Height          int
	Preview         []byte
	Tags            map[string]any
	Warnings        []string
}

func (r *Result) warn(msg string) {
	r.Warnings = append(r.Warnings, msg)
}

type ParseError struct {
	Op  string
	Err error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("raw parse %s: %v", e.Op, e.Err)
}

func (e *ParseError) Unwrap() error { return e.Err }

func wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return &ParseError{Op: op, Err: err}
}
