package main

import (
	"fmt"
	"os"

	"colorpixel/internal/sample"
)

func main() {
	dir := "./samples"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	for _, s := range sample.BuildCatalog() {
		f, err := sample.WriteFile(dir, s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", s.Filename, err)
			os.Exit(1)
		}
		fmt.Println(f.Path)
	}
}
