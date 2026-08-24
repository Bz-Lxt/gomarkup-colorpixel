package raw

func applySonyMakerNote(parent *tiffFile, payload []byte, base int64, res *Result, window int64, lim Limits) {
	d, err := parseIFD(NewBytes(payload), 0, parent.order, lim, new(int))
	if err != nil {
		tf, err := openTIFF(NewBytes(payload), 0, lim)
		if err != nil {
			return
		}
		for _, id := range tf.ifds {
			if e := id.find(tagPreviewImageStartSony); e != nil {
				trySonyPreview(parent, e, id, res, window, lim)
			}
		}
		return
	}
	if e := d.find(tagPreviewImageStartSony); e != nil {
		trySonyPreview(parent, e, d, res, window, lim)
	}
	_ = base
}

func trySonyPreview(tf *tiffFile, e *ifdEntry, d *ifd, res *Result, window int64, lim Limits) {
	off := tf.base + int64(e.Raw)
	n := 0
	if lenE := d.find(0x2002); lenE != nil {
		n = int(lenE.Raw)
	}
	if n <= 0 {
		n = lim.PreviewMax
		if int64(n) > tf.ra.Size()-off {
			n = int(tf.ra.Size() - off)
		}
	}
	b, mode, err := extractJPEGAt(tf.ra, off, n, window, lim)
	if err == nil {
		res.Preview = b
		res.ExtractionMode = mode
	}
}

func parseARW(ra RandomAccess, window int64, lim Limits) (*Result, error) {
	res, err := parseTIFFFamily(ra, window, lim, FormatARW)
	if err != nil {
		return nil, err
	}
	if tf, e := openTIFF(ra, 0, lim); e == nil {
		parseMakerNotes(tf, res, window, lim)
	}
	return res, nil
}
