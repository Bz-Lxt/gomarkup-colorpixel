package raw

func parseCR2(ra RandomAccess, window int64, lim Limits) (*Result, error) {
	res, err := parseTIFFFamily(ra, window, lim, FormatCR2)
	if err != nil {
		return nil, err
	}
	if tf, e := openTIFF(ra, 0, lim); e == nil {
		parseMakerNotes(tf, res, window, lim)
	}
	return res, nil
}

func parseDNG(ra RandomAccess, window int64, lim Limits) (*Result, error) {
	res, err := parseTIFFFamily(ra, window, lim, FormatDNG)
	if err != nil {
		return nil, err
	}
	if tf, e := openTIFF(ra, 0, lim); e == nil {
		parseMakerNotes(tf, res, window, lim)
	}
	return res, nil
}
