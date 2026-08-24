package raw

import "fmt"

func rejectBigTIFF(head []byte) error {
	if len(head) < 8 {
		return nil
	}
	if (head[0] == 'I' && head[1] == 'I' && head[2] == 43 && head[3] == 0) ||
		(head[0] == 'M' && head[1] == 'M' && head[2] == 0 && head[3] == 43) {
		return wrap("tiff", fmt.Errorf("BigTIFF is not supported"))
	}
	return nil
}

func clampPreview(n, max int) (int, error) {
	if n <= 0 {
		return 0, fmt.Errorf("empty preview")
	}
	if n > max {
		return 0, fmt.Errorf("preview %d exceeds %d", n, max)
	}
	return n, nil
}
