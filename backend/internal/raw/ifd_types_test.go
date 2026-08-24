package raw

import (
	"encoding/binary"
	"testing"
)

func TestTypeSizeTable(t *testing.T) {
	cases := []struct {
		t uint16
		n int
	}{
		{1, 1}, {2, 1}, {3, 2}, {4, 4}, {5, 8}, {6, 1}, {7, 1}, {8, 2}, {9, 4}, {10, 8}, {11, 4}, {12, 8},
	}
	for _, c := range cases {
		if typeSize(c.t) != c.n {
			t.Fatalf("type %d size %d", c.t, typeSize(c.t))
		}
	}
}

func TestDecodeHelpers(t *testing.T) {
	o := binary.LittleEndian
	b := make([]byte, 8)
	o.PutUint32(b[0:], 1)
	o.PutUint32(b[4:], 200)
	if decodeRational(b, o, false) != 1.0/200.0 {
		t.Fatal(decodeRational(b, o, false))
	}
	o.PutUint32(b[0:], uint32(^uint32(9))) // -10 two's complement
	o.PutUint32(b[4:], 100)
	if decodeRational(b, o, true) != -0.1 {
		t.Fatal(decodeRational(b, o, true))
	}
	if shutterText(0.005) != "1/200" {
		t.Fatal(shutterText(0.005))
	}
	if shutterText(2) != "2.0s" {
		t.Fatal(shutterText(2))
	}
	if decodeASCII([]byte{'A', 'B', 0, 'C'}) != "AB" {
		t.Fatal(decodeASCII([]byte{'A', 'B', 0}))
	}
}

func TestInWindow(t *testing.T) {
	if !inWindow(0, 16, 16) {
		t.Fatal("edge")
	}
	if inWindow(10, 16, 16) {
		t.Fatal("overflow")
	}
	if inWindow(-1, 1, 16) {
		t.Fatal("neg")
	}
}

func TestOrderFromMarker(t *testing.T) {
	o, err := orderFromMarker([]byte{'I', 'I'})
	if err != nil || o != binary.LittleEndian {
		t.Fatal(err)
	}
	o, err = orderFromMarker([]byte{'M', 'M'})
	if err != nil || o != binary.BigEndian {
		t.Fatal(err)
	}
	if _, err := orderFromMarker([]byte{'X', 'X'}); err == nil {
		t.Fatal("bad")
	}
}

func TestDetectExt(t *testing.T) {
	if detectExt("a.cr3") != FormatCR3 || detectExt("b.NEF") != FormatNEF {
		t.Fatal("ext")
	}
	if detectExt("noext") != FormatUNK {
		t.Fatal("unk")
	}
}

func TestBytesAccess(t *testing.T) {
	ra := NewBytes([]byte{1, 2, 3, 4})
	p := make([]byte, 2)
	n, err := ra.ReadAt(p, 2)
	if err != nil || n != 2 || p[0] != 3 {
		t.Fatalf("%d %v %v", n, err, p)
	}
	if _, err := ra.ReadAt(p, 8); err == nil {
		t.Fatal("eof")
	}
}
