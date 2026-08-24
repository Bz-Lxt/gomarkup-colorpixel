package raw

import "testing"

func TestTagNameCoverage(t *testing.T) {
	need := []uint16{0x010F, 0x0110, 0x829A, 0x829D, 0x8827, 0x9003, 0x920A, 0xA434, 0xA405, 0x8769}
	for _, id := range need {
		if TagName(id) == "" {
			t.Fatalf("missing name for 0x%04X", id)
		}
	}
	if TagName(0xFFFF) != "" {
		t.Fatal("unknown tag should be empty")
	}
}

func TestRejectBigTIFF(t *testing.T) {
	ii := []byte{'I', 'I', 43, 0, 8, 0, 0, 0}
	if err := rejectBigTIFF(ii); err == nil {
		t.Fatal("expected error")
	}
	mm := []byte{'M', 'M', 0, 43, 0, 0, 0, 8}
	if err := rejectBigTIFF(mm); err == nil {
		t.Fatal("expected error")
	}
	if err := rejectBigTIFF([]byte{'I', 'I', 42, 0}); err != nil {
		t.Fatal(err)
	}
}

func TestClampPreview(t *testing.T) {
	if _, err := clampPreview(0, 10); err == nil {
		t.Fatal("empty")
	}
	if _, err := clampPreview(20, 10); err == nil {
		t.Fatal("over")
	}
	n, err := clampPreview(8, 10)
	if err != nil || n != 8 {
		t.Fatalf("%d %v", n, err)
	}
}

func TestEnrichTagNames(t *testing.T) {
	res := &Result{Tags: map[string]any{"0x010F": "Canon"}}
	enrichTagNames(res)
	if res.Tags["Make"] != "Canon" {
		t.Fatalf("%v", res.Tags)
	}
}
