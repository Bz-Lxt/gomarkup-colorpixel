package tiles_test

import (
	"os"
	"testing"

	"colorpixel/internal/sample"
	"colorpixel/internal/tiles"
)

func TestBuildWritesThumbAndLevelZero(t *testing.T) {
	jpg, err := sample.EncodeScene(300, 220, 2, 1.2)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	res, img, err := tiles.Build(jpg, dir)
	if err != nil {
		t.Fatal(err)
	}
	if img == nil || res.Width == 0 {
		t.Fatal("empty")
	}
	if _, err := os.Stat(dir + "/thumb.jpg"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(tiles.Path(dir, 0, 0, 0)); err != nil {
		t.Fatal(err)
	}
}
