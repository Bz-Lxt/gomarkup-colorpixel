package raw

import (
	"encoding/binary"
	"fmt"
	"math"
)

type ifdEntry struct {
	Tag   uint16
	Type  uint16
	Count uint32
	Raw   uint32
	Off   int64
}

type ifd struct {
	Offset  int64
	Entries []ifdEntry
	Next    uint32
}

func parseIFD(ra RandomAccess, off int64, o binary.ByteOrder, lim Limits, seen *int) (*ifd, error) {
	if off < 0 || off >= ra.Size() {
		return nil, wrap("ifd", fmt.Errorf("offset %d out of range", off))
	}
	*seen++
	if *seen > lim.MaxIFDs {
		return nil, wrap("ifd", fmt.Errorf("too many IFDs (>%d)", lim.MaxIFDs))
	}
	hb, err := readExact(ra, off, 2)
	if err != nil {
		return nil, wrap("ifd.count", err)
	}
	count, _ := readU16(hb, o)
	if int(count) > 512 {
		return nil, wrap("ifd", fmt.Errorf("IFD entry count %d exceeds 512", count))
	}
	need := 2 + int(count)*12 + 4
	body, err := readExact(ra, off, need)
	if err != nil {
		return nil, wrap("ifd.body", err)
	}
	d := &ifd{Offset: off, Entries: make([]ifdEntry, 0, count)}
	p := 2
	for i := 0; i < int(count); i++ {
		e := ifdEntry{
			Tag:   o.Uint16(body[p:]),
			Type:  o.Uint16(body[p+2:]),
			Count: o.Uint32(body[p+4:]),
			Raw:   o.Uint32(body[p+8:]),
			Off:   off + int64(p),
		}
		d.Entries = append(d.Entries, e)
		p += 12
	}
	d.Next = o.Uint32(body[p:])
	return d, nil
}

func (d *ifd) find(tag uint16) *ifdEntry {
	for i := range d.Entries {
		if d.Entries[i].Tag == tag {
			return &d.Entries[i]
		}
	}
	return nil
}

func (e ifdEntry) payload(ra RandomAccess, o binary.ByteOrder, base int64, lim Limits) ([]byte, int64, error) {
	sz := typeSize(e.Type)
	if sz <= 0 {
		return nil, 0, wrap("payload", fmt.Errorf("unknown type %d", e.Type))
	}
	if e.Count > uint32(lim.MaxAlloc) {
		return nil, 0, wrap("payload", fmt.Errorf("count %d exceeds alloc cap", e.Count))
	}
	nbytes := int64(sz) * int64(e.Count)
	if nbytes < 0 || nbytes > int64(lim.PreviewMax)*2 {
		return nil, 0, wrap("payload", fmt.Errorf("payload size %d rejected", nbytes))
	}
	if nbytes <= 4 {
		tmp := make([]byte, 4)
		o.PutUint32(tmp, e.Raw)
		return tmp[:nbytes], e.Off + 8, nil
	}
	off := base + int64(e.Raw)
	if off < 0 || off+nbytes > ra.Size() {
		return nil, off, wrap("payload", fmt.Errorf("offset %d+%d out of range", off, nbytes))
	}
	b, err := readExact(ra, off, int(nbytes))
	return b, off, err
}

func decodeASCII(b []byte) string {
	n := 0
	for n < len(b) && b[n] != 0 {
		n++
	}
	return string(b[:n])
}

func decodeShorts(b []byte, o binary.ByteOrder, count int) []uint16 {
	out := make([]uint16, 0, count)
	for i := 0; i < count && (i+1)*2 <= len(b); i++ {
		out = append(out, o.Uint16(b[i*2:]))
	}
	return out
}

func decodeLongs(b []byte, o binary.ByteOrder, count int) []uint32 {
	out := make([]uint32, 0, count)
	for i := 0; i < count && (i+1)*4 <= len(b); i++ {
		out = append(out, o.Uint32(b[i*4:]))
	}
	return out
}

func decodeRational(b []byte, o binary.ByteOrder, signed bool) float64 {
	if len(b) < 8 {
		return 0
	}
	if signed {
		n := int32(o.Uint32(b[0:4]))
		d := int32(o.Uint32(b[4:8]))
		if d == 0 {
			return 0
		}
		return float64(n) / float64(d)
	}
	n := o.Uint32(b[0:4])
	d := o.Uint32(b[4:8])
	if d == 0 {
		return 0
	}
	return float64(n) / float64(d)
}

func decodeRationals(b []byte, o binary.ByteOrder, count int, signed bool) []float64 {
	out := make([]float64, 0, count)
	for i := 0; i < count && (i+1)*8 <= len(b); i++ {
		out = append(out, decodeRational(b[i*8:], o, signed))
	}
	return out
}

func firstU16(b []byte, o binary.ByteOrder) uint16 {
	if len(b) < 2 {
		return 0
	}
	return o.Uint16(b)
}

func firstU32(b []byte, o binary.ByteOrder) uint32 {
	if len(b) < 4 {
		return 0
	}
	return o.Uint32(b)
}

func shutterText(sec float64) string {
	if sec <= 0 {
		return ""
	}
	if sec >= 1 {
		return fmt.Sprintf("%.1fs", sec)
	}
	den := int(math.Round(1 / sec))
	if den < 1 {
		den = 1
	}
	return fmt.Sprintf("1/%d", den)
}
