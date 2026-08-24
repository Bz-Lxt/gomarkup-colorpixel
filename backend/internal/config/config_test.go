package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("SAMPLE_MODE", "1")
	c := Load()
	if c.PreviewWindowBytes != 16<<20 {
		t.Fatal(c.PreviewWindowBytes)
	}
	if !c.SampleMode {
		t.Fatal("sample")
	}
}

func TestEnvInt(t *testing.T) {
	os.Setenv("MAX_IFDS", "12")
	defer os.Unsetenv("MAX_IFDS")
	c := Load()
	if c.MaxIFDs != 12 {
		t.Fatal(c.MaxIFDs)
	}
}
