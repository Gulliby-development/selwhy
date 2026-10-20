package match

import (
	"testing"

	"github.com/Gulliby-development/selwhy/internal/markup"
	"github.com/Gulliby-development/selwhy/internal/selector"
)

func parseMatchDocument(t *testing.T, text string) *markup.Doc {
	t.Helper()
	doc, err := markup.Parse(text)
	if err != nil {
		t.Fatalf("markup.Parse(%q): %v", text, err)
	}
	return doc
}

func parseMatchSimple(t *testing.T, text string) selector.Simple {
	t.Helper()
	list, err := selector.Parse(text)
	if err != nil {
		t.Fatalf("selector.Parse(%q): %v", text, err)
	}
	if len(list) != 1 || list[0] == nil || len(list[0].Compounds) != 1 || len(list[0].Compounds[0]) != 1 {
		t.Fatalf("selector.Parse(%q) did not produce exactly one simple selector", text)
	}
	return list[0].Compounds[0][0]
}

func TestSimpleMatchesLiteralCases(t *testing.T) {
	doc := parseMatchDocument(t, `<r><A id="x" class="one two" data-v="Ab-c tail" empty="" u="Ä"/></r>`)
	if len(doc.Elements) != 2 {
		t.Fatalf("markup.Parse produced %d elements, want 2", len(doc.Elements))
	}
	element := doc.Elements[1]

	cases := []struct {
		name     string
		selector string
		want     bool
	}{
		{"exact tag", "A", true},
		{"tag case differs", "a", false},
		{"universal selector", "*", true},
		{"exact id", "#x", true},
		{"id case differs", "#X", false},
		{"class token", ".two", true},
		{"class substring is not a token", ".on", false},
		{"present empty attribute", "[empty]", true},
		{"absent attribute", "[missing]", false},
		{"absent attribute with empty equality", `[missing=""]`, false},
		{"empty attribute equality", `[empty=""]`, true},
		{"ASCII-insensitive equality", `[data-v="ab-c tail" i]`, true},
		{"ASCII-sensitive equality", `[data-v="ab-c tail" s]`, false},
		{"default sensitive equality", `[data-v="ab-c tail"]`, false},
		{"token match", `[data-v~="tail"]`, true},
		{"token cannot contain a blank", `[data-v~="c tail"]`, false},
		{"empty token never matches", `[data-v~=""]`, false},
		{"hyphen-prefix equality", `[data-v|="Ab"]`, true},
		{"hyphen-prefix needs whole prefix", `[data-v|="A"]`, false},
		{"prefix match", `[data-v^="Ab"]`, true},
		{"suffix match", `[data-v$="tail"]`, true},
		{"substring match", `[data-v*="-c "]`, true},
		{"empty prefix never matches", `[empty^=""]`, false},
		{"empty suffix never matches", `[empty$=""]`, false},
		{"empty substring never matches", `[empty*=""]`, false},
		{"ASCII-insensitive does not fold non-ASCII", `[u="ä" i]`, false},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			simple := parseMatchSimple(t, test.selector)
			if got := SimpleMatches(simple, element); got != test.want {
				t.Errorf("SimpleMatches(%q) = %t, want %t", test.selector, got, test.want)
			}
		})
	}
}

func TestSimpleMatchesLiteralBlankTokenCases(t *testing.T) {
	doc := parseMatchDocument(t, "<r><A class=\"one\ttwo\rthree\nfour\ffive\" data-v=\"one\ttwo\rthree\nfour\ffive\"/></r>")
	element := doc.Elements[1]

	cases := []struct {
		name     string
		selector string
		want     bool
	}{
		{"class token after tab", ".two", true},
		{"class token after carriage return", ".three", true},
		{"class token after line feed", ".four", true},
		{"class token after form feed", ".five", true},
		{"attribute token after tab", `[data-v~="two"]`, true},
		{"attribute token after carriage return", `[data-v~="three"]`, true},
		{"attribute token after line feed", `[data-v~="four"]`, true},
		{"attribute token after form feed", `[data-v~="five"]`, true},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			simple := parseMatchSimple(t, test.selector)
			if got := SimpleMatches(simple, element); got != test.want {
				t.Errorf("SimpleMatches(%q) = %t, want %t", test.selector, got, test.want)
			}
		})
	}
}

func TestSimpleMatchesLiteralNonBreakingSpaceTokens(t *testing.T) {
	doc := parseMatchDocument(t, "<r><A class=\"one two\" data-v=\"one two\"/></r>")
	element := doc.Elements[1]

	classWhole := selector.Simple{Kind: selector.ClassSel, Name: "one two"}
	if got := SimpleMatches(classWhole, element); !got {
		t.Error("SimpleMatches did not keep a non-breaking space inside one class token")
	}
	classPart := selector.Simple{Kind: selector.ClassSel, Name: "one"}
	if got := SimpleMatches(classPart, element); got {
		t.Error("SimpleMatches split a class token at a non-breaking space")
	}

	for _, test := range []struct {
		selector string
		want     bool
	}{
		{"[data-v~=\"one two\"]", true},
		{"[data-v~=\"one\"]", false},
	} {
		simple := parseMatchSimple(t, test.selector)
		if got := SimpleMatches(simple, element); got != test.want {
			t.Errorf("SimpleMatches(%q) = %t, want %t", test.selector, got, test.want)
		}
	}
}

func TestMatchingFixtureParsersReportLiteralErrors(t *testing.T) {
	cases := []struct {
		name string
		run  func() error
		want string
	}{
		{
			name: "markup fixture",
			run: func() error {
				_, err := markup.Parse("<a>")
				return err
			},
			want: "unclosed element a starting at line 1 column 1",
		},
		{
			name: "selector fixture",
			run: func() error {
				_, err := selector.Parse("a%")
				return err
			},
			want: `unexpected "%" at column 2`,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			if err == nil || err.Error() != test.want {
				t.Errorf("parse error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestPositionLiteralCases(t *testing.T) {
	doc := parseMatchDocument(t, "<r><a/><b/><a> </a><a><!--x--></a></r>")
	cases := []struct {
		name     string
		element  int
		position string
		wantPos  int
		wantAll  int
	}{
		{"root among siblings", 0, "nth-last-child", 1, 1},
		{"mixed-type child position", 3, "nth-child", 3, 4},
		{"same-type position", 3, "nth-of-type", 2, 3},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			gotPos, gotAll := Position(test.position, doc.Elements[test.element])
			if gotPos != test.wantPos || gotAll != test.wantAll {
				t.Errorf("Position(%q, element %d) = (%d, %d), want (%d, %d)",
					test.position, test.element+1, gotPos, gotAll, test.wantPos, test.wantAll)
			}
		})
	}
}

func TestStructuralAndListPseudoLiteralCases(t *testing.T) {
	doc := parseMatchDocument(t, "<r><a/><b/><a> </a><a><!--x--></a></r>")
	cases := []struct {
		name     string
		selector string
		element  int
		want     bool
	}{
		{"root predicate", ":root", 0, true},
		{"first child", ":first-child", 1, true},
		{"last child", ":last-child", 4, true},
		{"not only child", ":only-child", 1, false},
		{"root is only child", ":only-child", 0, true},
		{"first of type", ":first-of-type", 2, true},
		{"last of type", ":last-of-type", 4, true},
		{"unique type", ":only-of-type", 2, true},
		{"forward child formula", ":nth-child(odd)", 3, true},
		{"reverse child formula", ":nth-last-child(2)", 3, true},
		{"filtered child formula", ":nth-of-type(2)", 3, true},
		{"filtered reverse formula", ":nth-last-of-type(2)", 3, true},
		{"blank text is not empty", ":empty", 3, false},
		{"comment does not prevent empty", ":empty", 4, true},
		{"element child prevents empty", ":empty", 0, false},
		{"negation rejects matching argument", ":not(a, b)", 3, false},
		{"is accepts matching argument", ":is(b, a)", 3, true},
		{"where accepts matching argument", ":where(b, a)", 3, true},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			simple := parseMatchSimple(t, test.selector)
			if got := SimpleMatches(simple, doc.Elements[test.element]); got != test.want {
				t.Errorf("SimpleMatches(%q, element %d) = %t, want %t",
					test.selector, test.element+1, got, test.want)
			}
		})
	}
}

func TestAC003NamedElementCases(t *testing.T) {
	doc := parseMatchDocument(t, "<r><a/><b/><a> </a><a><!--x--></a></r>")
	E1, E2, E3, E4, E5 := doc.Elements[0], doc.Elements[1], doc.Elements[2], doc.Elements[3], doc.Elements[4]

	if pos, total := Position("nth-child", E4); pos != 3 || total != 4 {
		t.Errorf("Position(\"nth-child\", E4) = (%d, %d), want (3, 4)", pos, total)
	}
	if pos, total := Position("nth-of-type", E4); pos != 2 || total != 3 {
		t.Errorf("Position(\"nth-of-type\", E4) = (%d, %d), want (2, 3)", pos, total)
	}
	if pos, total := Position("nth-last-child", E1); pos != 1 || total != 1 {
		t.Errorf("Position(\"nth-last-child\", E1) = (%d, %d), want (1, 1)", pos, total)
	}

	cases := []struct {
		selector string
		element  *markup.Element
		want     bool
	}{
		{":root", E1, true},
		{":first-child", E2, true},
		{":last-child", E5, true},
		{":only-child", E2, false},
		{":only-child", E1, true},
		{":first-of-type", E3, true},
		{":last-of-type", E5, true},
		{":only-of-type", E3, true},
		{":nth-child(odd)", E4, true},
		{":nth-last-child(2)", E4, true},
		{":nth-of-type(2)", E4, true},
		{":nth-last-of-type(2)", E4, true},
		{":empty", E4, false},
		{":empty", E5, true},
		{":empty", E1, false},
		{":not(a, b)", E4, false},
		{":is(b, a)", E4, true},
		{":where(b, a)", E4, true},
	}

	for _, test := range cases {
		simple := parseMatchSimple(t, test.selector)
		if got := SimpleMatches(simple, test.element); got != test.want {
			t.Errorf("SimpleMatches(%q, element %d) = %t, want %t", test.selector, test.element.Num, got, test.want)
		}
	}
}

func TestComplexMatchesLiteralRelationCases(t *testing.T) {
	doc := parseMatchDocument(t, "<r><a><b><b><x/></b></b></a><c/><d/><d/><z/></r>")
	cases := []struct {
		name     string
		selector string
		element  int
		want     bool
	}{
		{"child relation", "b > x", 4, true},
		{"wrong parent", "a > x", 4, false},
		{"ancestor prefix backtracking", "a > b x", 4, true},
		{"adjacent sibling relation", "c + d", 6, true},
		{"adjacent relation is not a later sibling", "c + d", 7, false},
		{"earlier sibling relation", "c ~ d", 7, true},
		{"sibling prefix backtracking", "c + d ~ z", 8, true},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			list, err := selector.Parse(test.selector)
			if err != nil {
				t.Fatalf("selector.Parse(%q): %v", test.selector, err)
			}
			if got := ComplexMatches(list[0], doc.Elements[test.element]); got != test.want {
				t.Errorf("ComplexMatches(%q, element %d) = %t, want %t",
					test.selector, test.element+1, got, test.want)
			}
		})
	}
}

func TestSimpleMatchesLiteralHasAnchorCases(t *testing.T) {
	doc := parseMatchDocument(t, "<r><a><b><b><x/></b></b></a><c/><d/><d/><z/></r>")

	cases := []struct {
		name     string
		selector string
		element  int
		want     bool
	}{
		{"child-led anchored chain", ":has(> b > b > x)", 1, true},
		{"child lead cannot skip anchor child", ":has(> b > x)", 1, false},
		{"descendant-led chain", ":has(b x)", 1, true},
		{"anchor is not its own descendant", ":has(a x)", 1, false},
		{"next-sibling-led chain", ":has(+ c + d)", 1, true},
		{"later sibling is not next", ":has(+ d)", 1, false},
		{"later-sibling-led chain", ":has(~ d + z)", 1, true},
		{"a sibling is not a child", ":has(> d)", 5, false},
		{"relative selector alternatives", ":has(> q, + c)", 1, true},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			simple := parseMatchSimple(t, test.selector)
			if got := SimpleMatches(simple, doc.Elements[test.element]); got != test.want {
				t.Errorf("SimpleMatches(%q, element %d) = %t, want %t",
					test.selector, test.element+1, got, test.want)
			}
		})
	}
}

func TestListMatchesAndSelectLiteralUnionCases(t *testing.T) {
	doc := parseMatchDocument(t, "<r><a/><b/><a/></r>")
	cases := []struct {
		name      string
		selectors string
		element   int
		want      bool
	}{
		{"selector list union", "b, a", 2, true},
		{"no selector in list matches", "b, z", 1, false},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			list, err := selector.Parse(test.selectors)
			if err != nil {
				t.Fatalf("selector.Parse(%q): %v", test.selectors, err)
			}
			if got := ListMatches(list, doc.Elements[test.element]); got != test.want {
				t.Errorf("ListMatches(%q, element %d) = %t, want %t",
					test.selectors, test.element+1, got, test.want)
			}
		})
	}

	if got := ListMatches(selector.List{}, doc.Elements[1]); got {
		t.Error("ListMatches(empty, element 2) = true, want false")
	}

	selectCases := []struct {
		name      string
		selectors string
		want      []int
	}{
		{"reordered repeated overlapping union", "b, a, a, *", []int{1, 2, 3, 4}},
		{"filtered document order", "a", []int{2, 4}},
		{"absent selector", "z", []int{}},
	}
	for _, test := range selectCases {
		t.Run(test.name, func(t *testing.T) {
			list, err := selector.Parse(test.selectors)
			if err != nil {
				t.Fatalf("selector.Parse(%q): %v", test.selectors, err)
			}
			for run := 0; run < 2; run++ {
				selected := Select(list, doc)
				got := make([]int, 0, len(selected))
				for _, element := range selected {
					got = append(got, element.Num)
				}
				if len(got) != len(test.want) {
					t.Errorf("Select(%q) run %d returned element numbers %v, want %v",
						test.selectors, run+1, got, test.want)
					continue
				}
				for index, number := range got {
					if number != test.want[index] {
						t.Errorf("Select(%q) run %d returned element numbers %v, want %v",
							test.selectors, run+1, got, test.want)
						break
					}
				}
			}
		})
	}

	if got := Select(selector.List{}, doc); len(got) != 0 {
		t.Errorf("Select(empty) returned %d elements, want none", len(got))
	}
}

func TestAC004NamedElementCases(t *testing.T) {
	doc := parseMatchDocument(t, "<r><a><b><b><x/></b></b></a><c/><d/><d/><z/></r>")
	E6, E7, E8, E9 := doc.Elements[5], doc.Elements[6], doc.Elements[7], doc.Elements[8]

	cases := []struct {
		name     string
		selector string
		element  *markup.Element
		want     bool
	}{
		{"c + d at E7", "c + d", E7, true},
		{"c + d at E8", "c + d", E8, false},
		{"c + d ~ z at E9", "c + d ~ z", E9, true},
	}

	for _, test := range cases {
		list, err := selector.Parse(test.selector)
		if err != nil {
			t.Fatalf("selector.Parse(%q): %v", test.selector, err)
		}
		if got := ComplexMatches(list[0], test.element); got != test.want {
			t.Errorf("ComplexMatches(%q, element %d) = %t, want %t",
				test.selector, test.element.Num, got, test.want)
		}
	}

	if got := SimpleMatches(parseMatchSimple(t, ":has(> d)"), E6); got {
		t.Errorf(`SimpleMatches(":has(> d)", E6) = true, want false`)
	}
}

func TestAC005NamedElementUnionCases(t *testing.T) {
	doc := parseMatchDocument(t, "<r><a/><b/><a/></r>")
	E2 := doc.Elements[1]

	// AC-005: ListMatches(Parse("b, a"), E2) and ListMatches(Parse("b, z"), E2).
	listBA, err := selector.Parse("b, a")
	if err != nil {
		t.Fatalf("selector.Parse(%q): %v", "b, a", err)
	}
	if got := ListMatches(listBA, E2); !got {
		t.Error(`ListMatches(Parse("b, a"), E2) = false, want true`)
	}

	listBZ, err := selector.Parse("b, z")
	if err != nil {
		t.Fatalf("selector.Parse(%q): %v", "b, z", err)
	}
	if got := ListMatches(listBZ, E2); got {
		t.Error(`ListMatches(Parse("b, z"), E2) = true, want false`)
	}

	// AC-005: Select(Parse("b, a, a, *"), doc) has element numbers [1, 2, 3, 4].
	selectors, err := selector.Parse("b, a, a, *")
	if err != nil {
		t.Fatalf("selector.Parse(%q): %v", "b, a, a, *", err)
	}
	selected := Select(selectors, doc)
	got := make([]int, 0, len(selected))
	for _, element := range selected {
		got = append(got, element.Num)
	}
	want := []int{1, 2, 3, 4}
	if len(got) != len(want) {
		t.Fatalf(`Select(Parse("b, a, a, *"), doc) returned %v, want %v`, got, want)
	}
	for index, number := range got {
		if number != want[index] {
			t.Errorf(`Select(Parse("b, a, a, *"), doc) returned %v, want %v`, got, want)
			break
		}
	}
}

func TestReasonLiteralFailureCases(t *testing.T) {
	doc := parseMatchDocument(t, `<r><a id="x" class="one two" v="é">t<b/></a><a/></r>`)
	cases := []struct {
		name     string
		selector string
		element  int
		want     string
	}{
		{"tag", "z", 1, `fails [z] (the tag is "a")`},
		{"existing id", "#y", 1, `fails [#y] (the id is "x")`},
		{"missing id", "#y", 3, `fails [#y] (there is no id)`},
		{"existing classes", ".z", 1, `fails [.z] (the classes are [one two])`},
		{"missing classes", ".z", 3, `fails [.z] (the classes are [])`},
		{"attribute replacement", `[v="x"]`, 1, `fails [[v="x"]] (the value is "?")`},
		{"missing attribute", `[v]`, 3, `fails [[v]] (there is no such attribute)`},
		{"empty children", ":empty", 1, `fails [:empty] (the element has 2 children)`},
		{"root", ":root", 1, `fails [:root] (the element is not the root)`},
		{"child position", ":first-child", 3, `fails [:first-child] (the position is 2 of 2)`},
		{"nth-last reports forward position", ":nth-last-child(2)", 3, `fails [:nth-last-child(2)] (the position is 2 of 2)`},
		{"of-type total", ":only-of-type", 1, `fails [:only-of-type] (the position is 1 of 2)`},
		{"first matching not argument", ":not(a, #x)", 1, `fails [:not(a, #x)] (the element matches [a])`},
		{"is with no match", ":is(z, q)", 1, `fails [:is(z, q)] (the element matches none of the arguments)`},
		{"where with no match", ":where(z)", 1, `fails [:where(z)] (the element matches none of the arguments)`},
		{"has with no relative match", ":has(> z)", 1, `fails [:has(> z)] (no relative selector matches)`},
		{"first failing part", "z#y", 1, `fails [z] (the tag is "a")`},
		{"last compound part before relation", "q > a.z", 1, `fails [.z] (the classes are [one two])`},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			list, err := selector.Parse(test.selector)
			if err != nil {
				t.Fatalf("selector.Parse(%q): %v", test.selector, err)
			}
			got := Reason(list[0], doc.Elements[test.element])
			if got != test.want {
				t.Errorf("Reason(%q) = %q, want %q", test.selector, got, test.want)
			}
		})
	}
}

func TestReasonLiteralRelationFailureCases(t *testing.T) {
	doc := parseMatchDocument(t, `<r><a/><b/></r>`)
	cases := []struct {
		name     string
		selector string
		element  int
		want     string
	}{
		{"parent", "z > b", 2, `no parent matches [z]`},
		{"ancestor", "z b", 2, `no ancestor matches [z]`},
		{"previous sibling", "z + b", 2, `no previous sibling matches [z]`},
		{"earlier sibling", "z ~ b", 2, `no earlier sibling matches [z]`},
		{"full prefix", "r > z + b", 2, `no previous sibling matches [r > z]`},
		{"absent previous sibling", "a + a", 1, `no previous sibling matches [a]`},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			list, err := selector.Parse(test.selector)
			if err != nil {
				t.Fatalf("selector.Parse(%q): %v", test.selector, err)
			}
			got := Reason(list[0], doc.Elements[test.element])
			if got != test.want {
				t.Errorf("Reason(%q) = %q, want %q", test.selector, got, test.want)
			}
		})
	}
}
