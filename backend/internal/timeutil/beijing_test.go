package timeutil

import (
	"testing"
	"time"
)

func TestParseAndFormat(t *testing.T) {
	tm, err := ParseEXIFDate("2026:03:12 09:14:00")
	if err != nil || tm.Hour() != 9 {
		t.Fatalf("%v %v", tm, err)
	}
	if FormatDisplay(tm) != "2026-03-12 09:14:00" {
		t.Fatal(FormatDisplay(tm))
	}
	if !ToBeijing(time.Time{}).IsZero() {
		t.Fatal("zero")
	}
}

func TestNowZone(t *testing.T) {
	n := Now()
	if n.Location() != Beijing {
		t.Fatal(n.Location())
	}
}
