package logger

import (
	"bytes"
	"testing"
)

func TestInitLevels(t *testing.T) {
	var buf bytes.Buffer
	Init("debug", &buf)
	L().Debug("d")
	Init("error", &buf)
	L().Info("hidden")
	if L() == nil {
		t.Fatal("nil")
	}
}
