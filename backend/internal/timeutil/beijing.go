package timeutil

import (
	"time"
)

var Beijing = time.FixedZone("CST", 8*60*60)

func Now() time.Time {
	return time.Now().In(Beijing)
}

func ToBeijing(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.In(Beijing)
}

func ParseEXIFDate(s string) (time.Time, error) {
	s = trimASCII(s)
	if s == "" {
		return time.Time{}, nil
	}
	layouts := []string{
		"2006:01:02 15:04:05",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	var last error
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, s, Beijing)
		if err == nil {
			return t, nil
		}
		last = err
	}
	return time.Time{}, last
}

func FormatDisplay(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return ToBeijing(t).Format("2006-01-02 15:04:05")
}

func trimASCII(s string) string {
	for len(s) > 0 && (s[len(s)-1] == 0 || s[len(s)-1] == ' ') {
		s = s[:len(s)-1]
	}
	return s
}
