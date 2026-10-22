package cli

import "testing"

func TestTreeDepth(t *testing.T) {
	const wantAll = "n  path                 tag\n" +
		"1  /div[1]              [<div id=\"app\" class=\"page wide\">]\n" +
		"2  /div[1]/ul[1]        [<ul class=\"nav\">]\n" +
		"3  /div[1]/ul[1]/li[1]  [<li class=\"on\">]\n" +
		"4  /div[1]/ul[1]/li[2]  [<li data-k=\"b-2\">]\n" +
		"5  /div[1]/ul[1]/li[3]  [<li>]\n" +
		"6  /div[1]/p[1]         [<p lang=\"en\">]\n" +
		"elements    6\n"
	const wantTwo = "n  path           tag\n" +
		"1  /div[1]        [<div id=\"app\" class=\"page wide\">]\n" +
		"2  /div[1]/ul[1]  [<ul class=\"nav\">]\n" +
		"6  /div[1]/p[1]   [<p lang=\"en\">]\n" +
		"elements    3\n"

	tests := []struct {
		name   string
		argv   []string
		stdin  []byte
		reads  int
		stdout string
		stderr string
		code   int
	}{
		{
			name:   "depth 1 keeps only the root",
			argv:   []string{"tree", "--depth", "1", commandTestDOC1},
			stdout: "n  path     tag\n1  /div[1]  [<div id=\"app\" class=\"page wide\">]\nelements    1\n",
		},
		{
			name:   "depth 2 keeps numbers and counts the rows listed",
			argv:   []string{"tree", "--depth", "2", commandTestDOC1},
			stdout: wantTwo,
		},
		{
			name:   "equals form and a depth that reaches every element",
			argv:   []string{"tree", "--depth=3", commandTestDOC1},
			stdout: wantAll,
		},
		{
			name:   "the largest depth lists everything",
			argv:   []string{"tree", commandTestDOC1, "--depth", "12"},
			stdout: wantAll,
		},
		{
			name:   "the last repeated depth wins",
			argv:   []string{"tree", "--depth", "1", "--depth", "2", commandTestDOC1},
			stdout: wantTwo,
		},
		{
			name:   "a deeper element is dropped but its later sibling stays",
			argv:   []string{"tree", "--depth", "3", "<r><a><b><c/></b></a><d/></r>"},
			stdout: "n  path             tag\n1  /r[1]            [<r>]\n2  /r[1]/a[1]       [<a>]\n3  /r[1]/a[1]/b[1]  [<b>]\n5  /r[1]/d[1]       [<d>]\nelements    4\n",
		},
		{
			name:   "depth with a document on standard input",
			argv:   []string{"tree", "--depth", "1", "-"},
			stdin:  []byte("<a><b/></a>\n"),
			reads:  1,
			stdout: "n  path   tag\n1  /a[1]  [<a>]\nelements    1\n",
		},
		{
			name:   "no option is the same as before",
			argv:   []string{"tree", commandTestDOC1},
			stdout: wantAll,
		},
		{
			name:   "zero",
			argv:   []string{"tree", "--depth", "0", commandTestDOC1},
			stderr: "selwhy: invalid --depth \"0\" (expected a whole number from 1 to 12)\n",
			code:   2,
		},
		{
			name:   "thirteen",
			argv:   []string{"tree", "--depth", "13", commandTestDOC1},
			stderr: "selwhy: invalid --depth \"13\" (expected a whole number from 1 to 12)\n",
			code:   2,
		},
		{
			name:   "leading zero",
			argv:   []string{"tree", "--depth", "01", commandTestDOC1},
			stderr: "selwhy: invalid --depth \"01\" (expected a whole number from 1 to 12)\n",
			code:   2,
		},
		{
			name:   "decimal point",
			argv:   []string{"tree", "--depth", "1.5", commandTestDOC1},
			stderr: "selwhy: invalid --depth \"1.5\" (expected a whole number from 1 to 12)\n",
			code:   2,
		},
		{
			name:   "minus sign",
			argv:   []string{"tree", "--depth", "-1", commandTestDOC1},
			stderr: "selwhy: invalid --depth \"-1\" (expected a whole number from 1 to 12)\n",
			code:   2,
		},
		{
			name:   "empty value",
			argv:   []string{"tree", "--depth", "", commandTestDOC1},
			stderr: "selwhy: invalid --depth \"\" (expected a whole number from 1 to 12)\n",
			code:   2,
		},
		{
			name:   "missing value",
			argv:   []string{"tree", commandTestDOC1, "--depth"},
			stderr: "selwhy: --depth needs a value\n",
			code:   2,
		},
		{
			name:   "unknown option comes first",
			argv:   []string{"tree", "--dep", "0", commandTestDOC1},
			stderr: "selwhy: unknown option \"--dep\"\n",
			code:   2,
		},
		{
			name:   "invalid depth precedes the missing document",
			argv:   []string{"tree", "--depth", "0"},
			stderr: "selwhy: invalid --depth \"0\" (expected a whole number from 1 to 12)\n",
			code:   2,
		},
		{
			name:   "valid depth with no document",
			argv:   []string{"tree", "--depth", "2"},
			stderr: "selwhy: missing document\n",
			code:   2,
		},
		{
			name:   "arity precedes the standard input read",
			argv:   []string{"tree", "--depth", "2", "-", "extra"},
			stderr: "selwhy: too many arguments\n",
			code:   2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			io := &commandTestIO{stdin: test.stdin}
			if got := Run(test.argv, io); got != test.code {
				t.Fatalf("Run(%q) = %d, want %d", test.argv, got, test.code)
			}
			wantOut := []string{}
			if test.stdout != "" {
				wantOut = []string{test.stdout}
			}
			if len(io.outputs) != len(wantOut) || (len(wantOut) == 1 && io.outputs[0] != wantOut[0]) {
				t.Fatalf("Run(%q) stdout = %q, want %q", test.argv, io.outputs, wantOut)
			}
			wantErr := []string{}
			if test.stderr != "" {
				wantErr = []string{test.stderr}
			}
			if len(io.errors) != len(wantErr) || (len(wantErr) == 1 && io.errors[0] != wantErr[0]) {
				t.Fatalf("Run(%q) stderr = %q, want %q", test.argv, io.errors, wantErr)
			}
			if io.reads != test.reads {
				t.Fatalf("Run(%q) made %d stdin reads, want %d", test.argv, io.reads, test.reads)
			}
		})
	}
}
