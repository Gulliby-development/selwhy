package markup

import "testing"

func TestDepthCountsTheElementsFromTheRoot(t *testing.T) {
	doc := mustParse(t, "<r><a><b/></a><c/></r>")

	want := []int{1, 2, 3, 2}
	if len(doc.Elements) != len(want) {
		t.Fatalf("len(Elements) = %d, want %d", len(doc.Elements), len(want))
	}
	for i, element := range doc.Elements {
		if got := Depth(element); got != want[i] {
			t.Errorf("Depth(%s) = %d, want %d", Path(element), got, want[i])
		}
	}
}

func TestDepthOfNilIsZero(t *testing.T) {
	if got := Depth(nil); got != 0 {
		t.Errorf("Depth(nil) = %d, want 0", got)
	}
}
