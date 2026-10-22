<p align="center"><img src="docs/logo.svg" width="96" alt="" /></p>

# selwhy

selwhy is an offline command-line tool for web developers and testers who write CSS selectors, style sheets, and test locators. It helps explain why a selector matches or does not match a small markup document, inspect selector specificity and canonical form, and see which style declaration wins. It runs locally and does not load pages or files.

[![CI](https://github.com/Gulliby-development/selwhy/actions/workflows/ci.yml/badge.svg)](https://github.com/Gulliby-development/selwhy/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Install

Install a recent Go toolchain, then build the executable from the project directory:

```sh
go build -o selwhy .
```

On Windows, choose an executable name such as `selwhy.exe`. You can also run the program without keeping a built binary:

```sh
go run . --help
```

## Usage

The command forms are:

```text
selwhy tree [--depth N] DOC
selwhy select SELECTOR DOC
selwhy explain SELECTOR DOC N
selwhy specificity SELECTOR
selwhy fmt SELECTOR
selwhy nth [--count N] FORMULA
selwhy cascade SHEET DOC
```

Use `-` in place of a text argument to read that argument from standard input. Each command can read at most one `-`. Options are parsed before positional arguments; `--` ends option parsing. For `nth`, repeated `--count` options use the last value.

In the examples, "$DOC1" is this document:

```html
<div id="app" class="page wide"><ul class="nav"><li class="on">Home</li><li data-k="b-2">Docs &amp; FAQ</li><li></li></ul><p lang="en">Hi</p></div>
```

## Commands

- `tree [--depth N] DOC` lists elements in document order with their paths and opening tags; `--depth N` (1 to 12) keeps only the elements at most N levels deep, the root being level 1.
- `select SELECTOR DOC` lists elements matched by a selector list.
- `explain SELECTOR DOC N` reports whether each selector matches the numbered element and gives a reason for each non-match.
- `specificity SELECTOR` lists each selector and its specificity.
- `fmt SELECTOR` prints the canonical selector form.
- `nth [--count N] FORMULA` shows the positions selected by an `an+b` formula (the default count is 12).
- `cascade SHEET DOC` lists the winning declaration for each styled element and property.

The tree example is `selwhy tree "$DOC1"`:

```text
n  path                 tag
1  /div[1]              [<div id="app" class="page wide">]
2  /div[1]/ul[1]        [<ul class="nav">]
3  /div[1]/ul[1]/li[1]  [<li class="on">]
4  /div[1]/ul[1]/li[2]  [<li data-k="b-2">]
5  /div[1]/ul[1]/li[3]  [<li>]
6  /div[1]/p[1]         [<p lang="en">]
elements    6
```

With `--depth`, the elements keep their numbers and the `elements` row counts the rows listed. `selwhy tree --depth 2 "$DOC1"` prints:

```text
n  path           tag
1  /div[1]        [<div id="app" class="page wide">]
2  /div[1]/ul[1]  [<ul class="nav">]
6  /div[1]/p[1]   [<p lang="en">]
elements    3
```

The first selection example is `selwhy select 'ul > li:not(.on)' "$DOC1"`:

```text
n  path                 tag
4  /div[1]/ul[1]/li[2]  [<li data-k="b-2">]
5  /div[1]/ul[1]/li[3]  [<li>]
matches     2
```

The cascade example is `selwhy cascade 'li{margin:0} /* menu */ ul > li.on{--x:2;margin:1} #app li.on{--x:3} .on{margin:4 !important}' "$DOC1"`:

```text
n  path                 property  value  from
3  /div[1]/ul[1]/li[1]  --x       [3]    rule 3 (1,1,1)
3  /div[1]/ul[1]/li[1]  margin    [4]    rule 4 (0,1,0) !important
4  /div[1]/ul[1]/li[2]  margin    [0]    rule 1 (0,0,1)
5  /div[1]/ul[1]/li[3]  margin    [0]    rule 1 (0,0,1)
rules       4
styled      3
```

## Supported syntax and limitations

selwhy works on a small markup string supplied as an argument or through standard input; it does not fetch a page or read application files. It supports one root element, nested elements, attributes, text and comments. Its selector subset includes type, universal, ID, class and attribute selectors; descendant, child and sibling combinators; structural pseudo-classes; `:not()`, `:is()`, `:where()`, `:has()`; and the four `nth-*` formula pseudo-classes. Style sheets support selector lists, declarations, comments and `!important` for cascade resolution.

This is a focused inspection tool, not a browser or a complete HTML/CSS implementation. It does not apply browser parsing or layout, load external resources, or implement selector and style-sheet features outside the documented subset.

## Exit codes

- `0`: the command completed successfully.
- `1`: `select` found no matches, or `explain` found that none of the selectors matched the selected element.
- `2`: an input, option or command error occurred; a diagnostic is written to standard error.

## Development

The project uses only the Go standard library. From the repository root, run the formatting, static analysis and test checks:

```sh
gofmt -l .
go vet ./...
go test ./...
```

## License

MIT, see [LICENSE](LICENSE).

---

Built with BAKED
