package raw

import (
	"bytes"
	"fmt"
)

func parseCR3(ra RandomAccess, window int64, lim Limits) (*Result, error) {
	res := &Result{Format: FormatCR3, Tags: map[string]any{}, ExtractionMode: ModeNone}
	end := ra.Size()
	var previewOff int64
	var previewN int
	err := walkBoxes(ra, 0, end, 0, lim, func(b box) error {
		typ := b.Type
		switch typ {
		case "PRVW", "THMB":
			n := int(b.Size - b.Header)
			if n > lim.PreviewMax {
				n = lim.PreviewMax
			}
			if n > 0 && (previewN == 0 || typ == "PRVW") {
				previewOff = b.Offset + b.Header
				previewN = n
			}
		case "uuid":
			payload, err := boxPayload(ra, b, lim.PreviewMax)
			if err != nil || len(payload) < 4 {
				return nil
			}
			name := ""
			rest := payload
			if len(payload) >= 4 && isFourCC(payload[:4]) {
				name = string(payload[:4])
				rest = payload[4:]
			}
			if name == "CMT1" || name == "CMT2" || name == "CMT3" || name == "CMT4" || looksTIFF(rest) {
				if looksTIFF(rest) {
					sub := NewBytes(rest)
					tf, err := openTIFF(sub, 0, lim)
					if err == nil {
						applyIFDs(tf, res)
						if res.Preview == nil {
							_ = extractPreview(tf, res, int64(len(rest)), lim)
						}
					}
				}
			}
			if jpeg := findJPEGIn(payload); jpeg != nil && (res.Preview == nil || len(jpeg) > len(res.Preview)) {
				mode := ModeStream
				if !inWindow(b.Offset, int(b.Size), window) {
					mode = ModeDeferred
				}
				res.Preview = jpeg
				res.ExtractionMode = mode
			}
		}
		if typ == "uuid" || typ == "moov" {
			return nil
		}
		return nil
	})
	if err != nil {
		return res, err
	}
	if res.Preview == nil && previewN > 0 {
		b, mode, err := extractJPEGAt(ra, previewOff, previewN, window, lim)
		if err == nil {
			res.Preview = b
			res.ExtractionMode = mode
		} else {
			res.warn(err.Error())
		}
	}
	if res.Make == "" {
		res.Make = "Canon"
	}
	if res.ExtractionMode == "" {
		res.ExtractionMode = ModeNone
	}
	if res.Preview == nil && previewN == 0 {
		res.warn("CR3 preview box not found")
	}
	_ = isCanonUUID
	return res, nil
}

func looksTIFF(b []byte) bool {
	return len(b) >= 8 && ((b[0] == 'I' && b[1] == 'I') || (b[0] == 'M' && b[1] == 'M'))
}

func isFourCC(b []byte) bool {
	if len(b) < 4 {
		return false
	}
	for _, c := range b[:4] {
		if c < 0x20 || c > 0x7e {
			return false
		}
	}
	return true
}

func (r *Result) ensureCR3() {
	if r.Format == "" {
		r.Format = FormatCR3
	}
}

var _ = bytes.Equal
var _ = fmt.Sprintf
