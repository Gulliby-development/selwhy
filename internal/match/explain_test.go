package match

import (
	"testing"

	"github.com/Gulliby-development/selwhy/internal/markup"
	"github.com/Gulliby-development/selwhy/internal/selector"
)

func parseExplanationDocument(t *testing.T, text string) *markup.Doc {
	t.Helper()
	doc, err := markup.Parse(text)
	if err != nil {
		t.Fatalf("markup.Parse(%q): %v", text, err)
	}
	return doc
}

func explanationFor(t *testing.T, text string, e *markup.Element) string {
	t.Helper()
	list, err := selector.Parse(text)
	if err != nil {
		t.Fatalf("selector.Parse(%q): %v", text, err)
	}
	if len(list) != 1 {
		t.Fatalf("selector.Parse(%q) returned %d selectors, want 1", text, len(list))
	}
	return Reason(list[0], e)
}

func TestReasonAC006LiteralCases(t *testing.T) {
	doc := parseExplanationDocument(t, `<r><a id="x" class="one two" v="é">t<b/></a><a/></r>`)
	cases := []struct {
		name     string
		selector string
		element  int
		want     string
	}{
		{"E2 type", "z", 2, `fails [z] (the tag is "a")`},
		{"E2 existing id", "#y", 2, `fails [#y] (the id is "x")`},
		{"E4 missing id", "#y", 4, `fails [#y] (there is no id)`},
		{"E2 existing classes", ".z", 2, `fails [.z] (the classes are [one two])`},
		{"E4 missing classes", ".z", 4, `fails [.z] (the classes are [])`},
		{"E2 ASCII replacement", `[v="x"]`, 2, `fails [[v="x"]] (the value is "?")`},
		{"E4 missing attribute", "[v]", 4, `fails [[v]] (there is no such attribute)`},
		{"E2 element and text children", ":empty", 2, `fails [:empty] (the element has 2 children)`},
		{"E2 nonroot", ":root", 2, `fails [:root] (the element is not the root)`},
		{"E4 child position", ":first-child", 4, `fails [:first-child] (the position is 2 of 2)`},
		{"E4 nth-last uses forward position", ":nth-last-child(2)", 4, `fails [:nth-last-child(2)] (the position is 2 of 2)`},
		{"E2 of-type total", ":only-of-type", 2, `fails [:only-of-type] (the position is 1 of 2)`},
		{"E2 first matching not argument", ":not(a, #x)", 2, `fails [:not(a, #x)] (the element matches [a])`},
		{"E2 is without matching argument", ":is(z, q)", 2, `fails [:is(z, q)] (the element matches none of the arguments)`},
		{"E2 where without matching argument", ":where(z)", 2, `fails [:where(z)] (the element matches none of the arguments)`},
		{"E2 has without relative match", ":has(> z)", 2, `fails [:has(> z)] (no relative selector matches)`},
		{"E2 first failing part", "z#y", 2, `fails [z] (the tag is "a")`},
		{"E2 last compound part before relation", "q > a.z", 2, `fails [.z] (the classes are [one two])`},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if test.element < 1 || test.element > len(doc.Elements) {
				t.Fatalf("fixture has %d elements, cannot select element %d", len(doc.Elements), test.element)
			}
			got := explanationFor(t, test.selector, doc.Elements[test.element-1])
			if got != test.want {
				t.Errorf("Reason(%q) = %q, want %q", test.selector, got, test.want)
			}
		})
	}
}

func TestReasonAC007LiteralRelationCases(t *testing.T) {
	doc := parseExplanationDocument(t, `<r><a/><b/></r>`)
	cases := []struct {
		name     string
		selector string
		element  int
		want     string
	}{
		{"E3 parent", "z > b", 3, "no parent matches [z]"},
		{"E3 ancestor", "z b", 3, "no ancestor matches [z]"},
		{"E3 previous sibling", "z + b", 3, "no previous sibling matches [z]"},
		{"E3 earlier sibling", "z ~ b", 3, "no earlier sibling matches [z]"},
		{"E3 complete prefix", "r > z + b", 3, "no previous sibling matches [r > z]"},
		{"E2 absent previous sibling", "a + a", 2, "no previous sibling matches [a]"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := explanationFor(t, test.selector, doc.Elements[test.element-1])
			if got != test.want {
				t.Errorf("Reason(%q) = %q, want %q", test.selector, got, test.want)
			}
		})
	}
}

func TestReasonForMatchingSelectorIsEmpty(t *testing.T) {
	doc := parseExplanationDocument(t, `<r><a id="x"/></r>`)
	if got := explanationFor(t, "#x", doc.Elements[1]); got != "" {
		t.Errorf("Reason(#x) = %q, want empty string", got)
	}
}

func TestReasonIsRepeatable(t *testing.T) {
	doc := parseExplanationDocument(t, `<r><a id="x" class="one two" v="é">t<b/></a><a/></r>`)
	const selectorText = `[v="x"]`
	const want = `fails [[v="x"]] (the value is "?")`
	element := doc.Elements[1]

	first := explanationFor(t, selectorText, element)
	second := explanationFor(t, selectorText, element)
	if first != want || second != want {
		t.Errorf("Reason(%q) results = (%q, %q), want (%q, %q)",
			selectorText, first, second, want, want)
	}
}
