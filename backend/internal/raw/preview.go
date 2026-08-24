package raw

import (
	"bytes"
	"fmt"
)

func extractPreview(tf *tiffFile, res *Result, window int64, lim Limits) error {
	type cand struct {
		off, n int64
		score  int
	}
	var cs []cand
	for _, d := range tf.ifds {
		offE := d.find(tagJPEGIFOffset)
		lenE := d.find(tagJPEGIFByteCount)
		if offE != nil && lenE != nil {
			off := tf.base + int64(offE.Raw)
			n := int64(lenE.Raw)
			if lenE.Type == 4 && int64(lenE.Count)*4 > 4 {
				if b, _, err := lenE.payload(tf.ra, tf.order, tf.base, lim); err == nil {
					ls := decodeLongs(b, tf.order, int(lenE.Count))
					if len(ls) > 0 {
						n = int64(ls[0])
					}
				}
			}
			score := 1
			if n > 100_000 {
				score = 3
			}
			cs = append(cs, cand{off, n, score})
		}
		comp := d.find(tagCompression)
		if comp != nil && (comp.Raw == 6 || firstU16(mustPayload(tf, *comp), tf.order) == 6) {
			so := d.find(tagStripOffsets)
			sc := d.find(tagStripByteCounts)
			if so != nil && sc != nil {
				off := tf.base + int64(so.Raw)
				n := int64(sc.Raw)
				cs = append(cs, cand{off, n, 2})
			}
		}
	}
	if len(cs) == 0 {
		res.ExtractionMode = ModeNone
		return fmt.Errorf("no embedded JPEG preview")
	}
	best := cs[0]
	for _, c := range cs[1:] {
		if c.score > best.score || (c.score == best.score && c.n > best.n) {
			best = c
		}
	}
	if best.n <= 0 || best.n > int64(lim.PreviewMax) {
		return fmt.Errorf("preview size %d rejected", best.n)
	}
	mode := ModeStream
	if !inWindow(best.off, int(best.n), window) {
		mode = ModeDeferred
	}
	raw, err := readExact(tf.ra, best.off, int(best.n))
	if err != nil {
		return err
	}
	if i := bytes.Index(raw, []byte{0xFF, 0xD8}); i >= 0 {
		raw = raw[i:]
	}
	if !bytes.HasPrefix(raw, []byte{0xFF, 0xD8}) {
		return fmt.Errorf("preview is not JPEG")
	}
	res.Preview = raw
	res.ExtractionMode = mode
	return nil
}

func mustPayload(tf *tiffFile, e ifdEntry) []byte {
	b, _, err := e.payload(tf.ra, tf.order, tf.base, tf.lim)
	if err != nil {
		return nil
	}
	return b
}

func extractJPEGAt(ra RandomAccess, off int64, n int, window int64, lim Limits) ([]byte, string, error) {
	if n <= 0 || n > lim.PreviewMax {
		return nil, ModeNone, fmt.Errorf("preview size %d rejected", n)
	}
	mode := ModeStream
	if !inWindow(off, n, window) {
		mode = ModeDeferred
	}
	b, err := readExact(ra, off, n)
	if err != nil {
		return nil, mode, err
	}
	if i := bytes.Index(b, []byte{0xFF, 0xD8}); i >= 0 {
		b = b[i:]
	}
	if !bytes.HasPrefix(b, []byte{0xFF, 0xD8}) {
		return nil, mode, fmt.Errorf("not jpeg")
	}
	return b, mode, nil
}

func findJPEGIn(b []byte) []byte {
	i := bytes.Index(b, []byte{0xFF, 0xD8})
	if i < 0 {
		return nil
	}
	j := bytes.LastIndex(b, []byte{0xFF, 0xD9})
	if j > i {
		return b[i : j+2]
	}
	return b[i:]
}
