package base

import (
	"reflect"
	"testing"
)

func TestShowAndCell(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "empty show", got: Show(""), want: "\"\""},
		{name: "printable endpoints", got: Show(" A~"), want: "\" A~\""},
		{name: "replacement", got: Show("é🙂\n\t\x7f"), want: "\"?????\""},
		{name: "no escaping", got: Show("a\"\\b"), want: "\"a\"\\b\""},
		{name: "empty cell", got: Cell(""), want: "[]"},
		{name: "printable brackets", got: Cell("a[b]"), want: "[a[b]]"},
		{name: "cell replacement", got: Cell("é\r"), want: "[??]"},
		{name: "invalid decoding unit", got: Show("\xff"), want: "\"?\""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("got %q, want %q", test.got, test.want)
			}
		})
	}
}

func TestLine(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		want  string
	}{
		{name: "short key", key: "a", value: "b", want: "a           b"},
		{name: "full-width key", key: "0123456789", value: "v", want: "0123456789  v"},
		{name: "long key", key: "01234567890", value: "", want: "01234567890  "},
		{name: "code-point width", key: "é", value: "x", want: "é           x"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Line(test.key, test.value); got != test.want {
				t.Fatalf("Line(%q, %q) = %q, want %q", test.key, test.value, got, test.want)
			}
		})
	}
}

func TestTable(t *testing.T) {
	tests := []struct {
		name   string
		header []string
		rows   [][]string
		want   []string
	}{
		{
			name:   "row-driven widths",
			header: []string{"n", "tag"},
			rows:   [][]string{{"1", "a"}, {"12", "long"}},
			want:   []string{"n   tag", "1   a", "12  long"},
		},
		{
			name:   "header-driven width",
			header: []string{"name", "v"},
			rows:   [][]string{{"x", "y"}},
			want:   []string{"name  v", "x     y"},
		},
		{
			name:   "header only",
			header: []string{"n", "tag"},
			rows:   nil,
			want:   []string{"n  tag"},
		},
		{name: "empty input", header: nil, rows: nil, want: []string{""}},
		{
			name:   "ragged rows",
			header: []string{"a"},
			rows:   [][]string{{"b", "c"}, {}},
			want:   []string{"a", "b  c", ""},
		},
		{
			name:   "Unicode width and trimming",
			header: []string{"é", "x"},
			rows:   [][]string{{"ab", ""}},
			want:   []string{"é   x", "ab"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Table(test.header, test.rows); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Table(%q, %q) = %#v, want %#v", test.header, test.rows, got, test.want)
			}
		})
	}
}

func TestParseOptions(t *testing.T) {
	specs := []Spec{{Name: "--count"}}
	tests := []struct {
		name string
		argv []string
		want Parsed
	}{
		{
			name: "empty",
			argv: []string{},
			want: Parsed{Values: map[string][]string{}, Args: []string{}},
		},
		{
			name: "interleaved repetition",
			argv: []string{"x", "--count", "2", "y", "--count=3"},
			want: Parsed{
				Values: map[string][]string{"--count": []string{"2", "3"}},
				Args:   []string{"x", "y"},
			},
		},
		{
			name: "first equals",
			argv: []string{"--count=a=b"},
			want: Parsed{
				Values: map[string][]string{"--count": []string{"a=b"}},
				Args:   []string{},
			},
		},
		{
			name: "empty value",
			argv: []string{"--count="},
			want: Parsed{
				Values: map[string][]string{"--count": []string{""}},
				Args:   []string{},
			},
		},
		{
			name: "option-looking value",
			argv: []string{"--count", "--other"},
			want: Parsed{
				Values: map[string][]string{"--count": []string{"--other"}},
				Args:   []string{},
			},
		},
		{
			name: "consumed terminator",
			argv: []string{"--count", "--", "x"},
			want: Parsed{
				Values: map[string][]string{"--count": []string{"--"}},
				Args:   []string{"x"},
			},
		},
		{
			name: "ordinary arguments and terminator",
			argv: []string{"-", "-x", "--", "--count", "--", "z"},
			want: Parsed{
				Values: map[string][]string{},
				Args:   []string{"-", "-x", "--count", "--", "z"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Parse(test.argv, specs)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", test.argv, err)
			}
			if got.Values == nil {
				t.Fatal("Parse returned a nil Values map")
			}
			if !reflect.DeepEqual(got.Values, test.want.Values) {
				t.Fatalf("Parse(%q) Values = %#v, want %#v", test.argv, got.Values, test.want.Values)
			}
			if len(got.Args) != len(test.want.Args) {
				t.Fatalf("Parse(%q) Args = %#v, want %#v", test.argv, got.Args, test.want.Args)
			}
			for i, want := range test.want.Args {
				if got.Args[i] != want {
					t.Fatalf("Parse(%q) Args = %#v, want %#v", test.argv, got.Args, test.want.Args)
				}
			}
		})
	}
}

func TestParseOptionErrors(t *testing.T) {
	specs := []Spec{{Name: "--count"}}
	tests := []struct {
		name string
		argv []string
		want string
	}{
		{name: "unknown name", argv: []string{"--cnt", "3"}, want: "unknown option \"--cnt\""},
		{name: "full unknown argument", argv: []string{"--cnt=3"}, want: "unknown option \"--cnt=3\""},
		{name: "ASCII diagnostic", argv: []string{"--é=1"}, want: "unknown option \"--?=1\""},
		{name: "missing value", argv: []string{"2n", "--count"}, want: "--count needs a value"},
		{name: "first error", argv: []string{"--bad", "--count"}, want: "unknown option \"--bad\""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.argv, specs)
			if err == nil {
				t.Fatalf("Parse(%q) returned no error, want %q", test.argv, test.want)
			}
			if got := err.Error(); got != test.want {
				t.Fatalf("Parse(%q) error = %q, want %q", test.argv, got, test.want)
			}
		})
	}
}

type fakeTextArgIO struct {
	stdin    []byte
	reads    int
	outCalls int
	errCalls int
}

func (io *fakeTextArgIO) ReadStdin() []byte {
	io.reads++
	return io.stdin
}

func (io *fakeTextArgIO) Out(string) {
	io.outCalls++
}

func (io *fakeTextArgIO) Err(string) {
	io.errCalls++
}

func TestTextArg(t *testing.T) {
	tests := []struct {
		name  string
		arg   string
		stdin []byte
		want  string
		err   string
		reads int
	}{
		{name: "direct argument", arg: "é\n", stdin: []byte("ignored"), want: "é\n"},
		{name: "empty argument", arg: "", stdin: []byte("ignored"), want: ""},
		{name: "invalid direct bytes unchanged", arg: "\xff", stdin: []byte("ignored"), want: "\xff"},
		{name: "empty stdin", arg: "-", stdin: []byte(""), want: "", reads: 1},
		{name: "valid UTF-8", arg: "-", stdin: []byte("é🙂"), want: "é🙂", reads: 1},
		{name: "BOM and CRLF", arg: "-", stdin: []byte("\ufeffabc\r\n"), want: "abc", reads: 1},
		{name: "one leading BOM", arg: "-", stdin: []byte("\ufeff\ufeffx"), want: "\ufeffx", reads: 1},
		{name: "one final LF", arg: "-", stdin: []byte("a\nb\n\n"), want: "a\nb\n", reads: 1},
		{name: "one final CRLF", arg: "-", stdin: []byte("x\r\n\r\n"), want: "x\r\n", reads: 1},
		{name: "lone CR", arg: "-", stdin: []byte("x\r"), want: "x\r", reads: 1},
		{name: "interior BOM", arg: "-", stdin: []byte("x\ufeffy"), want: "x\ufeffy", reads: 1},
		{name: "invalid byte", arg: "-", stdin: []byte("\x65\x63\xff\x0a"), err: "<stdin>: input is not valid UTF-8", reads: 1},
		{name: "overlong encoding", arg: "-", stdin: []byte("\xc0\xaf"), err: "<stdin>: input is not valid UTF-8", reads: 1},
		{name: "encoded surrogate", arg: "-", stdin: []byte("\xed\xa0\x80"), err: "<stdin>: input is not valid UTF-8", reads: 1},
		{name: "truncated encoding", arg: "-", stdin: []byte("\xe2\x82"), err: "<stdin>: input is not valid UTF-8", reads: 1},
		{name: "above Unicode range", arg: "-", stdin: []byte("\xf4\x90\x80\x80"), err: "<stdin>: input is not valid UTF-8", reads: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			io := &fakeTextArgIO{stdin: test.stdin}
			got, err := TextArg(io, test.arg)
			if test.err == "" {
				if err != nil {
					t.Fatalf("TextArg(%q) returned error: %v", test.arg, err)
				}
				if got != test.want {
					t.Fatalf("TextArg(%q) = %q, want %q", test.arg, got, test.want)
				}
			} else {
				if err == nil {
					t.Fatalf("TextArg(%q) returned no error, want %q", test.arg, test.err)
				}
				if got := err.Error(); got != test.err {
					t.Fatalf("TextArg(%q) error = %q, want %q", test.arg, got, test.err)
				}
			}
			if io.reads != test.reads {
				t.Fatalf("TextArg(%q) made %d reads, want %d", test.arg, io.reads, test.reads)
			}
			if io.outCalls != 0 || io.errCalls != 0 {
				t.Fatalf("TextArg(%q) wrote output: Out calls = %d, Err calls = %d", test.arg, io.outCalls, io.errCalls)
			}
		})
	}
}
