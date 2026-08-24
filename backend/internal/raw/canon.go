package raw

func applyCanonMakerNote(tf *tiffFile, payload []byte, base int64, res *Result, window int64, lim Limits) {
	off := 0
	if len(payload) > 6 && string(payload[:5]) == "Canon" {
		off = 6
	}
	if off+8 > len(payload) {
		return
	}
	sub, err := openTIFF(NewBytes(payload[off:]), 0, lim)
	if err != nil {
		head := payload[off:]
		if len(head) >= 2 && (head[0] == 'I' || head[0] == 'M') {
			return
		}
		d, err := parseIFD(NewBytes(payload), int64(off), tf.order, lim, new(int))
		if err != nil {
			return
		}
		harvestMakerIFD(tf, d, int64(off), res)
		return
	}
	for _, d := range sub.ifds {
		harvestMakerIFD(tf, d, base+int64(off), res)
	}
	_ = window
}

func harvestMakerIFD(tf *tiffFile, d *ifd, _ int64, res *Result) {
	if e := d.find(0x0095); e != nil && res.LensModel == "" {
		if b, _, err := e.payload(tf.ra, tf.order, tf.base, tf.lim); err == nil {
			res.LensModel = decodeASCII(b)
		}
	}
}
