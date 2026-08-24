package ingest

import (
	"io"
	"os"
	"sync"
)

const captureBufferSize = 32 * 1024

var captureBuffers = sync.Pool{
	New: func() any { return new([captureBufferSize]byte) },
}

func captureBuffer() []byte {
	buf := captureBuffers.Get().(*[captureBufferSize]byte)
	captureBuffers.Put(buf)
	return buf[:]
}

// Window keeps the first N bytes of a stream in memory and the rest on disk.
// The original stream is never fully buffered.
type Window struct {
	head []byte
	file *os.File
	size int64
}

func Capture(r io.Reader, destPath string, windowBytes int) (*Window, error) {
	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}
	w := &Window{file: f, head: make([]byte, 0, windowBytes)}
	buf := captureBuffer()
	tee := io.TeeReader(r, f)
	for {
		n, err := tee.Read(buf)
		if n > 0 {
			w.size += int64(n)
			if len(w.head) < windowBytes {
				need := windowBytes - len(w.head)
				if n < need {
					w.head = append(w.head, buf[:n]...)
				} else {
					w.head = append(w.head, buf[:need]...)
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = f.Close()
			return nil, err
		}
	}
	return w, nil
}

func (w *Window) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, io.ErrUnexpectedEOF
	}
	headN := int64(len(w.head))
	if off < headN {
		n := copy(p, w.head[off:])
		if n == len(p) {
			return n, nil
		}
		if w.file == nil {
			return n, io.EOF
		}
		m, err := w.file.ReadAt(p[n:], off+int64(n))
		return n + m, err
	}
	if w.file == nil {
		return 0, io.EOF
	}
	return w.file.ReadAt(p, off)
}

func (w *Window) Size() int64 { return w.size }

func (w *Window) WindowSize() int64 { return int64(len(w.head)) }

func (w *Window) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (w *Window) Path() string {
	if w.file == nil {
		return ""
	}
	return w.file.Name()
}
