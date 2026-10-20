package cli

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Gulliby-development/selwhy/internal/base"
	"github.com/Gulliby-development/selwhy/internal/selector"
)

// runNth reports the positions selected by an an+b formula.
func runNth(args []string, io base.IO) (int, error) {
	parsed, err := base.Parse(args, []base.Spec{{Name: "--count"}})
	if err != nil {
		return 0, err
	}

	count := 12
	if values := parsed.Values["--count"]; len(values) > 0 {
		count, err = parseNthCount(values[len(values)-1])
		if err != nil {
			return 0, err
		}
	}

	if len(parsed.Args) == 0 {
		return 0, errors.New("missing formula")
	}
	if len(parsed.Args) > 1 {
		return 0, errors.New("too many arguments")
	}

	text, err := base.TextArg(io, parsed.Args[0])
	if err != nil {
		return 0, err
	}
	a, b, ok := selector.ParseNth(text)
	if !ok {
		trimmed := strings.Trim(text, " \t\r\n\f")
		return 0, errors.New("invalid an+b formula " + base.Show(trimmed))
	}

	matches := make([]string, 0, count)
	for pos := 1; pos <= count; pos++ {
		if selector.NthMatches(a, b, pos) {
			matches = append(matches, strconv.Itoa(pos))
		}
	}
	matchText := strings.Join(matches, " ")
	if matchText == "" {
		matchText = "-"
	}

	output := base.Line("formula", base.Cell(selector.NthText(a, b))) + "\n" +
		base.Line("a", strconv.Itoa(a)) + "\n" +
		base.Line("b", strconv.Itoa(b)) + "\n" +
		base.Line("matches", matchText) + "\n"
	io.Out(output)
	return 0, nil
}

// parseNthCount validates the bounded ASCII count syntax before conversion.
func parseNthCount(text string) (int, error) {
	valid := len(text) >= 1 && len(text) <= 3
	for i := 0; valid && i < len(text); i++ {
		valid = text[i] >= '0' && text[i] <= '9'
	}
	if valid {
		value, err := strconv.Atoi(text)
		if err == nil && value >= 1 && value <= 100 {
			return value, nil
		}
	}
	return 0, errors.New("invalid --count " + base.Show(text) + " (expected a whole number from 1 to 100)")
}
