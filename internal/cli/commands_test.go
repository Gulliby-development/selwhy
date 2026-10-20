package cli

import (
	"testing"

	"github.com/Gulliby-development/selwhy/internal/base"
)

const commandTestDOC1 = "<div id=\"app\" class=\"page wide\"><ul class=\"nav\"><li class=\"on\">Home</li><li data-k=\"b-2\">Docs &amp; FAQ</li><li></li></ul><p lang=\"en\">Hi</p></div>"

type commandTestIO struct {
	stdin   []byte
	reads   int
	outputs []string
	errors  []string
}

func (io *commandTestIO) ReadStdin() []byte {
	io.reads++
	return io.stdin
}

func (io *commandTestIO) Out(text string) {
	io.outputs = append(io.outputs, text)
}

func (io *commandTestIO) Err(text string) {
	io.errors = append(io.errors, text)
}

var _ base.IO = (*commandTestIO)(nil)

func TestCommandExamplesAndWorkedErrors(t *testing.T) {
	tests := []struct {
		name   string
		argv   []string
		stdin  []byte
		reads  int
		code   int
		stdout string
		stderr string
	}{
		{
			name: "tree example",
			argv: []string{"tree", commandTestDOC1},
			code: 0,
			stdout: "n  path                 tag\n" +
				"1  /div[1]              [<div id=\"app\" class=\"page wide\">]\n" +
				"2  /div[1]/ul[1]        [<ul class=\"nav\">]\n" +
				"3  /div[1]/ul[1]/li[1]  [<li class=\"on\">]\n" +
				"4  /div[1]/ul[1]/li[2]  [<li data-k=\"b-2\">]\n" +
				"5  /div[1]/ul[1]/li[3]  [<li>]\n" +
				"6  /div[1]/p[1]         [<p lang=\"en\">]\n" +
				"elements    6\n",
		},
		{
			name: "select example",
			argv: []string{"select", "ul > li:not(.on)", commandTestDOC1},
			code: 0,
			stdout: "n  path                 tag\n" +
				"4  /div[1]/ul[1]/li[2]  [<li data-k=\"b-2\">]\n" +
				"5  /div[1]/ul[1]/li[3]  [<li>]\n" +
				"matches     2\n",
		},
		{
			name: "select list example",
			argv: []string{"select", "li:nth-child(odd), p[lang|=en], div:has(> ul li:empty)", commandTestDOC1},
			code: 0,
			stdout: "n  path                 tag\n" +
				"1  /div[1]              [<div id=\"app\" class=\"page wide\">]\n" +
				"3  /div[1]/ul[1]/li[1]  [<li class=\"on\">]\n" +
				"5  /div[1]/ul[1]/li[3]  [<li>]\n" +
				"6  /div[1]/p[1]         [<p lang=\"en\">]\n" +
				"matches     4\n",
		},
		{
			name: "explain mixed results example",
			argv: []string{"explain", "ul > li.on, section li, li:nth-child(2n), li:empty", commandTestDOC1, "5"},
			code: 0,
			stdout: "element     5 /div[1]/ul[1]/li[3]\n" +
				"tag         [<li>]\n" +
				"selector    [ul > li.on]\n" +
				"result      no match\n" +
				"reason      fails [.on] (the classes are [])\n" +
				"selector    [section li]\n" +
				"result      no match\n" +
				"reason      no ancestor matches [section]\n" +
				"selector    [li:nth-child(2n)]\n" +
				"result      no match\n" +
				"reason      fails [:nth-child(2n)] (the position is 3 of 3)\n" +
				"selector    [li:empty]\n" +
				"result      match\n",
		},
		{
			name: "explain no matches example",
			argv: []string{"explain", "li:not(.on), li:is(.a, .b), li:has(> a)", commandTestDOC1, "3"},
			code: 1,
			stdout: "element     3 /div[1]/ul[1]/li[1]\n" +
				"tag         [<li class=\"on\">]\n" +
				"selector    [li:not(.on)]\n" +
				"result      no match\n" +
				"reason      fails [:not(.on)] (the element matches [.on])\n" +
				"selector    [li:is(.a, .b)]\n" +
				"result      no match\n" +
				"reason      fails [:is(.a, .b)] (the element matches none of the arguments)\n" +
				"selector    [li:has(> a)]\n" +
				"result      no match\n" +
				"reason      fails [:has(> a)] (no relative selector matches)\n",
		},
		{
			name: "specificity example",
			argv: []string{"specificity", "#app .nav > li:not(.on, #x), :where(p) em, a:nth-child(2n+1)"},
			code: 0,
			stdout: "selector                       specificity\n" +
				"[#app .nav > li:not(.on, #x)]  2,1,1\n" +
				"[:where(p) em]                 0,0,1\n" +
				"[a:nth-child(2n+1)]            0,1,1\n",
		},
		{
			name:   "fmt example",
			argv:   []string{"fmt", "A[ title = \"say \\\"hi\\\"\" i ]>b +c ~ d   e:nth-child( -n + 3 ):has(> f, + g)"},
			code:   0,
			stdout: "A[title=\"say \\\"hi\\\"\" i] > b + c ~ d e:nth-child(-n+3):has(> f, + g)\n",
		},
		{
			name: "nth example",
			argv: []string{"nth", "--count", "5", "-n+3"},
			code: 0,
			stdout: "formula     [-n+3]\n" +
				"a           -1\n" +
				"b           3\n" +
				"matches     1 2 3\n",
		},
		{
			name: "cascade example",
			argv: []string{"cascade", "li{margin:0} /* menu */ ul > li.on{--x:2;margin:1} #app li.on{--x:3} .on{margin:4 !important}", commandTestDOC1},
			code: 0,
			stdout: "n  path                 property  value  from\n" +
				"3  /div[1]/ul[1]/li[1]  --x       [3]    rule 3 (1,1,1)\n" +
				"3  /div[1]/ul[1]/li[1]  margin    [4]    rule 4 (0,1,0) !important\n" +
				"4  /div[1]/ul[1]/li[2]  margin    [0]    rule 1 (0,0,1)\n" +
				"5  /div[1]/ul[1]/li[3]  margin    [0]    rule 1 (0,0,1)\n" +
				"rules       4\n" +
				"styled      3\n",
		},
		{
			name:   "unknown option",
			argv:   []string{"nth", "--cnt", "3", "2n"},
			code:   2,
			stderr: "selwhy: unknown option \"--cnt\"\n",
		},
		{
			name:   "missing count value",
			argv:   []string{"nth", "2n", "--count"},
			code:   2,
			stderr: "selwhy: --count needs a value\n",
		},
		{
			name:   "invalid count",
			argv:   []string{"nth", "--count", "0", "2n"},
			code:   2,
			stderr: "selwhy: invalid --count \"0\" (expected a whole number from 1 to 100)\n",
		},
		{
			name:   "two stdin arguments",
			argv:   []string{"select", "-", "-"},
			code:   2,
			stderr: "selwhy: only one argument can be -\n",
		},
		{
			name:   "invalid element number",
			argv:   []string{"explain", "a", "<a/>", "0"},
			code:   2,
			stderr: "selwhy: invalid element number \"0\" (expected a whole number from 1 to 999999)\n",
		},
		{
			name:   "element does not exist",
			argv:   []string{"explain", "a", "<a><b/><c/></a>", "4"},
			code:   2,
			stderr: "selwhy: element 4 does not exist (the document has 3 elements)\n",
		},
		{
			name:   "invalid formula",
			argv:   []string{"nth", "2 n"},
			code:   2,
			stderr: "selwhy: invalid an+b formula \"2 n\"\n",
		},
		{
			name:   "outside text",
			argv:   []string{"tree", "a"},
			code:   2,
			stderr: "selwhy: text outside the root element at line 1 column 1\n",
		},
		{
			name:   "mismatched closing tag",
			argv:   []string{"tree", "<a><b></a>"},
			code:   2,
			stderr: "selwhy: closing tag </a> does not match <b> at line 1 column 7\n",
		},
		{
			name:   "duplicate attribute",
			argv:   []string{"tree", "<a x=\"1\" x=\"2\"></a>"},
			code:   2,
			stderr: "selwhy: duplicate attribute x at line 1 column 10\n",
		},
		{
			name:   "unterminated attribute value",
			argv:   []string{"tree", "<a b=\"c></a>"},
			code:   2,
			stderr: "selwhy: unterminated attribute value starting at line 1 column 6\n",
		},
		{
			name:   "unclosed element",
			argv:   []string{"tree", "<a>\n  <b>"},
			code:   2,
			stderr: "selwhy: unclosed element b starting at line 2 column 3\n",
		},
		{
			name:   "unknown entity",
			argv:   []string{"tree", "<a>&bogus;</a>"},
			code:   2,
			stderr: "selwhy: unknown entity at line 1 column 4\n",
		},
		{
			name:   "nesting limit",
			argv:   []string{"tree", "<a><a><a><a><a><a><a><a><a><a><a><a><a>"},
			code:   2,
			stderr: "selwhy: nesting is deeper than 12 at line 1 column 37\n",
		},
		{
			name:   "expected selector after combinator",
			argv:   []string{"fmt", "a >"},
			code:   2,
			stderr: "selwhy: expected a selector at column 4\n",
		},
		{
			name:   "unexpected selector character",
			argv:   []string{"fmt", "a%"},
			code:   2,
			stderr: "selwhy: unexpected \"%\" at column 2\n",
		},
		{
			name:   "missing attribute bracket",
			argv:   []string{"fmt", "[a"},
			code:   2,
			stderr: "selwhy: expected ] at column 3\n",
		},
		{
			name:   "attribute operator missing equals",
			argv:   []string{"fmt", "[a~b]"},
			code:   2,
			stderr: "selwhy: expected = at column 4\n",
		},
		{
			name:   "unterminated selector string",
			argv:   []string{"fmt", "[a=\"x"},
			code:   2,
			stderr: "selwhy: unterminated string starting at column 4\n",
		},
		{
			name:   "unsupported selector string escape",
			argv:   []string{"fmt", "[a=\"\\n\"]"},
			code:   2,
			stderr: "selwhy: unsupported escape at column 5\n",
		},
		{
			name:   "pseudo-class arguments are forbidden",
			argv:   []string{"fmt", ":first-child(1)"},
			code:   2,
			stderr: "selwhy: :first-child takes no arguments at column 1\n",
		},
		{
			name:   "unknown pseudo-class",
			argv:   []string{"fmt", "a:hover"},
			code:   2,
			stderr: "selwhy: unknown pseudo-class \":hover\" at column 2\n",
		},
		{
			name:   "invalid nth formula",
			argv:   []string{"fmt", ":nth-child(2n+)"},
			code:   2,
			stderr: "selwhy: invalid an+b formula \"2n+\" at column 12\n",
		},
		{
			name:   "unclosed selector list argument",
			argv:   []string{"fmt", ":is(a"},
			code:   2,
			stderr: "selwhy: expected ) at column 6\n",
		},
		{
			name:   "nested has pseudo-class",
			argv:   []string{"fmt", ":has(:has(a))"},
			code:   2,
			stderr: "selwhy: :has() cannot be nested in :has() at column 6\n",
		},
		{
			name:   "selector nesting limit",
			argv:   []string{"fmt", ":not(:not(:not(:not(:not(:not(:not(:not(:not(a)))))))))"},
			code:   2,
			stderr: "selwhy: nesting is deeper than 8 at column 41\n",
		},
		{
			name:   "sheet selector error",
			argv:   []string{"cascade", "a:foo{x:y}", "<a/>"},
			code:   2,
			stderr: "selwhy: rule 1: unknown pseudo-class \":foo\" at column 2\n",
		},
		{
			name:   "declaration without colon",
			argv:   []string{"cascade", "a{x:1;y}", "<a/>"},
			code:   2,
			stderr: "selwhy: rule 1: declaration 2 has no colon\n",
		},
		{
			name:   "invalid property name",
			argv:   []string{"cascade", "a{1x:2}", "<a/>"},
			code:   2,
			stderr: "selwhy: rule 1: invalid property name \"1x\"\n",
		},
		{
			name:   "property without value",
			argv:   []string{"cascade", "a{x: !important}", "<a/>"},
			code:   2,
			stderr: "selwhy: rule 1: property x has no value\n",
		},
		{
			name:   "cascade '/* x' '<a/>'",
			argv:   []string{"cascade", "/* x", "<a/>"},
			code:   2,
			stderr: "selwhy: unterminated comment starting at line 1 column 1\n",
		},
		{
			name:   "nth unknown option takes precedence over invalid count",
			argv:   []string{"nth", "--cnt", "3", "--count", "0"},
			code:   2,
			stderr: "selwhy: unknown option \"--cnt\"\n",
		},
		{
			name:   "nth invalid count precedes missing formula",
			argv:   []string{"nth", "--count", "0"},
			code:   2,
			stderr: "selwhy: invalid --count \"0\" (expected a whole number from 1 to 100)\n",
		},
		{
			name:   "tree missing document",
			argv:   []string{"tree"},
			code:   2,
			stderr: "selwhy: missing document\n",
		},
		{
			name:   "select missing selector",
			argv:   []string{"select"},
			code:   2,
			stderr: "selwhy: missing selector\n",
		},
		{
			name:   "select missing document",
			argv:   []string{"select", "a"},
			code:   2,
			stderr: "selwhy: missing document\n",
		},
		{
			name:   "explain missing element number",
			argv:   []string{"explain", "a", "<a/>"},
			code:   2,
			stderr: "selwhy: missing element number\n",
		},
		{
			name:   "specificity missing selector",
			argv:   []string{"specificity"},
			code:   2,
			stderr: "selwhy: missing selector\n",
		},
		{
			name:   "fmt missing selector",
			argv:   []string{"fmt"},
			code:   2,
			stderr: "selwhy: missing selector\n",
		},
		{
			name:   "nth missing formula",
			argv:   []string{"nth"},
			code:   2,
			stderr: "selwhy: missing formula\n",
		},
		{
			name:   "cascade missing style sheet",
			argv:   []string{"cascade"},
			code:   2,
			stderr: "selwhy: missing style sheet\n",
		},
		{
			name:   "cascade missing document",
			argv:   []string{"cascade", "a{x:y}"},
			code:   2,
			stderr: "selwhy: missing document\n",
		},
		{
			name:   "tree arity precedes stdin read",
			argv:   []string{"tree", "-", "extra"},
			stdin:  []byte("<a/>"),
			code:   2,
			stderr: "selwhy: too many arguments\n",
		},
		{
			name:   "select arity precedes duplicate stdin check",
			argv:   []string{"select", "-", "-", "extra"},
			stdin:  []byte("<a/>"),
			code:   2,
			stderr: "selwhy: too many arguments\n",
		},
		{
			name:   "selector is processed before document",
			argv:   []string{"select", "a%", "<"},
			code:   2,
			stderr: "selwhy: unexpected \"%\" at column 2\n",
		},
		{
			name:   "document is processed before element number",
			argv:   []string{"explain", "a", "<", "0"},
			code:   2,
			stderr: "selwhy: malformed tag at line 1 column 1\n",
		},
		{
			name:   "style sheet is processed before document",
			argv:   []string{"cascade", "a{x:}", "<"},
			code:   2,
			stderr: "selwhy: rule 1: property x has no value\n",
		},
		{
			name:   "last count option wins",
			argv:   []string{"nth", "--count", "0", "--count", "5", "-n+3"},
			code:   0,
			stdout: "formula     [-n+3]\na           -1\nb           3\nmatches     1 2 3\n",
		},
		{
			name:   "option terminator makes later option text formula",
			argv:   []string{"nth", "--", "--count"},
			code:   2,
			stderr: "selwhy: invalid an+b formula \"--count\"\n",
		},
		{
			name:   "fmt stdin strips one BOM and CRLF",
			argv:   []string{"fmt", "-"},
			stdin:  []byte("\xef\xbb\xbf" + "a > b\r\n"),
			reads:  1,
			code:   0,
			stdout: "a > b\n",
		},
		{
			name:   "tree stdin parses document",
			argv:   []string{"tree", "-"},
			stdin:  []byte("<a/>\n"),
			reads:  1,
			code:   0,
			stdout: "n  path   tag\n1  /a[1]  [<a>]\nelements    1\n",
		},
		{
			name:   "select stdin parses document",
			argv:   []string{"select", "a", "-"},
			stdin:  []byte("<a/>"),
			reads:  1,
			code:   0,
			stdout: "n  path   tag\n1  /a[1]  [<a>]\nmatches     1\n",
		},
		{
			name:   "nth stdin parses formula",
			argv:   []string{"nth", "--count", "3", "-"},
			stdin:  []byte("even\n"),
			reads:  1,
			code:   0,
			stdout: "formula     [2n]\na           2\nb           0\nmatches     2\n",
		},
		{
			name:   "fmt replaces non-ASCII output",
			argv:   []string{"fmt", "[a=\"é\"]"},
			code:   0,
			stdout: "[a=\"?\"]\n",
		},
		{
			name:   "select with no matches",
			argv:   []string{"select", "b", "<a/>"},
			code:   1,
			stdout: "n  path  tag\nmatches     0\n",
		},
		{
			name:   "nth formula has no matching positions",
			argv:   []string{"nth", "--count", "3", "0"},
			code:   0,
			stdout: "formula     [0]\na           0\nb           0\nmatches     -\n",
		},
		{
			name:   "cascade with no styled elements",
			argv:   []string{"cascade", "b{x:y}", "<a/>"},
			code:   0,
			stdout: "n  path  property  value  from\nrules       1\nstyled      0\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			io := &commandTestIO{stdin: test.stdin}
			if got := Run(test.argv, io); got != test.code {
				t.Fatalf("Run(%q) = %d, want %d", test.argv, got, test.code)
			}
			if test.stdout == "" {
				if len(io.outputs) != 0 {
					t.Fatalf("Run(%q) made %d Out calls, want none: %#v", test.argv, len(io.outputs), io.outputs)
				}
			} else {
				if len(io.outputs) != 1 {
					t.Fatalf("Run(%q) made %d Out calls, want one", test.argv, len(io.outputs))
				}
				if io.outputs[0] != test.stdout {
					t.Fatalf("Run(%q) stdout = %q, want %q", test.argv, io.outputs[0], test.stdout)
				}
			}
			if test.stderr == "" {
				if len(io.errors) != 0 {
					t.Fatalf("Run(%q) made %d Err calls, want none: %#v", test.argv, len(io.errors), io.errors)
				}
			} else {
				if len(io.errors) != 1 {
					t.Fatalf("Run(%q) made %d Err calls, want one", test.argv, len(io.errors))
				}
				if io.errors[0] != test.stderr {
					t.Fatalf("Run(%q) stderr = %q, want %q", test.argv, io.errors[0], test.stderr)
				}
			}
			if io.reads != test.reads {
				t.Fatalf("Run(%q) made %d stdin reads, want %d", test.argv, io.reads, test.reads)
			}
		})
	}
}
