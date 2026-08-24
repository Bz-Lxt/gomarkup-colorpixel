package raw

import (
	"io"
)

// RandomAccess is a bounded, offset-addressable view of a RAW container.
// Implementations must not load the entire backing file into memory.
type RandomAccess interface {
	ReadAt(p []byte, off int64) (n int, err error)
	Size() int64
}

type sliceAccess struct {
	b []byte
}

func NewBytes(b []byte) RandomAccess { return sliceAccess{b: b} }

func (s sliceAccess) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, io.ErrUnexpectedEOF
	}
	if off >= int64(len(s.b)) {
		return 0, io.EOF
	}
	n := copy(p, s.b[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (s sliceAccess) Size() int64 { return int64(len(s.b)) }

func readExact(ra RandomAccess, off int64, n int) ([]byte, error) {
	if n < 0 {
		return nil, errShort
	}
	if n == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, n)
	got, err := ra.ReadAt(buf, off)
	if got != n {
		if err == nil {
			err = io.ErrUnexpectedEOF
		}
		return buf[:got], wrap("readExact", err)
	}
	return buf, nil
}

func inWindow(off int64, length int, window int64) bool {
	if off < 0 || length < 0 {
		return false
	}
	end := off + int64(length)
	return end <= window
}
