package base

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// IO is the in-memory boundary for command input and output.
type IO interface {
	ReadStdin() []byte
	Out(text string)
	Err(text string)
}

// TextArg returns arg unchanged unless it denotes stdin, in which case it
// strictly decodes the input and removes one BOM and one final line ending.
func TextArg(io IO, arg string) (string, error) {
	if arg != "-" {
		return arg, nil
	}

	input := io.ReadStdin()
	if !utf8.Valid(input) {
		return "", errors.New("<stdin>: input is not valid UTF-8")
	}

	text := string(input)
	text = strings.TrimPrefix(text, "\uFEFF")
	if strings.HasSuffix(text, "\r\n") {
		text = text[:len(text)-2]
	} else if strings.HasSuffix(text, "\n") {
		text = text[:len(text)-1]
	}
	return text, nil
}
