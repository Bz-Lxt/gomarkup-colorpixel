package raw

import (
	"bytes"
	"strings"
)

func parseMakerNotes(tf *tiffFile, res *Result, window int64, lim Limits) {
	for _, d := range tf.ifds {
		e := d.find(tagMakerNote)
		if e == nil {
			continue
		}
		b, off, err := e.payload(tf.ra, tf.order, tf.base, lim)
		if err != nil || len(b) < 8 {
			continue
		}
		make := strings.ToLower(res.Make)
		switch {
		case strings.Contains(make, "canon") || bytes.HasPrefix(b, []byte("Canon")):
			applyCanonMakerNote(tf, b, off, res, window, lim)
		case strings.Contains(make, "nikon") || bytes.HasPrefix(b, []byte("Nikon")):
			applyNikonMakerNote(tf, b, off, res, window, lim)
		case strings.Contains(make, "sony") || bytes.HasPrefix(b, []byte("SONY")):
			applySonyMakerNote(tf, b, off, res, window, lim)
		default:
			if bytes.HasPrefix(b, []byte("Nikon")) {
				applyNikonMakerNote(tf, b, off, res, window, lim)
			}
		}
	}
}
