package raw

import (
	"bytes"
	"strings"
)

func Detect(head []byte, filename string) Format {
	if f := detectMagic(head); f != FormatUNK {
		return f
	}
	return detectExt(filename)
}

func detectMagic(head []byte) Format {
	if len(head) >= 12 && string(head[4:8]) == "ftyp" {
		brand := string(head[8:min(12, len(head))])
		if brand == "crx " || brand == "CRX " || bytes.Contains(head[:min(32, len(head))], []byte("crx ")) {
			return FormatCR3
		}
		if bytes.Contains(head[:min(64, len(head))], []byte("heic")) {
			return FormatUNK
		}
	}
	if len(head) >= 8 {
		if (head[0] == 'I' && head[1] == 'I' && head[2] == 0x2a && head[3] == 0x00) ||
			(head[0] == 'M' && head[1] == 'M' && head[2] == 0x00 && head[3] == 0x2a) {
			return FormatUNK
		}
	}
	return FormatUNK
}

func detectExt(name string) Format {
	i := strings.LastIndex(name, ".")
	if i < 0 {
		return FormatUNK
	}
	switch strings.ToUpper(name[i+1:]) {
	case "CR3":
		return FormatCR3
	case "CR2":
		return FormatCR2
	case "NEF":
		return FormatNEF
	case "ARW":
		return FormatARW
	case "DNG":
		return FormatDNG
	default:
		return FormatUNK
	}
}

func Parse(ra RandomAccess, filename string, windowBytes int64, lim Limits) (*Result, error) {
	if lim.MaxIFDs == 0 {
		lim = DefaultLimits()
	}
	head, _ := readExact(ra, 0, 64)
	if err := rejectBigTIFF(head); err != nil {
		return nil, err
	}
	format := Detect(head, filename)
	switch format {
	case FormatCR3:
		return parseCR3(ra, windowBytes, lim)
	case FormatCR2:
		return parseCR2(ra, windowBytes, lim)
	case FormatNEF:
		return parseNEF(ra, windowBytes, lim)
	case FormatARW:
		return parseARW(ra, windowBytes, lim)
	case FormatDNG:
		return parseDNG(ra, windowBytes, lim)
	default:
		if len(head) >= 4 && ((head[0] == 'I' && head[1] == 'I') || (head[0] == 'M' && head[1] == 'M')) {
			res, err := parseTIFFFamily(ra, windowBytes, lim, FormatUNK)
			if res != nil {
				res.Format = FormatUNK
			}
			return res, err
		}
		res := &Result{Format: FormatUNK, ExtractionMode: ModeNone, Tags: map[string]any{}}
		res.warn("unrecognized container; file-level metadata only")
		return res, nil
	}
}
