package selector

import (
	"strconv"
	"testing"
)

func TestPublicASTDeclarations(t *testing.T) {
	kinds := [...]Kind{TypeSel, AnySel, IDSel, ClassSel, AttrSel, PseudoSel}
	for i := 0; i < len(kinds); i++ {
		for j := i + 1; j < len(kinds); j++ {
			if kinds[i] == kinds[j] {
				t.Fatalf("Kind constants %d and %d are not distinct", i, j)
			}
		}
	}

	complex := Complex{
		Lead: ">",
		Compounds: []Compound{
			{Simple{
				Kind: PseudoSel, Name: "nth-child", Op: "=", Value: "x",
				Flag: "i", Args: List{}, A: 2, B: 1,
			}},
			{Simple{Kind: AnySel}},
		},
		Combs: []string{">"},
	}
	list := List{&complex}
	if complex.Lead != ">" || complex.Combs[0] != ">" ||
		complex.Compounds[0][0].Name != "nth-child" ||
		complex.Compounds[0][0].Value != "x" ||
		complex.Compounds[0][0].Flag != "i" ||
		complex.Compounds[0][0].A != 2 ||
		complex.Compounds[0][0].B != 1 ||
		len(list) != 1 || list[0] != &complex {
		t.Fatal("public selector AST fields did not retain their literal values")
	}
}

func TestParseNthAcceptedLiterals(t *testing.T) {
	cases := []struct {
		input string
		a     int
		b     int
	}{
		{"odd", 2, 1},
		{"even", 2, 0},
		{"n", 1, 0},
		{"+n", 1, 0},
		{"-n+3", -1, 3},
		{"2n - 1", 2, -1},
		{"0n+2", 0, 2},
		{"-12", 0, -12},
		{"+000002", 0, 2},
		{" \t\r\n\fodd\f\n\r\t ", 2, 1},
		{"999999n-999999", 999999, -999999},
	}

	for _, test := range cases {
		a, b, ok := ParseNth(test.input)
		if !ok || a != test.a || b != test.b {
			t.Errorf("ParseNth(%q) = (%d, %d, %t), want (%d, %d, true)",
				test.input, a, b, ok, test.a, test.b)
		}
	}
}

func TestParseNthRejectedLiterals(t *testing.T) {
	cases := []string{
		"",
		"2 n",
		"2n+",
		"ODD",
		"N",
		"1.5n",
		"1000000",
		"1000000n",
		"n+1000000",
		"n 1",
		"n+-1",
		"\vodd",
		"\u00a0odd",
	}

	for _, input := range cases {
		_, _, ok := ParseNth(input)
		if ok {
			t.Errorf("ParseNth(%q) accepted an invalid formula", input)
		}
	}
}

func TestNthMatchesLiteralCases(t *testing.T) {
	cases := []struct {
		a, b, pos int
		want      bool
	}{
		{2, 1, 3, true},
		{2, 1, 2, false},
		{2, 5, 3, false},
		{-1, 3, 1, true},
		{-1, 3, 4, false},
		{0, 2, 2, true},
		{0, 2, 3, false},
		{1, 0, 0, true},
		{-2, -1, -3, true},
		{1, -2147483648, 2147483647, true},
		{-1, 2147483647, -2147483648, true},
		{-2147483648, 0, -2147483648, true},
	}

	for _, test := range cases {
		if got := NthMatches(test.a, test.b, test.pos); got != test.want {
			t.Errorf("NthMatches(%d, %d, %d) = %t, want %t",
				test.a, test.b, test.pos, got, test.want)
		}
	}
}

func TestNthMatchesHostIntBoundaries(t *testing.T) {
	cases := []struct {
		wordBits  int
		a, b, pos int64
		want      bool
	}{
		{32, 1, -2147483648, 2147483647, true},
		{32, -1, 2147483647, -2147483648, true},
		{32, -2147483648, 0, -2147483648, true},
		{32, -2147483648, -2147483648, -2147483648, true},
		{64, 1, -9223372036854775808, 9223372036854775807, true},
		{64, -1, 9223372036854775807, -9223372036854775808, true},
		{64, -9223372036854775808, 0, -9223372036854775808, true},
		{64, -9223372036854775808, -9223372036854775808, -9223372036854775808, true},
	}

	for _, test := range cases {
		if test.wordBits != strconv.IntSize {
			continue
		}
		if got := NthMatches(int(test.a), int(test.b), int(test.pos)); got != test.want {
			t.Errorf("NthMatches(%d, %d, %d) = %t, want %t",
				int(test.a), int(test.b), int(test.pos), got, test.want)
		}
	}
}

func TestNthTextLiteralCases(t *testing.T) {
	cases := []struct {
		a, b int
		want string
	}{
		{0, 0, "0"},
		{0, -12, "-12"},
		{1, 0, "n"},
		{-1, 0, "-n"},
		{2, 1, "2n+1"},
		{2, 0, "2n"},
		{-1, 3, "-n+3"},
		{2, -1, "2n-1"},
		{-2147483648, -2147483648, "-2147483648n-2147483648"},
	}

	for _, test := range cases {
		if got := NthText(test.a, test.b); got != test.want {
			t.Errorf("NthText(%d, %d) = %q, want %q", test.a, test.b, got, test.want)
		}
	}
}

func TestSimpleTextLiteralCases(t *testing.T) {
	cases := []struct {
		name   string
		simple Simple
		want   string
	}{
		{"type", Simple{Kind: TypeSel, Name: "A"}, "A"},
		{"universal", Simple{Kind: AnySel}, "*"},
		{"ID", Simple{Kind: IDSel, Name: "app"}, "#app"},
		{"class", Simple{Kind: ClassSel, Name: "on"}, ".on"},
		{"presence", Simple{Kind: AttrSel, Name: "title"}, "[title]"},
		{"escaped value", Simple{Kind: AttrSel, Name: "title", Op: "=", Value: "say \"hi\"\\", Flag: "i"}, "[title=\"say \\\"hi\\\"\\\\\" i]"},
		{"empty value with sensitive flag", Simple{Kind: AttrSel, Name: "x", Op: "=", Flag: "s"}, "[x=\"\" s]"},
		{"ordinary pseudo", Simple{Kind: PseudoSel, Name: "root"}, ":root"},
		{"nth child", Simple{Kind: PseudoSel, Name: "nth-child", A: 2, B: 1}, ":nth-child(2n+1)"},
		{"reverse child", Simple{Kind: PseudoSel, Name: "nth-last-child", A: -1, B: 3}, ":nth-last-child(-n+3)"},
		{"nth type", Simple{Kind: PseudoSel, Name: "nth-of-type", B: 2}, ":nth-of-type(2)"},
		{"reverse type", Simple{Kind: PseudoSel, Name: "nth-last-of-type", A: 1}, ":nth-last-of-type(n)"},
		{"word operator with insensitive flag", Simple{Kind: AttrSel, Name: "data-k", Op: "~=", Value: "word", Flag: "i"}, "[data-k~=\"word\" i]"},
		{"hyphen operator with sensitive flag", Simple{Kind: AttrSel, Name: "lang", Op: "|=", Value: "en", Flag: "s"}, "[lang|=\"en\" s]"},
		{"prefix operator", Simple{Kind: AttrSel, Name: "name", Op: "^=", Value: "pre"}, "[name^=\"pre\"]"},
		{"suffix operator", Simple{Kind: AttrSel, Name: "name", Op: "$=", Value: "post"}, "[name$=\"post\"]"},
		{"substring operator", Simple{Kind: AttrSel, Name: "name", Op: "*=", Value: "mid"}, "[name*=\"mid\"]"},
		{"non-ASCII value preserved", Simple{Kind: AttrSel, Name: "title", Op: "=", Value: "café 東京"}, "[title=\"café 東京\"]"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := SimpleText(test.simple); got != test.want {
				t.Errorf("SimpleText(%#v) = %q, want %q", test.simple, got, test.want)
			}
		})
	}
}

func TestCompoundAndComplexTextLiteralCases(t *testing.T) {
	cases := []struct {
		name    string
		complex Complex
		want    string
	}{
		{"compound order", Complex{Compounds: []Compound{{{Kind: TypeSel, Name: "li"}, {Kind: ClassSel, Name: "on"}}}}, "li.on"},
		{"all combinators", Complex{
			Compounds: []Compound{{{Kind: TypeSel, Name: "a"}}, {{Kind: TypeSel, Name: "b"}}, {{Kind: TypeSel, Name: "c"}}, {{Kind: TypeSel, Name: "d"}}, {{Kind: TypeSel, Name: "e"}}},
			Combs:     []string{">", "+", "~", " "},
		}, "a > b + c ~ d e"},
		{"child lead", Complex{Lead: ">", Compounds: []Compound{{{Kind: TypeSel, Name: "f"}}}}, "> f"},
		{"adjacent lead", Complex{Lead: "+", Compounds: []Compound{{{Kind: TypeSel, Name: "g"}}}}, "+ g"},
		{"later sibling lead", Complex{Lead: "~", Compounds: []Compound{{{Kind: TypeSel, Name: "h"}}}}, "~ h"},
		{"empty complex", Complex{}, ""},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := ComplexText(&test.complex); got != test.want {
				t.Errorf("ComplexText(%#v) = %q, want %q", test.complex, got, test.want)
			}
		})
	}

	if got := CompoundText(Compound{}); got != "" {
		t.Errorf("CompoundText(empty) = %q, want empty text", got)
	}
}

func TestFunctionalAndListTextLiteralCases(t *testing.T) {
	cases := []struct {
		name   string
		simple Simple
		want   string
	}{
		{"relative has list", Simple{Kind: PseudoSel, Name: "has", Args: List{
			&Complex{Lead: ">", Compounds: []Compound{{{Kind: TypeSel, Name: "f"}}}},
			&Complex{Lead: "+", Compounds: []Compound{{{Kind: TypeSel, Name: "g"}}}},
		}}, ":has(> f, + g)"},
		{"negation list", Simple{Kind: PseudoSel, Name: "not", Args: List{
			&Complex{Compounds: []Compound{{{Kind: ClassSel, Name: "on"}}}},
			&Complex{Compounds: []Compound{{{Kind: IDSel, Name: "x"}}}},
		}}, ":not(.on, #x)"},
		{"inclusion list", Simple{Kind: PseudoSel, Name: "is", Args: List{
			&Complex{Compounds: []Compound{{{Kind: TypeSel, Name: "a"}}}},
			&Complex{Compounds: []Compound{{{Kind: TypeSel, Name: "b"}}}},
		}}, ":is(a, b)"},
		{"where list", Simple{Kind: PseudoSel, Name: "where", Args: List{
			&Complex{Compounds: []Compound{{{Kind: TypeSel, Name: "p"}}}},
		}}, ":where(p)"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := SimpleText(test.simple); got != test.want {
				t.Errorf("SimpleText(%#v) = %q, want %q", test.simple, got, test.want)
			}
		})
	}

	list := List{
		&Complex{Compounds: []Compound{{{Kind: TypeSel, Name: "a"}}}},
		&Complex{Compounds: []Compound{{{Kind: TypeSel, Name: "b"}}}},
	}
	if got := ListText(list); got != "a, b" {
		t.Errorf("ListText(%#v) = %q, want %q", list, got, "a, b")
	}
	if got := ListText(List{}); got != "" {
		t.Errorf("ListText(empty) = %q, want empty text", got)
	}
}

func TestWeighLiteralCases(t *testing.T) {
	cases := []struct {
		name     string
		selector Complex
		want     Weight
	}{
		{
			name: "*",
			selector: Complex{Compounds: []Compound{
				{Simple{Kind: AnySel}},
			}},
			want: Weight{0, 0, 0},
		},
		{
			name: "div#app.on[title]:root:nth-child(2n+1)",
			selector: Complex{Compounds: []Compound{
				{
					Simple{Kind: TypeSel, Name: "div"},
					Simple{Kind: IDSel, Name: "app"},
					Simple{Kind: ClassSel, Name: "on"},
					Simple{Kind: AttrSel, Name: "title"},
					Simple{Kind: PseudoSel, Name: "root"},
					Simple{Kind: PseudoSel, Name: "nth-child", A: 2, B: 1},
				},
			}},
			want: Weight{1, 4, 1},
		},
		{
			name: "#app .nav > li:not(.on, #x)",
			selector: Complex{
				Compounds: []Compound{
					{Simple{Kind: IDSel, Name: "app"}},
					{Simple{Kind: ClassSel, Name: "nav"}},
					{
						Simple{Kind: TypeSel, Name: "li"},
						Simple{Kind: PseudoSel, Name: "not", Args: List{
							&Complex{Compounds: []Compound{{Simple{Kind: ClassSel, Name: "on"}}}},
							&Complex{Compounds: []Compound{{Simple{Kind: IDSel, Name: "x"}}}},
						}},
					},
				},
				Combs: []string{" ", ">"},
			},
			want: Weight{2, 1, 1},
		},
		{
			name: ":where(#x.on) em",
			selector: Complex{Compounds: []Compound{
				{
					Simple{Kind: PseudoSel, Name: "where", Args: List{
						&Complex{Compounds: []Compound{{
							Simple{Kind: IDSel, Name: "x"},
							Simple{Kind: ClassSel, Name: "on"},
						}}},
					}},
				},
				{Simple{Kind: TypeSel, Name: "em"}},
			}},
			want: Weight{0, 0, 1},
		},
		{
			name: ":is(.a.b, #x)",
			selector: Complex{Compounds: []Compound{{
				Simple{Kind: PseudoSel, Name: "is", Args: List{
					&Complex{Compounds: []Compound{{
						Simple{Kind: ClassSel, Name: "a"},
						Simple{Kind: ClassSel, Name: "b"},
					}}},
					&Complex{Compounds: []Compound{{Simple{Kind: IDSel, Name: "x"}}}},
				}},
			}}},
			want: Weight{1, 0, 0},
		},
		{
			name: ":has(> a.on, + #x)",
			selector: Complex{Compounds: []Compound{{
				Simple{Kind: PseudoSel, Name: "has", Args: List{
					&Complex{Lead: ">", Compounds: []Compound{{
						Simple{Kind: TypeSel, Name: "a"},
						Simple{Kind: ClassSel, Name: "on"},
					}}},
					&Complex{Lead: "+", Compounds: []Compound{{Simple{Kind: IDSel, Name: "x"}}}},
				}},
			}}},
			want: Weight{1, 0, 0},
		},
		{
			name: ":not(:is(a, .on))",
			selector: Complex{Compounds: []Compound{{
				Simple{Kind: PseudoSel, Name: "not", Args: List{
					&Complex{Compounds: []Compound{{
						Simple{Kind: PseudoSel, Name: "is", Args: List{
							&Complex{Compounds: []Compound{{Simple{Kind: TypeSel, Name: "a"}}}},
							&Complex{Compounds: []Compound{{Simple{Kind: ClassSel, Name: "on"}}}},
						}},
					}}},
				}},
			}}},
			want: Weight{0, 1, 0},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := Weigh(&test.selector); got != test.want {
				t.Errorf("Weigh(%#v) = %v, want %v", test.selector, got, test.want)
			}
		})
	}
}

func TestWeightComparisonAndTextLiterals(t *testing.T) {
	cases := []struct {
		name  string
		left  Weight
		right Weight
		less  bool
	}{
		{"first coordinate", Weight{0, 9, 9}, Weight{1, 0, 0}, true},
		{"second coordinate", Weight{1, 0, 9}, Weight{1, 1, 0}, true},
		{"third coordinate", Weight{1, 1, 1}, Weight{1, 1, 2}, true},
		{"greater", Weight{1, 1, 2}, Weight{1, 1, 1}, false},
		{"equal", Weight{1, 1, 1}, Weight{1, 1, 1}, false},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := test.left.Less(test.right); got != test.less {
				t.Errorf("%v.Less(%v) = %t, want %t", test.left, test.right, got, test.less)
			}
		})
	}

	if got := (Weight{2, 1, 1}).String(); got != "2,1,1" {
		t.Errorf("Weight{2, 1, 1}.String() = %q, want %q", got, "2,1,1")
	}
}
