package markup

import (
	"reflect"
	"testing"
)

func mustParse(t *testing.T, input string) *Doc {
	t.Helper()
	doc, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", input, err)
	}
	return doc
}

func wantParseError(t *testing.T, input, want string) {
	t.Helper()
	doc, err := Parse(input)
	if err == nil {
		t.Fatalf("Parse(%q) returned document %#v, want error %q", input, doc, want)
	}
	if got := err.Error(); got != want {
		t.Errorf("Parse(%q) error = %q, want %q", input, got, want)
	}
}

func TestParseDocumentOrderRelationshipsAndPaths(t *testing.T) {
	doc := mustParse(t, "<r><a/><b/><a><a/></a></r>")

	wantNames := []string{"r", "a", "b", "a", "a"}
	wantNumbers := []int{1, 2, 3, 4, 5}
	if len(doc.Elements) != len(wantNames) {
		t.Fatalf("len(Elements) = %d, want %d", len(doc.Elements), len(wantNames))
	}
	for i, element := range doc.Elements {
		if element.Name != wantNames[i] {
			t.Errorf("Elements[%d].Name = %q, want %q", i, element.Name, wantNames[i])
		}
		if element.Num != wantNumbers[i] {
			t.Errorf("Elements[%d].Num = %d, want %d", i, element.Num, wantNumbers[i])
		}
	}
	root := doc.Root
	if root != doc.Elements[0] {
		t.Fatalf("Root is not Elements[0]")
	}
	if root.Parent != nil {
		t.Errorf("root Parent = %p, want nil", root.Parent)
	}
	if len(root.Children) != 3 || root.Children[0] != doc.Elements[1] || root.Children[1] != doc.Elements[2] || root.Children[2] != doc.Elements[3] {
		t.Errorf("root Children = %#v, want Elements[1], Elements[2], Elements[3]", root.Children)
	}
	if doc.Elements[1].Parent != root || doc.Elements[2].Parent != root || doc.Elements[3].Parent != root {
		t.Errorf("root child Parent pointers do not reference Root")
	}
	if len(doc.Elements[3].Children) != 1 || doc.Elements[3].Children[0] != doc.Elements[4] || doc.Elements[4].Parent != doc.Elements[3] {
		t.Errorf("nested parent/child pointers do not reference Elements[3] and Elements[4]")
	}

	wantPaths := []string{"/r[1]", "/r[1]/a[1]", "/r[1]/b[1]", "/r[1]/a[2]", "/r[1]/a[2]/a[1]"}
	for i, element := range doc.Elements {
		if got := Path(element); got != wantPaths[i] {
			t.Errorf("Path(Elements[%d]) = %q, want %q", i, got, wantPaths[i])
		}
	}
	if got := root.Siblings(); !reflect.DeepEqual(got, []*Element{root}) {
		t.Errorf("root Siblings() = %#v, want the root alone", got)
	}
	if got := doc.Elements[1].Siblings(); !reflect.DeepEqual(got, []*Element{doc.Elements[1], doc.Elements[2], doc.Elements[3]}) {
		t.Errorf("child Siblings() = %#v, want all root children in order", got)
	}
}

func TestAttributesAndTagTextPreserveSourceOrder(t *testing.T) {
	doc := mustParse(t, "<A-2 _x y = 'v' z=\"\"/>")
	root := doc.Root
	wantAttrs := []Attr{{Name: "_x", Value: ""}, {Name: "y", Value: "v"}, {Name: "z", Value: ""}}
	if !reflect.DeepEqual(root.Attrs, wantAttrs) {
		t.Errorf("Attrs = %#v, want %#v", root.Attrs, wantAttrs)
	}
	if got, ok := root.Attr("y"); got != "v" || !ok {
		t.Errorf("Attr(y) = (%q, %t), want (%q, true)", got, ok, "v")
	}
	if got, ok := root.Attr("z"); got != "" || !ok {
		t.Errorf("Attr(z) = (%q, %t), want empty and present", got, ok)
	}
	if got, ok := root.Attr("Y"); got != "" || ok {
		t.Errorf("Attr(Y) = (%q, %t), want empty and absent", got, ok)
	}
	if got := TagText(root); got != "<A-2 _x=\"\" y=\"v\" z=\"\">" {
		t.Errorf("TagText(root) = %q, want source-order attributes", got)
	}
}

func TestClassesSplitOnlyOnTheFiveMarkupBlanks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "all five blanks and non-breaking space",
			input: "<r class=\" a\tb\rc\nd\fe  a b \"/>",
			want:  []string{"a", "b", "c", "d", "e", "a b"},
		},
		{name: "no class attribute", input: "<r/>", want: nil},
		{name: "empty class", input: "<r class=\"\"/>", want: nil},
		{name: "only blanks", input: "<r class=\" \t\r\n\f \"/>", want: nil},
		{name: "repeated spaces", input: "<r class=\"alpha  beta\"/>", want: []string{"alpha", "beta"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doc := mustParse(t, test.input)
			got := doc.Root.Classes()
			if len(got) != len(test.want) {
				t.Fatalf("Classes() = %#v, want %#v", got, test.want)
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Errorf("Classes()[%d] = %q, want %q", i, got[i], test.want[i])
				}
			}
		})
	}
}

func TestParseTextRunsAndAttributeValueForms(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		rootTexts  int
		childTexts int
	}{
		{name: "runs around comments and child", input: "<r> <!--c-->x<a>y</a>z</r>", rootTexts: 3, childTexts: 1},
		{name: "empty self-closing element", input: "<r/>", rootTexts: 0},
		{name: "surrounding blanks and comments", input: " \t\r\n\f<!--before--><r><!--inside--></r><!--after--> ", rootTexts: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doc := mustParse(t, test.input)
			if got := doc.Root.Texts; got != test.rootTexts {
				t.Errorf("root Texts = %d, want %d", got, test.rootTexts)
			}
			if test.childTexts != 0 {
				if len(doc.Root.Children) != 1 {
					t.Fatalf("root has %d children, want 1", len(doc.Root.Children))
				}
				if got := doc.Root.Children[0].Texts; got != test.childTexts {
					t.Errorf("child Texts = %d, want %d", got, test.childTexts)
				}
			}
		})
	}

	doc := mustParse(t, "<r a='x\ny' b=word c />")
	wantAttrs := []Attr{{Name: "a", Value: "x\ny"}, {Name: "b", Value: "word"}, {Name: "c", Value: ""}}
	if !reflect.DeepEqual(doc.Root.Attrs, wantAttrs) {
		t.Errorf("quoted, unquoted, and boolean attributes = %#v, want %#v", doc.Root.Attrs, wantAttrs)
	}
}

func TestParseUnquotedSlashBelongsToValue(t *testing.T) {
	wantParseError(t, "<r a=v/>", "unclosed element r starting at line 1 column 1")
}

func TestParseNamesAndAttributeSpacing(t *testing.T) {
	tests := []struct {
		input     string
		name      string
		attribute Attr
	}{
		{input: "<_x/>", name: "_x"},
		{input: "<A-2/>", name: "A-2"},
		{input: "<r a/>", name: "r", attribute: Attr{Name: "a", Value: ""}},
		{input: "<r a = 'v'/>", name: "r", attribute: Attr{Name: "a", Value: "v"}},
	}
	for _, test := range tests {
		doc := mustParse(t, test.input)
		if doc.Root.Name != test.name {
			t.Errorf("Parse(%q) root name = %q, want %q", test.input, doc.Root.Name, test.name)
		}
		if test.attribute.Name != "" && !reflect.DeepEqual(doc.Root.Attrs, []Attr{test.attribute}) {
			t.Errorf("Parse(%q) attributes = %#v, want %#v", test.input, doc.Root.Attrs, []Attr{test.attribute})
		}
	}
}

func TestParseAndDecodeValidEntities(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "named entities", input: "<r a=\"&amp;&lt;&gt;&quot;&apos;\"/>", want: "&<>\"'"},
		{name: "decimal and hexadecimal leading zeroes", input: "<r a=\"&#65;&#0000065;&#x41;&#x000041;\"/>", want: "AAAA"},
		{name: "valid scalar boundaries", input: "<r a=\"&#1;&#xD7FF;&#xE000;&#x10FFFF;\"/>", want: "\x01\uD7FF\uE000\U0010FFFF"},
		{name: "entities in text", input: "<r>&amp;&lt;&gt;&quot;&apos;&#65;&#x41;</r>", want: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doc := mustParse(t, test.input)
			if test.name == "entities in text" {
				if doc.Root.Texts != 1 || len(doc.Root.Children) != 0 {
					t.Errorf("text entity run has Texts=%d and %d children, want 1 and 0", doc.Root.Texts, len(doc.Root.Children))
				}
				return
			}
			if got, ok := doc.Root.Attr("a"); got != test.want || !ok {
				t.Errorf("Attr(a) = (%q, %t), want (%q, true)", got, ok, test.want)
			}
		})
	}

	doc := mustParse(t, "<r a=\"&amp;&lt;&gt;&quot;&apos;\"/>")
	if got := TagText(doc.Root); got != "<r a=\"&<>\"'\">" {
		t.Errorf("TagText(decoded value) = %q, want literal decoded value", got)
	}
}

func TestParseRejectsInvalidEntitiesAtAmpersand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "unknown name", input: "<r>&bogus;</r>", want: "unknown entity at line 1 column 4"},
		{name: "missing semicolon", input: "<r>&amp</r>", want: "unknown entity at line 1 column 4"},
		{name: "empty decimal", input: "<r>&#;</r>", want: "unknown entity at line 1 column 4"},
		{name: "empty hexadecimal", input: "<r>&#x;</r>", want: "unknown entity at line 1 column 4"},
		{name: "uppercase hexadecimal introducer", input: "<r>&#X41;</r>", want: "unknown entity at line 1 column 4"},
		{name: "too many decimal digits", input: "<r>&#00000065;</r>", want: "unknown entity at line 1 column 4"},
		{name: "too many hexadecimal digits", input: "<r>&#x0000041;</r>", want: "unknown entity at line 1 column 4"},
		{name: "zero decimal scalar", input: "<r>&#0;</r>", want: "unknown entity at line 1 column 4"},
		{name: "high surrogate", input: "<r>&#xD800;</r>", want: "unknown entity at line 1 column 4"},
		{name: "low surrogate", input: "<r>&#xDFFF;</r>", want: "unknown entity at line 1 column 4"},
		{name: "decimal above Unicode", input: "<r>&#1114112;</r>", want: "unknown entity at line 1 column 4"},
		{name: "hexadecimal above Unicode", input: "<r>&#x110000;</r>", want: "unknown entity at line 1 column 4"},
		{name: "invalid entity in attribute", input: "<r a=\"&bad;\"/>", want: "unknown entity at line 1 column 7"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wantParseError(t, test.input, test.want)
		})
	}
}

func TestParseRootAndCommentDiagnostics(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty input", input: "", want: "no root element"},
		{name: "comment-only input", input: " <!--ok--> ", want: "no root element"},
		{name: "outside text", input: "a", want: "text outside the root element at line 1 column 1"},
		{name: "trailing text", input: "<r/>x", want: "text outside the root element at line 1 column 5"},
		{name: "unterminated leading comment", input: "<!--x", want: "unterminated comment starting at line 1 column 1"},
		{name: "declaration before root", input: "<!x>", want: "unsupported markup at line 1 column 1"},
		{name: "processing instruction before root", input: "<?x>", want: "unsupported markup at line 1 column 1"},
		{name: "second element root", input: "<r/><s/>", want: "more than one root element at line 1 column 5"},
		{name: "unsupported markup after root", input: "<r/><!x>", want: "more than one root element at line 1 column 5"},
		{name: "bare less-than after root", input: "<r/><", want: "more than one root element at line 1 column 5"},
		{name: "unexpected trailing end tag", input: "<r/></r>", want: "unexpected closing tag at line 1 column 5"},
		{name: "unterminated trailing comment", input: "<r/><!--x", want: "unterminated comment starting at line 1 column 5"},
		{name: "outside text precedes entity validation", input: "&bad;", want: "text outside the root element at line 1 column 1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wantParseError(t, test.input, test.want)
		})
	}
}

func TestParseMalformedAndIncompleteTagDiagnostics(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "bare opening bracket", input: "<", want: "malformed tag at line 1 column 1"},
		{name: "digit cannot start a name", input: "<1>", want: "malformed tag at line 1 column 1"},
		{name: "non-ASCII name", input: "<é/>", want: "malformed tag at line 1 column 1"},
		{name: "bare closing introducer", input: "</", want: "malformed tag at line 1 column 1"},
		{name: "slash not followed by greater-than", input: "<r /x>", want: "malformed tag at line 1 column 1"},
		{name: "empty unquoted value", input: "<r a=>", want: "malformed tag at line 1 column 1"},
		{name: "missing attribute separator", input: "<r a=\"x\"b=\"y\">", want: "malformed tag at line 1 column 1"},
		{name: "start tag ends after name", input: "<r", want: "unterminated tag starting at line 1 column 1"},
		{name: "start tag ends after attribute name", input: "<r a", want: "unterminated tag starting at line 1 column 1"},
		{name: "start tag ends after equals", input: "<r a=", want: "unterminated tag starting at line 1 column 1"},
		{name: "duplicate attribute at second name", input: "<a x=\"1\" x=\"2\"></a>", want: "duplicate attribute x at line 1 column 10"},
		{name: "unterminated quoted value at quote", input: "<a b=\"c></a>", want: "unterminated attribute value starting at line 1 column 6"},
		{name: "mismatched end tag", input: "<a><b></a>", want: "closing tag </a> does not match <b> at line 1 column 7"},
		{name: "end tag without open element", input: "</a>", want: "unexpected closing tag at line 1 column 1"},
		{name: "element names are case-sensitive", input: "<A></a>", want: "closing tag </a> does not match <A> at line 1 column 4"},
		{name: "innermost element remains open", input: "<a>\n  <b>", want: "unclosed element b starting at line 2 column 3"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wantParseError(t, test.input, test.want)
		})
	}
}

func TestParseDepthBoundaryDiagnostics(t *testing.T) {
	doc := mustParse(t, "<a><a><a><a><a><a><a><a><a><a><a><a></a></a></a></a></a></a></a></a></a></a></a></a>")
	if len(doc.Elements) != 12 {
		t.Fatalf("len(Elements) = %d, want 12", len(doc.Elements))
	}

	wantParseError(t, "<a><a><a><a><a><a><a><a><a><a><a><a><a>", "nesting is deeper than 12 at line 1 column 37")
}

func TestParseUnicodeAndLineEndingLocations(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "multibyte rune column", input: "<r>é&bad;</r>", want: "unknown entity at line 1 column 5"},
		{name: "carriage return does not start a line", input: "<r>\r&bad;</r>", want: "unknown entity at line 1 column 5"},
		{name: "line feed starts a line after CRLF", input: "<r>\r\n&bad;</r>", want: "unknown entity at line 2 column 1"},
		{name: "duplicate after multibyte rune", input: "<r a=\"é\" a=\"2\"/>", want: "duplicate attribute a at line 1 column 10"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wantParseError(t, test.input, test.want)
		})
	}
}
