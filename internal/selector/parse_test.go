package selector

import (
	"reflect"
	"testing"
)

func TestParseLiteralASTsAndNesting(t *testing.T) {
	oneNot := List{
		&Complex{Compounds: []Compound{{Simple{Kind: TypeSel, Name: "a"}}}},
	}
	twoNots := List{
		&Complex{Compounds: []Compound{{
			Simple{Kind: PseudoSel, Name: "not", Args: oneNot},
		}}},
	}
	threeNots := List{
		&Complex{Compounds: []Compound{{
			Simple{Kind: PseudoSel, Name: "not", Args: twoNots},
		}}},
	}
	fourNots := List{
		&Complex{Compounds: []Compound{{
			Simple{Kind: PseudoSel, Name: "not", Args: threeNots},
		}}},
	}
	fiveNots := List{
		&Complex{Compounds: []Compound{{
			Simple{Kind: PseudoSel, Name: "not", Args: fourNots},
		}}},
	}
	sixNots := List{
		&Complex{Compounds: []Compound{{
			Simple{Kind: PseudoSel, Name: "not", Args: fiveNots},
		}}},
	}
	sevenNots := List{
		&Complex{Compounds: []Compound{{
			Simple{Kind: PseudoSel, Name: "not", Args: sixNots},
		}}},
	}
	eightNots := List{
		&Complex{Compounds: []Compound{{
			Simple{Kind: PseudoSel, Name: "not", Args: sevenNots},
		}}},
	}

	cases := []struct {
		name  string
		input string
		want  List
		err   string
	}{
		// AC-001: simple selectors, ordering, lists and combinators.
		{
			name:  "type",
			input: "a",
			want:  List{&Complex{Compounds: []Compound{{Simple{Kind: TypeSel, Name: "a"}}}}},
		},
		{
			name:  "universal",
			input: "*",
			want:  List{&Complex{Compounds: []Compound{{Simple{Kind: AnySel}}}}},
		},
		{
			name:  "identifier",
			input: "-_a9-X",
			want:  List{&Complex{Compounds: []Compound{{Simple{Kind: TypeSel, Name: "-_a9-X"}}}}},
		},
		{
			name:  "ordered compound",
			input: "A#id.c[x]:root",
			want: List{&Complex{Compounds: []Compound{{
				Simple{Kind: TypeSel, Name: "A"},
				Simple{Kind: IDSel, Name: "id"},
				Simple{Kind: ClassSel, Name: "c"},
				Simple{Kind: AttrSel, Name: "x"},
				Simple{Kind: PseudoSel, Name: "root"},
			}}}},
		},
		{
			name:  "selector list",
			input: "a, b",
			want: List{
				&Complex{Compounds: []Compound{{Simple{Kind: TypeSel, Name: "a"}}}},
				&Complex{Compounds: []Compound{{Simple{Kind: TypeSel, Name: "b"}}}},
			},
		},
		{
			name:  "all combinators",
			input: "a b>c+d~e",
			want: List{&Complex{
				Compounds: []Compound{
					{Simple{Kind: TypeSel, Name: "a"}},
					{Simple{Kind: TypeSel, Name: "b"}},
					{Simple{Kind: TypeSel, Name: "c"}},
					{Simple{Kind: TypeSel, Name: "d"}},
					{Simple{Kind: TypeSel, Name: "e"}},
				},
				Combs: []string{" ", ">", "+", "~"},
			}},
		},
		{
			name:  "blanks around combinator",
			input: " \t a\r\n>\f b ",
			want: List{&Complex{
				Compounds: []Compound{
					{Simple{Kind: TypeSel, Name: "a"}},
					{Simple{Kind: TypeSel, Name: "b"}},
				},
				Combs: []string{">"},
			}},
		},

		// AC-002: attributes, operators, flags, values and supported escapes.
		{
			name:  "attribute presence",
			input: "[x]",
			want:  List{&Complex{Compounds: []Compound{{Simple{Kind: AttrSel, Name: "x"}}}}},
		},
		{
			name:  "attribute equality",
			input: "[x=y]",
			want: List{&Complex{Compounds: []Compound{{Simple{
				Kind: AttrSel, Name: "x", Op: "=", Value: "y",
			}}}}},
		},
		{
			name:  "attribute word and flag",
			input: "[ x ~= y i ]",
			want: List{&Complex{Compounds: []Compound{{Simple{
				Kind: AttrSel, Name: "x", Op: "~=", Value: "y", Flag: "i",
			}}}}},
		},
		{
			name:  "attribute language operator",
			input: "[x|=en]",
			want: List{&Complex{Compounds: []Compound{{Simple{
				Kind: AttrSel, Name: "x", Op: "|=", Value: "en",
			}}}}},
		},
		{
			name:  "attribute prefix",
			input: "[x^='a']",
			want: List{&Complex{Compounds: []Compound{{Simple{
				Kind: AttrSel, Name: "x", Op: "^=", Value: "a",
			}}}}},
		},
		{
			name:  "attribute suffix and adjacent flag",
			input: "[x$=\"b\"s]",
			want: List{&Complex{Compounds: []Compound{{Simple{
				Kind: AttrSel, Name: "x", Op: "$=", Value: "b", Flag: "s",
			}}}}},
		},
		{
			name:  "empty substring value",
			input: `[x*=""]`,
			want: List{&Complex{Compounds: []Compound{{Simple{
				Kind: AttrSel, Name: "x", Op: "*=", Value: "",
			}}}}},
		},
		{
			name:  "Unicode attribute value",
			input: "[x=\"é\"]",
			want: List{&Complex{Compounds: []Compound{{Simple{
				Kind: AttrSel, Name: "x", Op: "=", Value: "é",
			}}}}},
		},
		{
			name:  "quoted escapes",
			input: `[x="a\"b\'c\\d"]`,
			want: List{&Complex{Compounds: []Compound{{Simple{
				Kind: AttrSel, Name: "x", Op: "=", Value: `a"b'c\d`,
			}}}}},
		},

		// AC-003: no-argument, formula and list pseudo-classes.
		{
			name:  "no-argument pseudo-classes",
			input: ":first-child:last-child:only-child:first-of-type:last-of-type:only-of-type:empty:root",
			want: List{&Complex{Compounds: []Compound{{
				Simple{Kind: PseudoSel, Name: "first-child"},
				Simple{Kind: PseudoSel, Name: "last-child"},
				Simple{Kind: PseudoSel, Name: "only-child"},
				Simple{Kind: PseudoSel, Name: "first-of-type"},
				Simple{Kind: PseudoSel, Name: "last-of-type"},
				Simple{Kind: PseudoSel, Name: "only-of-type"},
				Simple{Kind: PseudoSel, Name: "empty"},
				Simple{Kind: PseudoSel, Name: "root"},
			}}}},
		},
		{
			name:  "odd formula",
			input: ":nth-child(odd)",
			want: List{&Complex{Compounds: []Compound{{Simple{
				Kind: PseudoSel, Name: "nth-child", A: 2, B: 1,
			}}}}},
		},
		{
			name:  "even formula",
			input: ":nth-last-child(even)",
			want: List{&Complex{Compounds: []Compound{{Simple{
				Kind: PseudoSel, Name: "nth-last-child", A: 2, B: 0,
			}}}}},
		},
		{
			name:  "integer formula",
			input: ":nth-of-type(-12)",
			want: List{&Complex{Compounds: []Compound{{Simple{
				Kind: PseudoSel, Name: "nth-of-type", A: 0, B: -12,
			}}}}},
		},
		{
			name:  "spaced formula offset",
			input: ":nth-last-of-type(2n - 1)",
			want: List{&Complex{Compounds: []Compound{{Simple{
				Kind: PseudoSel, Name: "nth-last-of-type", A: 2, B: -1,
			}}}}},
		},
		{
			name:  "explicit positive formula sign",
			input: ":nth-child(+n)",
			want: List{&Complex{Compounds: []Compound{{Simple{
				Kind: PseudoSel, Name: "nth-child", A: 1, B: 0,
			}}}}},
		},
		{
			name:  "zero formula coefficient",
			input: ":nth-child(0n+2)",
			want: List{&Complex{Compounds: []Compound{{Simple{
				Kind: PseudoSel, Name: "nth-child", A: 0, B: 2,
			}}}}},
		},
		{
			name:  "list pseudo-classes",
			input: ":not(a,.b):is(#c):where(*)",
			want: List{&Complex{Compounds: []Compound{{
				Simple{Kind: PseudoSel, Name: "not", Args: List{
					&Complex{Compounds: []Compound{{Simple{Kind: TypeSel, Name: "a"}}}},
					&Complex{Compounds: []Compound{{Simple{Kind: ClassSel, Name: "b"}}}},
				}},
				Simple{Kind: PseudoSel, Name: "is", Args: List{
					&Complex{Compounds: []Compound{{Simple{Kind: IDSel, Name: "c"}}}},
				}},
				Simple{Kind: PseudoSel, Name: "where", Args: List{
					&Complex{Compounds: []Compound{{Simple{Kind: AnySel}}}},
				}},
			}}}},
		},

		// AC-004: relative lists, nesting, and has-context handling.
		{
			name:  "relative has list",
			input: ":has(> a, + b, ~ c, d e)",
			want: List{&Complex{Compounds: []Compound{{Simple{
				Kind: PseudoSel, Name: "has", Args: List{
					&Complex{Lead: ">", Compounds: []Compound{{Simple{Kind: TypeSel, Name: "a"}}}},
					&Complex{Lead: "+", Compounds: []Compound{{Simple{Kind: TypeSel, Name: "b"}}}},
					&Complex{Lead: "~", Compounds: []Compound{{Simple{Kind: TypeSel, Name: "c"}}}},
					&Complex{
						Compounds: []Compound{
							{Simple{Kind: TypeSel, Name: "d"}},
							{Simple{Kind: TypeSel, Name: "e"}},
						},
						Combs: []string{" "},
					},
				},
			}}}}},
		},
		{
			name:  "sibling has context restoration",
			input: ":has(a):has(b)",
			want: List{&Complex{Compounds: []Compound{{
				Simple{Kind: PseudoSel, Name: "has", Args: List{
					&Complex{Compounds: []Compound{{Simple{Kind: TypeSel, Name: "a"}}}},
				}},
				Simple{Kind: PseudoSel, Name: "has", Args: List{
					&Complex{Compounds: []Compound{{Simple{Kind: TypeSel, Name: "b"}}}},
				}},
			}}}},
		},
		{
			name:  "eight-level list nesting",
			input: ":not(:not(:not(:not(:not(:not(:not(:not(a))))))))",
			want:  List{&Complex{Compounds: []Compound{{Simple{Kind: PseudoSel, Name: "not", Args: eightNots}}}}},
		},
		{
			name:  "ninth-level list nesting",
			input: ":not(:not(:not(:not(:not(:not(:not(:not(:not(a)))))))))",
			err:   "nesting is deeper than 8 at column 41",
		},
		{
			name:  "direct nested has",
			input: ":has(:has(a))",
			err:   ":has() cannot be nested in :has() at column 6",
		},
		{
			name:  "indirect nested has",
			input: ":has(:is(:has(a)))",
			err:   ":has() cannot be nested in :has() at column 10",
		},
		{
			name:  "leading combinator outside has",
			input: ":is(> a)",
			err:   "expected a selector at column 5",
		},
		{
			name:  "relative has without compound",
			input: ":has(>)",
			err:   "expected a selector at column 7",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := Parse(test.input)
			if test.err != "" {
				if err == nil || err.Error() != test.err {
					t.Fatalf("Parse(%q) error = %v, want %q", test.input, err, test.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", test.input, err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Parse(%q) = %#v, want %#v", test.input, got, test.want)
			}
		})
	}
}

func TestParseLiteralStructureAndNameDiagnostics(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty selector", "", "expected a selector at column 1"},
		{"invalid selector start", "1a", "expected a selector at column 1"},
		{"top-level combinator", ">a", "expected a selector at column 1"},
		{"missing compound after combinator", "a >", "expected a selector at column 4"},
		{"missing list member", "a,", "expected a selector at column 3"},
		{"misplaced universal selector", "a*", "unexpected \"*\" at column 2"},
		{"unexpected suffix", "a%", "unexpected \"%\" at column 2"},
		{"unexpected top-level close", "a)", "unexpected \")\" at column 2"},
		{"missing id name", "#", "expected a name at column 2"},
		{"invalid class name", ".1", "expected a name at column 2"},
		{"Unicode character column", "[x=\"é\"]%", "unexpected \"%\" at column 8"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.input)
			if err == nil || err.Error() != test.want {
				t.Errorf("Parse(%q) error = %v, want %q", test.input, err, test.want)
			}
		})
	}
}

func TestParseLiteralAttributeDiagnostics(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"missing attribute name", "[]", "expected a name at column 2"},
		{"missing attribute close", "[a", "expected ] at column 3"},
		{"incomplete attribute operator", "[a~b]", "expected = at column 4"},
		{"unexpected attribute operator", "[a?]", "unexpected \"?\" at column 3"},
		{"missing attribute value", "[a=]", "expected an attribute value at column 4"},
		{"missing close after value", "[a=x", "expected ] at column 5"},
		{"unterminated attribute string", "[a=\"x", "unterminated string starting at column 4"},
		{"attribute string ends after backslash", "[a=\"x\\", "unterminated string starting at column 4"},
		{"unsupported literal backslash-n escape", "[a=\"\\n\"]", "unsupported escape at column 5"},
		{"line feed in attribute string", "[a=\"x\n\"]", "unterminated string starting at column 4"},
		{"unexpected attribute flag", "[a=\"x\"q]", "unexpected \"q\" at column 7"},
		{"invalid unquoted attribute value", "[a=1]", "expected an attribute value at column 4"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.input)
			if err == nil || err.Error() != test.want {
				t.Errorf("Parse(%q) error = %v, want %q", test.input, err, test.want)
			}
		})
	}
}

func TestParseLiteralPseudoFormulaAndListDiagnostics(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"pseudo-element", "::before", "pseudo-elements are not supported at column 1"},
		{"missing pseudo name", ":", "expected a name at column 2"},
		{"no-argument pseudo with arguments", ":first-child(1)", ":first-child takes no arguments at column 1"},
		{"unknown pseudo", "a:hover", "unknown pseudo-class \":hover\" at column 2"},
		{"formula pseudo without arguments", ":nth-child", ":nth-child needs arguments at column 1"},
		{"list pseudo without arguments", ":is", ":is needs arguments at column 1"},
		{"invalid formula", ":nth-child(2n+)", "invalid an+b formula \"2n+\" at column 12"},
		{"blank formula", ":nth-child(   )", "invalid an+b formula \"\" at column 15"},
		{"missing formula close wins", ":nth-child(2n+", "expected ) at column 15"},
		{"unclosed list pseudo", ":is(a", "expected ) at column 6"},
		{"empty list pseudo", ":is()", "expected a selector at column 5"},
		{"trailing list comma", ":is(a,)", "expected a selector at column 7"},
		{"list member unexpected suffix", ":is(a%)", "unexpected \"%\" at column 6"},
		{"unknown pseudo before argument parse", ":hover([a", "unknown pseudo-class \":hover\" at column 1"},
		{"attribute fault before following pseudo", "[a~b]:hover", "expected = at column 4"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.input)
			if err == nil || err.Error() != test.want {
				t.Errorf("Parse(%q) error = %v, want %q", test.input, err, test.want)
			}
		})
	}
}
