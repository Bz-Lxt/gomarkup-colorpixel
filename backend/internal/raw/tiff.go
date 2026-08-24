package raw

import (
	"encoding/binary"
	"fmt"
)

const (
	tagImageWidth               = 0x0100
	tagImageLength              = 0x0101
	tagCompression              = 0x0103
	tagMake                     = 0x010F
	tagModel                    = 0x0110
	tagOrientation              = 0x0112
	tagStripOffsets             = 0x0111
	tagStripByteCounts          = 0x0117
	tagDateTime                 = 0x0132
	tagSubIFD                   = 0x014A
	tagJPEGIFOffset             = 0x0201
	tagJPEGIFByteCount          = 0x0202
	tagJPEGTables               = 0x01B5
	tagExifIFD                  = 0x8769
	tagGPSIFD                   = 0x8825
	tagMakerNote                = 0x927C
	tagExposureTime             = 0x829A
	tagFNumber                  = 0x829D
	tagISO                      = 0x8827
	tagDateTimeOriginal         = 0x9003
	tagExposureBias             = 0x9204
	tagFocalLength              = 0x920A
	tagWhiteBalance             = 0xA403
	tagFocalLength35mm          = 0xA405
	tagLensSpecification        = 0xA432
	tagLensModel                = 0xA434
	tagPreviewImageStartSony    = 0x2001
	tagDNGPreview               = 0xC727
	tagNewSubfileType           = 0x00FE
)

type tiffFile struct {
	order  binary.ByteOrder
	base   int64
	lim    Limits
	ra     RandomAccess
	ifds   []*ifd
	seen   int
}

func openTIFF(ra RandomAccess, off int64, lim Limits) (*tiffFile, error) {
	head, err := readExact(ra, off, 8)
	if err != nil {
		return nil, wrap("tiff.header", err)
	}
	o, err := orderFromMarker(head)
	if err != nil {
		return nil, wrap("tiff.endian", err)
	}
	magic := o.Uint16(head[2:])
	if magic != 42 {
		return nil, wrap("tiff.magic", fmt.Errorf("magic %d not 42", magic))
	}
	ifd0 := o.Uint32(head[4:])
	tf := &tiffFile{order: o, base: off, lim: lim, ra: ra}
	if err := tf.walkIFDs(off+int64(ifd0), 0); err != nil && len(tf.ifds) == 0 {
		return nil, err
	}
	return tf, nil
}

func (tf *tiffFile) walkIFDs(off int64, depth int) error {
	if depth > tf.lim.MaxDepth {
		return wrap("tiff.walk", fmt.Errorf("IFD depth %d", depth))
	}
	guard := 0
	for off != tf.base && off != 0 {
		if guard > tf.lim.MaxIFDs {
			return wrap("tiff.walk", fmt.Errorf("IFD chain too long"))
		}
		d, err := parseIFD(tf.ra, off, tf.order, tf.lim, &tf.seen)
		if err != nil {
			return err
		}
		tf.ifds = append(tf.ifds, d)
		if sub := d.find(tagSubIFD); sub != nil {
			vals, _, err := sub.payload(tf.ra, tf.order, tf.base, tf.lim)
			if err == nil {
				offs := decodeLongs(vals, tf.order, int(sub.Count))
				if sub.Count == 1 && sub.Type == 4 && int64(sub.Count)*4 <= 4 {
					offs = []uint32{sub.Raw}
				}
				for _, so := range offs {
					_ = tf.walkIFDs(tf.base+int64(so), depth+1)
				}
			}
		}
		if ex := d.find(tagExifIFD); ex != nil {
			_ = tf.walkIFDs(tf.base+int64(ex.Raw), depth+1)
		}
		if d.Next == 0 {
			break
		}
		off = tf.base + int64(d.Next)
		guard++
	}
	return nil
}

func parseTIFFFamily(ra RandomAccess, window int64, lim Limits, format Format) (*Result, error) {
	tf, err := openTIFF(ra, 0, lim)
	if err != nil {
		return nil, err
	}
	res := &Result{Format: format, Tags: map[string]any{}, ExtractionMode: ModeStream}
	applyIFDs(tf, res)
	enrichTagNames(res)
	if err := extractPreview(tf, res, window, lim); err != nil {
		res.warn(err.Error())
	}
	return res, nil
}
