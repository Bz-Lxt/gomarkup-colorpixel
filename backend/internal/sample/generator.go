package sample

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"colorpixel/internal/timeutil"
)

type Spec struct {
	Filename         string
	Format           string
	Make             string
	Model            string
	Lens             string
	ApertureN        uint32
	ApertureD        uint32
	ShutterN         uint32
	ShutterD         uint32
	ISO              uint16
	FocalN           uint32
	FocalD           uint32
	Focal35          uint16
	ExposureBiasN    int32
	WhiteBalance     uint16
	ShotAt           time.Time
	Width            int
	Height           int
	Deferred         bool
	Seed             int
	LargeSize        bool
}

type File struct {
	Spec Spec
	Path string
	Data []byte
}

var profiles = []struct {
	make, model, lens string
	fmt               string
	f35               uint16
	apN, apD          uint32
}{
	{"Canon", "EOS R5", "RF 24-70mm F2.8 L IS USM", "CR3", 50, 28, 10},
	{"Canon", "EOS R6 Mark II", "RF 50mm F1.2 L USM", "CR3", 50, 14, 10},
	{"Nikon", "Z 8", "NIKKOR Z 24-70mm f/2.8 S", "NEF", 35, 40, 10},
	{"Nikon", "Z 6II", "NIKKOR Z 85mm f/1.8 S", "NEF", 85, 18, 10},
	{"SONY", "ILCE-7RM5", "FE 24-70mm F2.8 GM II", "ARW", 70, 28, 10},
	{"SONY", "ILCE-7M4", "FE 35mm F1.4 GM", "ARW", 35, 20, 10},
	{"Canon", "EOS 5D Mark IV", "EF 24-105mm f/4L IS II USM", "CR2", 24, 40, 10},
	{"Adobe", "Lightroom DNG", "Sigma 35mm F1.4 DG DN", "DNG", 35, 20, 10},
}

func BuildCatalog() []Spec {
	now := timeutil.Now()
	start := now.AddDate(-1, 0, 0)
	var out []Spec
	for i := 0; i < 288; i++ {
		p := profiles[i%len(profiles)]
		month := start.AddDate(0, i%12, (i*3)%27)
		iso := uint16(100)
		switch i % 5 {
		case 1:
			iso = 200
		case 2:
			iso = 400
		case 3:
			iso = 800
		case 4:
			iso = 1600
		}
		shutD := uint32(200)
		if i%3 == 0 {
			shutD = 80
		}
		if i%7 == 0 {
			shutD = 500
		}
		ext := p.fmt
		name := fmt.Sprintf("IMG_%04d.%s", 1000+i, ext)
		out = append(out, Spec{
			Filename:      name,
			Format:        p.fmt,
			Make:          p.make,
			Model:         p.model,
			Lens:          p.lens,
			ApertureN:     p.apN + uint32((i%3)*4),
			ApertureD:     p.apD,
			ShutterN:      1,
			ShutterD:      shutD,
			ISO:           iso,
			FocalN:        uint32(p.f35),
			FocalD:        1,
			Focal35:       p.f35,
			ExposureBiasN: int32((i%5)-2) * 10,
			WhiteBalance:  uint16(i % 2),
			ShotAt:        month,
			Width:         320,
			Height:        210,
			Deferred:      false,
			Seed:          i + 1,
		})
	}
	def := out[0]
	def.Filename = "DEFERRED_PREVIEW.NEF"
	def.Format = "NEF"
	def.Deferred = true
	def.Make = "Nikon"
	def.Model = "Z 8"
	def.Lens = "NIKKOR Z 24-70mm f/2.8 S"
	out = append(out, def)
	return out
}

func Render(s Spec) ([]byte, error) {
	jw, jh := 160, 105
	if s.Deferred {
		jw, jh = 120, 80
	}
	jpg, err := EncodeScene(jw, jh, s.Seed, float64(s.Seed))
	if err != nil {
		return nil, err
	}
	switch s.Format {
	case "CR3":
		return buildCR3(s, jpg)
	default:
		return buildTIFF(s, jpg)
	}
}

func WriteFile(dir string, s Spec) (File, error) {
	data, err := Render(s)
	if err != nil {
		return File{}, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return File{}, err
	}
	path := filepath.Join(dir, s.Filename)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return File{}, err
	}
	return File{Spec: s, Path: path, Data: data}, nil
}

func buildTIFF(s Spec, jpeg []byte) ([]byte, error) {
	w := newTIFF(true)
	exifEnts := []ifdEnt{
		ratEnt(0x829A, s.ShutterN, s.ShutterD),
		ratEnt(0x829D, s.ApertureN, s.ApertureD),
		shortEnt(0x8827, s.ISO),
		asciiEnt(0x9003, s.ShotAt.In(timeutil.Beijing).Format("2006:01:02 15:04:05")),
		sratEnt(0x9204, s.ExposureBiasN, 100),
		ratEnt(0x920A, s.FocalN, s.FocalD),
		shortEnt(0xA403, s.WhiteBalance),
		shortEnt(0xA405, s.Focal35),
		ratsEnt(0xA432, [][2]uint32{{s.FocalN, s.FocalD}, {s.FocalN, s.FocalD}, {s.ApertureN, s.ApertureD}, {s.ApertureN, s.ApertureD}}),
		asciiEnt(0xA434, s.Lens),
	}
	exifOff := w.writeIFD(exifEnts, 0)

	if s.Deferred {
		w.padTo(20 << 20)
	}
	jpgOff := w.appendJPEG(jpeg)

	ifd0 := []ifdEnt{
		shortEnt(0x0100, uint16(s.Width)),
		shortEnt(0x0101, uint16(s.Height)),
		asciiEnt(0x010F, s.Make),
		asciiEnt(0x0110, s.Model),
		shortEnt(0x0112, 1),
		asciiEnt(0x0132, s.ShotAt.In(timeutil.Beijing).Format("2006:01:02 15:04:05")),
		longEnt(0x8769, exifOff),
		longEnt(0x0201, jpgOff),
		longEnt(0x0202, uint32(len(jpeg))),
	}
	ifd0Off := w.writeIFD(ifd0, 0)
	w.setIFD0(ifd0Off)
	return w.buf, nil
}

func buildCR3(s Spec, jpeg []byte) ([]byte, error) {
	tiff, err := buildTIFF(s, jpeg)
	if err != nil {
		return nil, err
	}
	var out []byte
	out = appendBox(out, "ftyp", append([]byte("crx "), []byte{0, 0, 0, 0}...))
	cmt := append([]byte("CMT1"), tiff...)
	uuidPayload := append(append([]byte{}, canonUUID()...), cmt...)
	out = append(out, boxBytes("moov", uuidBox("uuid", uuidPayload, s.LargeSize))...)
	out = appendBox(out, "PRVW", jpeg)
	return out, nil
}

func appendBox(dst []byte, typ string, payload []byte) []byte {
	sz := uint32(8 + len(payload))
	b := make([]byte, 8)
	binary.BigEndian.PutUint32(b, sz)
	copy(b[4:], []byte(typ))
	return append(append(dst, b...), payload...)
}

func boxBytes(typ string, payload []byte) []byte {
	return appendBox(nil, typ, payload)
}

// uuidBox serialises a uuid box. When largeSize is true, the box uses the
// 64-bit largesize form (size field = 1, followed by an 8-byte extended size)
// as permitted by ISO/IEC 14496-12 §4.2.3. Some archiving systems rewrite
// box sizes this way when the payload crosses the 4 GiB threshold.
func uuidBox(typ string, payload []byte, largeSize bool) []byte {
	if !largeSize {
		return boxBytes(typ, payload)
	}
	total := 16 + len(payload)
	b := make([]byte, 0, total)
	hdr := make([]byte, 8)
	binary.BigEndian.PutUint32(hdr, 1) // largesize indicator
	copy(hdr[4:], []byte(typ))
	b = append(b, hdr...)
	ls := make([]byte, 8)
	binary.BigEndian.PutUint64(ls, uint64(total))
	b = append(b, ls...)
	b = append(b, payload...)
	return b
}

func canonUUID() []byte {
	return []byte{0x85, 0xc0, 0xb6, 0x87, 0x82, 0x0f, 0x11, 0xe0, 0x81, 0x11, 0xf4, 0xce, 0x46, 0x2b, 0x6a, 0x48}
}
