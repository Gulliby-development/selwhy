package cli

import (
	"os"
	"strings"
	"testing"
)

func TestREADMERequiredContent(t *testing.T) {
	content, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := strings.ReplaceAll(string(content), "\r\n", "\n")

	checks := []struct {
		name string
		want string
	}{
		{name: "install heading", want: "## Install"},
		{name: "usage heading", want: "## Usage"},
		{name: "commands heading", want: "## Commands"},
		{name: "exit codes heading", want: "## Exit codes"},
		{name: "development heading", want: "## Development"},
		{
			name: "purpose and audience",
			want: "selwhy is an offline command-line tool for web developers and testers who write CSS selectors, style sheets, and test locators.",
		},
		{
			name: "installation guidance",
			want: "## Install\n\nInstall a recent Go toolchain, then build the executable from the project directory:\n\n```sh\ngo build -o selwhy .\n```\n\nOn Windows, choose an executable name such as `selwhy.exe`. You can also run the program without keeping a built binary:\n\n```sh\ngo run . --help\n```",
		},
		{
			name: "ordered usage block",
			want: "The command forms are:\n\n```text\nselwhy tree [--depth N] DOC\nselwhy select [--limit N] SELECTOR DOC\nselwhy explain SELECTOR DOC N\nselwhy specificity SELECTOR\nselwhy fmt SELECTOR\nselwhy nth [--count N] FORMULA\nselwhy cascade SHEET DOC\n```",
		},
		{
			name: "standard input and options",
			want: "Use `-` in place of a text argument to read that argument from standard input. Each command can read at most one `-`. Options are parsed before positional arguments; `--` ends option parsing. For `nth`, repeated `--count` options use the last value.",
		},
		{
			name: "DOC1 definition",
			want: "In the examples, \"$DOC1\" is this document:\n\n```html\n<div id=\"app\" class=\"page wide\"><ul class=\"nav\"><li class=\"on\">Home</li><li data-k=\"b-2\">Docs &amp; FAQ</li><li></li></ul><p lang=\"en\">Hi</p></div>\n```",
		},
		{
			name: "tree example",
			want: "The tree example is `selwhy tree \"$DOC1\"`:\n\n```text\n" + `n  path                 tag
1  /div[1]              [<div id="app" class="page wide">]
2  /div[1]/ul[1]        [<ul class="nav">]
3  /div[1]/ul[1]/li[1]  [<li class="on">]
4  /div[1]/ul[1]/li[2]  [<li data-k="b-2">]
5  /div[1]/ul[1]/li[3]  [<li>]
6  /div[1]/p[1]         [<p lang="en">]
elements    6
` + "```",
		},
		{
			name: "first select example",
			want: "The first selection example is `selwhy select 'ul > li:not(.on)' \"$DOC1\"`:\n\n```text\n" + `n  path                 tag
4  /div[1]/ul[1]/li[2]  [<li data-k="b-2">]
5  /div[1]/ul[1]/li[3]  [<li>]
matches     2
` + "```",
		},
		{
			name: "cascade example",
			want: "The cascade example is `selwhy cascade 'li{margin:0} /* menu */ ul > li.on{--x:2;margin:1} #app li.on{--x:3} .on{margin:4 !important}' \"$DOC1\"`:\n\n```text\n" + `n  path                 property  value  from
3  /div[1]/ul[1]/li[1]  --x       [3]    rule 3 (1,1,1)
3  /div[1]/ul[1]/li[1]  margin    [4]    rule 4 (0,1,0) !important
4  /div[1]/ul[1]/li[2]  margin    [0]    rule 1 (0,0,1)
5  /div[1]/ul[1]/li[3]  margin    [0]    rule 1 (0,0,1)
rules       4
styled      3
` + "```",
		},
		{
			name: "supported subset",
			want: "selwhy works on a small markup string supplied as an argument or through standard input; it does not fetch a page or read application files. It supports one root element, nested elements, attributes, text and comments. Its selector subset includes type, universal, ID, class and attribute selectors; descendant, child and sibling combinators; structural pseudo-classes; `:not()`, `:is()`, `:where()`, `:has()`; and the four `nth-*` formula pseudo-classes. Style sheets support selector lists, declarations, comments and `!important` for cascade resolution.",
		},
		{
			name: "limitations",
			want: "This is a focused inspection tool, not a browser or a complete HTML/CSS implementation. It does not apply browser parsing or layout, load external resources, or implement selector and style-sheet features outside the documented subset.",
		},
		{
			name: "exit codes",
			want: "## Exit codes\n\n- `0`: the command completed successfully.\n- `1`: `select` found no matches, or `explain` found that none of the selectors matched the selected element.\n- `2`: an input, option or command error occurred; a diagnostic is written to standard error.",
		},
		{
			name: "development checks",
			want: "From the repository root, run the formatting, static analysis and test checks:\n\n```sh\ngofmt -l .\ngo vet ./...\ngo test ./...\n```",
		},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if !strings.Contains(readme, check.want) {
				t.Fatalf("README.md does not contain required %s", check.name)
			}
		})
	}
}
