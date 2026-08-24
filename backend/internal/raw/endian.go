package raw

import (
	"encoding/binary"
	"errors"
)

var (
	errShort     = errors.New("buffer too short")
	errBadEndian = errors.New("unrecognized byte order")
)

func readU16(b []byte, o binary.ByteOrder) (uint16, error) {
	if len(b) < 2 {
		return 0, errShort
	}
	return o.Uint16(b), nil
}

func readU32(b []byte, o binary.ByteOrder) (uint32, error) {
	if len(b) < 4 {
		return 0, errShort
	}
	return o.Uint32(b), nil
}

func readU64(b []byte, o binary.ByteOrder) (uint64, error) {
	if len(b) < 8 {
		return 0, errShort
	}
	return o.Uint64(b), nil
}

func orderFromMarker(b []byte) (binary.ByteOrder, error) {
	if len(b) < 2 {
		return nil, errShort
	}
	if b[0] == 'I' && b[1] == 'I' {
		return binary.LittleEndian, nil
	}
	if b[0] == 'M' && b[1] == 'M' {
		return binary.BigEndian, nil
	}
	return nil, errBadEndian
}

func typeSize(t uint16) int {
	switch t {
	case 1, 2, 6, 7:
		return 1
	case 3, 8:
		return 2
	case 4, 9, 11:
		return 4
	case 5, 10, 12:
		return 8
	default:
		return 1
	}
}
