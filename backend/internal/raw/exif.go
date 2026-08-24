package raw

import (
	"colorpixel/internal/timeutil"
	"fmt"
)

func applyIFDs(tf *tiffFile, res *Result) {
	for _, d := range tf.ifds {
		for _, e := range d.Entries {
			b, _, err := e.payload(tf.ra, tf.order, tf.base, tf.lim)
			if err != nil {
				continue
			}
			key := fmt.Sprintf("0x%04X", e.Tag)
			switch e.Type {
			case 2:
				res.Tags[key] = decodeASCII(b)
			case 3:
				s := decodeShorts(b, tf.order, int(e.Count))
				if len(s) == 1 {
					res.Tags[key] = s[0]
				} else {
					res.Tags[key] = s
				}
			case 4:
				l := decodeLongs(b, tf.order, int(e.Count))
				if len(l) == 1 {
					res.Tags[key] = l[0]
				} else {
					res.Tags[key] = l
				}
			case 5:
				rs := decodeRationals(b, tf.order, int(e.Count), false)
				if len(rs) == 1 {
					res.Tags[key] = rs[0]
				} else {
					res.Tags[key] = rs
				}
			case 10:
				rs := decodeRationals(b, tf.order, int(e.Count), true)
				if len(rs) == 1 {
					res.Tags[key] = rs[0]
				} else {
					res.Tags[key] = rs
				}
			default:
				if e.Count <= 16 && typeSize(e.Type)*int(e.Count) <= 64 {
					res.Tags[key] = fmt.Sprintf("type=%d count=%d", e.Type, e.Count)
				}
			}
			switch e.Tag {
			case tagMake:
				res.Make = decodeASCII(b)
			case tagModel:
				res.Model = decodeASCII(b)
			case tagLensModel:
				res.LensModel = decodeASCII(b)
			case tagLensSpecification:
				rs := decodeRationals(b, tf.order, int(e.Count), false)
				if len(rs) >= 4 {
					res.LensSpec = fmt.Sprintf("%.0f-%.0fmm f/%.1f-%.1f", rs[0], rs[1], rs[2], rs[3])
				}
			case tagFNumber:
				res.Aperture = decodeRational(b, tf.order, false)
			case tagExposureTime:
				res.ShutterSeconds = decodeRational(b, tf.order, false)
				res.ShutterText = shutterText(res.ShutterSeconds)
			case tagISO:
				if e.Type == 3 {
					res.ISO = int(firstU16(b, tf.order))
				} else {
					res.ISO = int(firstU32(b, tf.order))
				}
			case tagFocalLength:
				res.FocalLength = decodeRational(b, tf.order, false)
			case tagFocalLength35mm:
				res.FocalLength35mm = float64(firstU16(b, tf.order))
			case tagDateTimeOriginal, tagDateTime:
				if res.DateTimeOriginal.IsZero() {
					t, err := timeutil.ParseEXIFDate(decodeASCII(b))
					if err == nil {
						res.DateTimeOriginal = t
					}
				}
			case tagOrientation:
				res.Orientation = int(firstU16(b, tf.order))
			case tagWhiteBalance:
				if firstU16(b, tf.order) == 0 {
					res.WhiteBalance = "auto"
				} else {
					res.WhiteBalance = "manual"
				}
			case tagExposureBias:
				res.ExposureBias = decodeRational(b, tf.order, true)
			case tagImageWidth:
				if res.Width == 0 {
					if e.Type == 3 {
						res.Width = int(firstU16(b, tf.order))
					} else {
						res.Width = int(firstU32(b, tf.order))
					}
				}
			case tagImageLength:
				if res.Height == 0 {
					if e.Type == 3 {
						res.Height = int(firstU16(b, tf.order))
					} else {
						res.Height = int(firstU32(b, tf.order))
					}
				}
			}
		}
	}
}
