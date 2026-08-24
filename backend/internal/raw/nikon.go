package raw

import "bytes"

func applyNikonMakerNote(_ *tiffFile, payload []byte, base int64, res *Result, window int64, lim Limits) {
	// Nikon: "Nikon\0" + 0x02 0x00 0x00 0x00 + TIFF (offsets relative to TIFF start)
	p := payload
	if bytes.HasPrefix(p, []byte("Nikon")) {
		i := bytes.Index(p, []byte{'I', 'I'})
		j := bytes.Index(p, []byte{'M', 'M'})
		start := -1
		if i >= 0 && (j < 0 || i < j) {
			start = i
		} else if j >= 0 {
			start = j
		}
		if start < 0 {
			return
		}
		p = p[start:]
		base += int64(start)
	}
	tf, err := openTIFF(NewBytes(p), 0, lim)
	if err != nil {
		return
	}
	applyIFDs(tf, res)
	if res.Preview == nil {
		_ = extractPreview(tf, res, window, lim)
	}
}

func parseNEF(ra RandomAccess, window int64, lim Limits) (*Result, error) {
	res, err := parseTIFFFamily(ra, window, lim, FormatNEF)
	if err != nil {
		return nil, err
	}
	if tf, e := openTIFF(ra, 0, lim); e == nil {
		parseMakerNotes(tf, res, window, lim)
	}
	return res, nil
}
