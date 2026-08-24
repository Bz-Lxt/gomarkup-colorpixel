package raw

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type box struct {
	Type   string
	Offset int64
	Header int64
	Size   int64
	UUID   []byte
}

func walkBoxes(ra RandomAccess, start, end int64, depth int, lim Limits, fn func(box) error) error {
	if depth > lim.MaxDepth {
		return wrap("bmff", fmt.Errorf("box nest depth %d", depth))
	}
	off := start
	seen := 0
	for off < end {
		if seen > 1024 {
			return wrap("bmff", fmt.Errorf("too many boxes"))
		}
		seen++
		hdr, err := readExact(ra, off, 8)
		if err != nil {
			if err == io.EOF || off+8 > end {
				return nil
			}
			return wrap("bmff.header", err)
		}
		sz := int64(binary.BigEndian.Uint32(hdr[0:4]))
		typ := string(hdr[4:8])
		header := int64(8)
		if sz == 1 {
			ext, err := readExact(ra, off+8, 8)
			if err != nil {
				return wrap("bmff.largesize", err)
			}
			usz, _ := readU64(ext, binary.BigEndian)
			sz = int64(usz)
			header = 16
		} else if sz == 0 {
			sz = end - off
		}
		if sz < header || off+sz > end+8 && end > 0 {
			if sz < header {
				return wrap("bmff", fmt.Errorf("box %s size %d", typ, sz))
			}
		}
		b := box{Type: typ, Offset: off, Header: header, Size: sz}
		if typ == "uuid" {
			u, err := readExact(ra, off+header, 16)
			if err == nil {
				b.UUID = u
				b.Header += 16
			}
		}
		if err := fn(b); err != nil {
			return err
		}
		if isContainer(typ) {
			innerEnd := off + sz
			if innerEnd > end && end > 0 {
				innerEnd = end
			}
			if err := walkBoxes(ra, off+b.Header, innerEnd, depth+1, lim, fn); err != nil {
				return err
			}
		}
		if sz <= 0 {
			break
		}
		off += sz
	}
	return nil
}

func isContainer(typ string) bool {
	switch typ {
	case "moov", "trak", "mdia", "minf", "stbl", "dinf", "udta", "moof", "traf":
		return true
	default:
		return false
	}
}

func boxPayload(ra RandomAccess, b box, max int) ([]byte, error) {
	n := b.Size - b.Header
	if n <= 0 {
		return nil, nil
	}
	if n > int64(max) {
		n = int64(max)
	}
	return readExact(ra, b.Offset+b.Header, int(n))
}

var canonUUID = []byte{0x85, 0xc0, 0xb6, 0x87, 0x82, 0x0f, 0x11, 0xe0, 0x81, 0x11, 0xf4, 0xce, 0x46, 0x2b, 0x6a, 0x48}

func isCanonUUID(u []byte) bool {
	return bytes.Equal(u, canonUUID)
}
