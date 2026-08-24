package sample

import (
	"encoding/binary"
)

type tiffW struct {
	o   binary.ByteOrder
	buf []byte
}

func newTIFF(le bool) *tiffW {
	w := &tiffW{o: binary.LittleEndian}
	if !le {
		w.o = binary.BigEndian
	}
	w.buf = make([]byte, 8)
	if le {
		w.buf[0], w.buf[1] = 'I', 'I'
	} else {
		w.buf[0], w.buf[1] = 'M', 'M'
	}
	w.o.PutUint16(w.buf[2:], 42)
	return w
}

func (w *tiffW) align2() {
	if len(w.buf)%2 == 1 {
		w.buf = append(w.buf, 0)
	}
}

func (w *tiffW) putU16(v uint16) {
	b := make([]byte, 2)
	w.o.PutUint16(b, v)
	w.buf = append(w.buf, b...)
}

func (w *tiffW) putU32(v uint32) {
	b := make([]byte, 4)
	w.o.PutUint32(b, v)
	w.buf = append(w.buf, b...)
}

type ifdEnt struct {
	tag, typ uint16
	count    uint32
	inline   []byte
	ext      []byte
}

func asciiEnt(tag uint16, s string) ifdEnt {
	b := append([]byte(s), 0)
	return ifdEnt{tag: tag, typ: 2, count: uint32(len(b)), ext: b}
}

func shortEnt(tag uint16, v uint16) ifdEnt {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, v)
	return ifdEnt{tag: tag, typ: 3, count: 1, inline: b}
}

func longEnt(tag uint16, v uint32) ifdEnt {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return ifdEnt{tag: tag, typ: 4, count: 1, inline: b}
}

func ratEnt(tag uint16, n, d uint32) ifdEnt {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint32(b[0:], n)
	binary.LittleEndian.PutUint32(b[4:], d)
	return ifdEnt{tag: tag, typ: 5, count: 1, ext: b}
}

func sratEnt(tag uint16, n, d int32) ifdEnt {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint32(b[0:], uint32(n))
	binary.LittleEndian.PutUint32(b[4:], uint32(d))
	return ifdEnt{tag: tag, typ: 10, count: 1, ext: b}
}

func ratsEnt(tag uint16, nd [][2]uint32) ifdEnt {
	b := make([]byte, 8*len(nd))
	for i, p := range nd {
		binary.LittleEndian.PutUint32(b[i*8:], p[0])
		binary.LittleEndian.PutUint32(b[i*8+4:], p[1])
	}
	return ifdEnt{tag: tag, typ: 5, count: uint32(len(nd)), ext: b}
}

func (w *tiffW) rewriteInline(e *ifdEnt) {
	if w.o == binary.LittleEndian {
		return
	}
	switch e.typ {
	case 3:
		if len(e.inline) >= 2 {
			v := binary.LittleEndian.Uint16(e.inline)
			w.o.PutUint16(e.inline, v)
		}
	case 4:
		if len(e.inline) >= 4 {
			v := binary.LittleEndian.Uint32(e.inline)
			w.o.PutUint32(e.inline, v)
		}
	case 5, 10:
		src := e.ext
		for i := 0; i+8 <= len(src); i += 8 {
			n := binary.LittleEndian.Uint32(src[i:])
			d := binary.LittleEndian.Uint32(src[i+4:])
			w.o.PutUint32(src[i:], n)
			w.o.PutUint32(src[i+4:], d)
		}
	}
}

func (w *tiffW) writeIFD(ents []ifdEnt, next uint32) uint32 {
	for i := range ents {
		w.rewriteInline(&ents[i])
	}
	w.align2()
	start := uint32(len(w.buf))
	w.putU16(uint16(len(ents)))
	slots := make([]int, len(ents))
	for i, e := range ents {
		w.putU16(e.tag)
		w.putU16(e.typ)
		w.putU32(e.count)
		slots[i] = len(w.buf)
		w.putU32(0)
	}
	w.putU32(next)
	for i, e := range ents {
		payload := e.inline
		if len(e.ext) > 0 {
			payload = e.ext
		}
		if len(payload) <= 4 {
			copy(w.buf[slots[i]:slots[i]+4], pad4(payload))
			continue
		}
		w.align2()
		off := uint32(len(w.buf))
		w.buf = append(w.buf, payload...)
		tmp := make([]byte, 4)
		w.o.PutUint32(tmp, off)
		copy(w.buf[slots[i]:slots[i]+4], tmp)
	}
	return start
}

func pad4(b []byte) []byte {
	out := make([]byte, 4)
	copy(out, b)
	return out
}

func (w *tiffW) setIFD0(off uint32) {
	w.o.PutUint32(w.buf[4:], off)
}

func (w *tiffW) appendJPEG(jpeg []byte) uint32 {
	w.align2()
	off := uint32(len(w.buf))
	w.buf = append(w.buf, jpeg...)
	return off
}

func (w *tiffW) padTo(n int) {
	if len(w.buf) >= n {
		return
	}
	w.buf = append(w.buf, make([]byte, n-len(w.buf))...)
}
