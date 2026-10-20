package sheet

import (
	"reflect"
	"testing"

	"github.com/Gulliby-development/selwhy/internal/markup"
	"github.com/Gulliby-development/selwhy/internal/selector"
)

func TestParseLiteralSuccesses(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		selectors []string
		decls     [][]Decl
	}{
		{
			name:      "ordered declarations",
			input:     "a{x:1;y:2}",
			selectors: []string{"a"},
			decls:     [][]Decl{{{Prop: "x", Value: "1"}, {Prop: "y", Value: "2"}}},
		},
		{
			name:      "trim blanks and split at first colon",
			input:     " a , b { ; x : a:b ; ; } ",
			selectors: []string{"a, b"},
			decls:     [][]Decl{{{Prop: "x", Value: "a:b"}}},
		},
		{
			name:      "property name forms",
			input:     "a{_x:1;--x:2;-:3;X9:4}",
			selectors: []string{"a"},
			decls: [][]Decl{{
				{Prop: "_x", Value: "1"},
				{Prop: "--x", Value: "2"},
				{Prop: "-", Value: "3"},
				{Prop: "X9", Value: "4"},
			}},
		},
		{
			name:      "important marker with supported blanks",
			input:     "a{x:1 ! important;y:2!important;z:3 !\timportant}",
			selectors: []string{"a"},
			decls: [][]Decl{{
				{Prop: "x", Value: "1", Important: true},
				{Prop: "y", Value: "2", Important: true},
				{Prop: "z", Value: "3", Important: true},
			}},
		},
		{
			name:      "non-important values remain uninterpreted",
			input:     "a{x:!IMPORTANT;y:important;z:1!important tail}",
			selectors: []string{"a"},
			decls: [][]Decl{{
				{Prop: "x", Value: "!IMPORTANT"},
				{Prop: "y", Value: "important"},
				{Prop: "z", Value: "1!important tail"},
			}},
		},
		{
			name:      "empty bodies",
			input:     "a{} b{;;;}",
			selectors: []string{"a", "b"},
			decls:     [][]Decl{nil, nil},
		},
		{
			name:      "five specified blanks",
			input:     "\t\r\n\fa\t{\fx\r:\n1\t;}",
			selectors: []string{"a"},
			decls:     [][]Decl{{{Prop: "x", Value: "1"}}},
		},
		{
			name:      "comments replaced by one space",
			input:     "/* head */a/**/b{x:one/**/two}/* end */",
			selectors: []string{"a b"},
			decls:     [][]Decl{{{Prop: "x", Value: "one two"}}},
		},
		{
			name:      "comment between rules",
			input:     "a{x:1}/*ok*/b{y:2}",
			selectors: []string{"a", "b"},
			decls: [][]Decl{
				{{Prop: "x", Value: "1"}},
				{{Prop: "y", Value: "2"}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Parse(test.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", test.input, err)
			}
			if len(got) != len(test.selectors) {
				t.Fatalf("Parse(%q) returned %d rules, want %d", test.input, len(got), len(test.selectors))
			}
			for i, rule := range got {
				if gotSelector := selector.ListText(rule.Selectors); gotSelector != test.selectors[i] {
					t.Errorf("rule %d selector = %q, want %q", i+1, gotSelector, test.selectors[i])
				}
				if !reflect.DeepEqual(rule.Decls, test.decls[i]) {
					t.Errorf("rule %d declarations = %#v, want %#v", i+1, rule.Decls, test.decls[i])
				}
			}
		})
	}
}

func TestParseLiteralDiagnostics(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty input",
			input: "",
			want:  "the style sheet has no rules",
		},
		{
			name:  "comment-only input",
			input: " /*ok*/\t\n",
			want:  "the style sheet has no rules",
		},
		{
			name:  "unclosed comment",
			input: "/* x",
			want:  "unterminated comment starting at line 1 column 1",
		},
		{
			name:  "comment coordinates across lines",
			input: "/*é\n*/\n  /* x",
			want:  "unterminated comment starting at line 3 column 3",
		},
		{
			name:  "comment coordinates count runes",
			input: "é/* x",
			want:  "unterminated comment starting at line 1 column 2",
		},
		{
			name:  "comment preprocessing precedes rule errors",
			input: "@x /*",
			want:  "unterminated comment starting at line 1 column 4",
		},
		{
			name:  "at-rule",
			input: "@media{x:1}",
			want:  "rule 1: at-rules are not supported",
		},
		{
			name:  "missing opening brace",
			input: "a",
			want:  "rule 1: missing {",
		},
		{
			name:  "selector diagnostic",
			input: "a:foo{x:y}",
			want:  "rule 1: unknown pseudo-class \":foo\" at column 2",
		},
		{
			name:  "selector columns use trimmed text",
			input: "  a > {x:y}",
			want:  "rule 1: expected a selector at column 4",
		},
		{
			name:  "missing closing brace",
			input: "a{x:1",
			want:  "rule 1: missing }",
		},
		{
			name:  "missing closing brace precedes nested opening brace",
			input: "a{{",
			want:  "rule 1: missing }",
		},
		{
			name:  "nested opening brace",
			input: "a{{x:1}",
			want:  "rule 1: unexpected {",
		},
		{
			name:  "rule numbering",
			input: "a{} @x",
			want:  "rule 2: at-rules are not supported",
		},
		{
			name:  "selector error precedes missing closing brace",
			input: "a:foo{",
			want:  "rule 1: unknown pseudo-class \":foo\" at column 2",
		},
		{
			name:  "declaration without colon",
			input: "a{x:1;y}",
			want:  "rule 1: declaration 2 has no colon",
		},
		{
			name:  "declaration numbering skips blank parts",
			input: "a{;;x:1; ;y}",
			want:  "rule 1: declaration 2 has no colon",
		},
		{
			name:  "declaration numbering restarts per rule",
			input: "a{}b{y}",
			want:  "rule 2: declaration 1 has no colon",
		},
		{
			name:  "invalid property initial character",
			input: "a{1x:2}",
			want:  "rule 1: invalid property name \"1x\"",
		},
		{
			name:  "invalid property later character",
			input: "a{x.y:2}",
			want:  "rule 1: invalid property name \"x.y\"",
		},
		{
			name:  "empty property",
			input: "a{:2}",
			want:  "rule 1: invalid property name \"\"",
		},
		{
			name:  "invalid property diagnostic uses ASCII display",
			input: "a{é:2}",
			want:  "rule 1: invalid property name \"?\"",
		},
		{
			name:  "empty value",
			input: "a{x:}",
			want:  "rule 1: property x has no value",
		},
		{
			name:  "empty important value",
			input: "a{x: !important}",
			want:  "rule 1: property x has no value",
		},
		{
			name:  "empty spaced important value",
			input: "a{x: ! important}",
			want:  "rule 1: property x has no value",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.input)
			if err == nil {
				t.Fatalf("Parse(%q) returned no error, want %q", test.input, test.want)
			}
			if got := err.Error(); got != test.want {
				t.Errorf("Parse(%q) error = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func parseCascadeTarget(t *testing.T, text string, index int) *markup.Element {
	t.Helper()
	doc, err := markup.Parse(text)
	if err != nil {
		t.Fatalf("markup.Parse(%q): %v", text, err)
	}
	return doc.Elements[index]
}

func TestCascadeLiteralWinnersAndDeterminism(t *testing.T) {
	const document = `<a id="i" class="c"/>`
	cases := []struct {
		name       string
		sheet      string
		document   string
		target     int
		emptyRules bool
		want       []Winner
	}{
		{
			name:  "importance beats normal declaration",
			sheet: "a{x:1!important}#i{x:2}",
			want: []Winner{
				{Prop: "x", Value: "1", Important: true, Rule: 1, Weight: selector.Weight{0, 0, 1}},
			},
		},
		{
			name:  "id specificity",
			sheet: "#i{x:1}.c{x:2}",
			want: []Winner{
				{Prop: "x", Value: "1", Important: false, Rule: 1, Weight: selector.Weight{1, 0, 0}},
			},
		},
		{
			name:  "class specificity",
			sheet: ".c{x:1}a{x:2}",
			want: []Winner{
				{Prop: "x", Value: "1", Important: false, Rule: 1, Weight: selector.Weight{0, 1, 0}},
			},
		},
		{
			name:  "type specificity",
			sheet: "a{x:1}*{x:2}",
			want: []Winner{
				{Prop: "x", Value: "1", Important: false, Rule: 1, Weight: selector.Weight{0, 0, 1}},
			},
		},
		{
			name:  "later rule wins specificity tie",
			sheet: "a{x:1}a{x:2}",
			want: []Winner{
				{Prop: "x", Value: "2", Important: false, Rule: 2, Weight: selector.Weight{0, 0, 1}},
			},
		},
		{
			name:  "later declaration wins within rule",
			sheet: "a{x:1;x:2}",
			want: []Winner{
				{Prop: "x", Value: "2", Important: false, Rule: 1, Weight: selector.Weight{0, 0, 1}},
			},
		},
		{
			name:  "important beats later normal declaration",
			sheet: "a{x:1!important;x:2}",
			want: []Winner{
				{Prop: "x", Value: "1", Important: true, Rule: 1, Weight: selector.Weight{0, 0, 1}},
			},
		},
		{
			name:  "later important declaration wins tie",
			sheet: "a{x:1!important;x:2!important}",
			want: []Winner{
				{Prop: "x", Value: "2", Important: true, Rule: 1, Weight: selector.Weight{0, 0, 1}},
			},
		},
		{
			name:  "important specificity",
			sheet: "#i{x:1!important}.c{x:2!important}",
			want: []Winner{
				{Prop: "x", Value: "1", Important: true, Rule: 1, Weight: selector.Weight{1, 0, 0}},
			},
		},
		{
			name:  "later important rule wins tie",
			sheet: "a{x:1!important}a{x:2!important}",
			want: []Winner{
				{Prop: "x", Value: "2", Important: true, Rule: 2, Weight: selector.Weight{0, 0, 1}},
			},
		},
		{
			name:  "unmatched selectors do not contribute specificity",
			sheet: "a,#missing{x:1}.c{x:2}",
			want: []Winner{
				{Prop: "x", Value: "2", Important: false, Rule: 2, Weight: selector.Weight{0, 1, 0}},
			},
		},
		{
			name:  "greatest matching selector specificity",
			sheet: "a,#i{x:1}.c{x:2}",
			want: []Winner{
				{Prop: "x", Value: "1", Important: false, Rule: 1, Weight: selector.Weight{1, 0, 0}},
			},
		},
		{
			name:  "matching selector order does not affect greatest weight",
			sheet: "#i,a{x:1}.c{x:2}",
			want: []Winner{
				{Prop: "x", Value: "1", Important: false, Rule: 1, Weight: selector.Weight{1, 0, 0}},
			},
		},
		{
			name:  "properties sort by code point and unmatched rule is ignored",
			sheet: "b{x:1}a{z:9;A:1;--x:2;a:3;_x:4}",
			want: []Winner{
				{Prop: "--x", Value: "2", Important: false, Rule: 2, Weight: selector.Weight{0, 0, 1}},
				{Prop: "A", Value: "1", Important: false, Rule: 2, Weight: selector.Weight{0, 0, 1}},
				{Prop: "_x", Value: "4", Important: false, Rule: 2, Weight: selector.Weight{0, 0, 1}},
				{Prop: "a", Value: "3", Important: false, Rule: 2, Weight: selector.Weight{0, 0, 1}},
				{Prop: "z", Value: "9", Important: false, Rule: 2, Weight: selector.Weight{0, 0, 1}},
			},
		},
		{
			name:  "properties resolve independently",
			sheet: "a{x:1;y:2}.c{x:3}",
			want: []Winner{
				{Prop: "x", Value: "3", Important: false, Rule: 2, Weight: selector.Weight{0, 1, 0}},
				{Prop: "y", Value: "2", Important: false, Rule: 1, Weight: selector.Weight{0, 0, 1}},
			},
		},
		{
			name:  "unmatched rules have no winners",
			sheet: "b{x:1}",
			want:  []Winner{},
		},
		{
			name:  "matched rule with no declarations has no winners",
			sheet: "a{}",
			want:  []Winner{},
		},
		{
			name:       "empty rule slice has no winners",
			emptyRules: true,
			want:       []Winner{},
		},
		{
			name:     "declarations do not inherit",
			sheet:    "a{x:1}",
			document: "<a><b/></a>",
			target:   1,
			want:     []Winner{},
		},
	}

	defaultTarget := parseCascadeTarget(t, document, 0)
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			target := defaultTarget
			if test.document != "" {
				target = parseCascadeTarget(t, test.document, test.target)
			}
			rules := []Rule{}
			if !test.emptyRules {
				parsed, err := Parse(test.sheet)
				if err != nil {
					t.Fatalf("Parse(%q): %v", test.sheet, err)
				}
				rules = parsed
			}
			for run := 0; run < 2; run++ {
				got := Cascade(rules, target)
				if !reflect.DeepEqual(got, test.want) {
					t.Errorf("Cascade run %d = %#v, want %#v", run+1, got, test.want)
				}
			}
		})
	}
}
