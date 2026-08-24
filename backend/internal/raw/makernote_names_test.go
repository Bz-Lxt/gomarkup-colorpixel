package raw

import "testing"

func TestMakerTagNames(t *testing.T) {
	if canonTagName(0x0095) != "CanonLensModel" {
		t.Fatal(canonTagName(0x0095))
	}
	if nikonTagName(0x0084) != "NikonLens" {
		t.Fatal(nikonTagName(0x0084))
	}
	if sonyTagName(0x2001) != "SonyPreviewImage" {
		t.Fatal(sonyTagName(0x2001))
	}
	if canonTagName(0xFEFE) != "" {
		t.Fatal("unknown")
	}
}
