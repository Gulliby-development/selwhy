package cli

import "testing"

func TestSelectLimit(t *testing.T) {
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
			name:   "no option is the same as before",
			argv:   []string{"select", "ul > li:not(.on)", commandTestDOC1},
			stdout: "n  path                 tag\n4  /div[1]/ul[1]/li[2]  [<li data-k=\"b-2\">]\n5  /div[1]/ul[1]/li[3]  [<li>]\nmatches     2\n",
		},
		{
			name:   "limit 1 lists the first match and counts them all",
			argv:   []string{"select", "--limit", "1", "ul > li:not(.on)", commandTestDOC1},
			stdout: "n  path                 tag\n4  /div[1]/ul[1]/li[2]  [<li data-k=\"b-2\">]\nmatches     2\nlisted      1\n",
		},
		{
			name:   "equals form",
			argv:   []string{"select", "--limit=1", "li", commandTestDOC1},
			stdout: "n  path                 tag\n3  /div[1]/ul[1]/li[1]  [<li class=\"on\">]\nmatches     3\nlisted      1\n",
		},
		{
			name:   "a limit equal to the match count adds no row",
			argv:   []string{"select", "--limit", "2", "ul > li:not(.on)", commandTestDOC1},
			stdout: "n  path                 tag\n4  /div[1]/ul[1]/li[2]  [<li data-k=\"b-2\">]\n5  /div[1]/ul[1]/li[3]  [<li>]\nmatches     2\n",
		},
		{
			name:   "the largest limit lists everything",
			argv:   []string{"select", "--limit", "1000", "li", commandTestDOC1},
			stdout: "n  path                 tag\n3  /div[1]/ul[1]/li[1]  [<li class=\"on\">]\n4  /div[1]/ul[1]/li[2]  [<li data-k=\"b-2\">]\n5  /div[1]/ul[1]/li[3]  [<li>]\nmatches     3\n",
		},
		{
			name:   "the last repeated limit wins",
			argv:   []string{"select", "--limit", "1", "--limit", "2", "li", commandTestDOC1},
			stdout: "n  path                 tag\n3  /div[1]/ul[1]/li[1]  [<li class=\"on\">]\n4  /div[1]/ul[1]/li[2]  [<li data-k=\"b-2\">]\nmatches     3\nlisted      2\n",
		},
		{
			name:   "the option may follow the arguments",
			argv:   []string{"select", "li", commandTestDOC1, "--limit", "2"},
			stdout: "n  path                 tag\n3  /div[1]/ul[1]/li[1]  [<li class=\"on\">]\n4  /div[1]/ul[1]/li[2]  [<li data-k=\"b-2\">]\nmatches     3\nlisted      2\n",
		},
		{
			name:   "limit with a document on standard input",
			argv:   []string{"select", "--limit", "1", "li", "-"},
			stdin:  []byte("<a><li/><li/></a>\n"),
			reads:  1,
			stdout: "n  path         tag\n2  /a[1]/li[1]  [<li>]\nmatches     2\nlisted      1\n",
		},
		{
			name:   "no match with a limit still exits 1",
			argv:   []string{"select", "--limit", "1", "h1", commandTestDOC1},
			stdout: "n  path  tag\nmatches     0\n",
			code:   1,
		},
		{
			name:   "document order across nesting",
			argv:   []string{"select", "--limit", "1", "b", "<r><a><b/></a><b/></r>"},
			stdout: "n  path             tag\n3  /r[1]/a[1]/b[1]  [<b>]\nmatches     2\nlisted      1\n",
		},
		{
			name:   "zero",
			argv:   []string{"select", "--limit", "0", "li", commandTestDOC1},
			stderr: "selwhy: invalid --limit \"0\" (expected a whole number from 1 to 1000)\n",
			code:   2,
		},
		{
			name:   "one thousand and one",
			argv:   []string{"select", "--limit", "1001", "li", commandTestDOC1},
			stderr: "selwhy: invalid --limit \"1001\" (expected a whole number from 1 to 1000)\n",
			code:   2,
		},
		{
			name:   "five digits",
			argv:   []string{"select", "--limit", "00001", "li", commandTestDOC1},
			stderr: "selwhy: invalid --limit \"00001\" (expected a whole number from 1 to 1000)\n",
			code:   2,
		},
		{
			name:   "leading zero",
			argv:   []string{"select", "--limit", "01", "li", commandTestDOC1},
			stderr: "selwhy: invalid --limit \"01\" (expected a whole number from 1 to 1000)\n",
			code:   2,
		},
		{
			name:   "decimal point",
			argv:   []string{"select", "--limit", "1.5", "li", commandTestDOC1},
			stderr: "selwhy: invalid --limit \"1.5\" (expected a whole number from 1 to 1000)\n",
			code:   2,
		},
		{
			name:   "empty value",
			argv:   []string{"select", "--limit", "", "li", commandTestDOC1},
			stderr: "selwhy: invalid --limit \"\" (expected a whole number from 1 to 1000)\n",
			code:   2,
		},
		{
			name:   "missing value",
			argv:   []string{"select", "li", commandTestDOC1, "--limit"},
			stderr: "selwhy: --limit needs a value\n",
			code:   2,
		},
		{
			name:   "unknown option comes first",
			argv:   []string{"select", "--lim", "0", "li", commandTestDOC1},
			stderr: "selwhy: unknown option \"--lim\"\n",
			code:   2,
		},
		{
			name:   "invalid limit precedes the missing selector",
			argv:   []string{"select", "--limit", "0"},
			stderr: "selwhy: invalid --limit \"0\" (expected a whole number from 1 to 1000)\n",
			code:   2,
		},
		{
			name:   "valid limit with no selector",
			argv:   []string{"select", "--limit", "2"},
			stderr: "selwhy: missing selector\n",
			code:   2,
		},
		{
			name:   "tree still rejects the option",
			argv:   []string{"tree", "--limit", "1", commandTestDOC1},
			stderr: "selwhy: unknown option \"--limit\"\n",
			code:   2,
		},
		{
			name:   "explain still rejects the option",
			argv:   []string{"explain", "--limit", "1", "li", commandTestDOC1, "1"},
			stderr: "selwhy: unknown option \"--limit\"\n",
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
