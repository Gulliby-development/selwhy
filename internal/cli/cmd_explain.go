package cli

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Gulliby-development/selwhy/internal/base"
	"github.com/Gulliby-development/selwhy/internal/markup"
	"github.com/Gulliby-development/selwhy/internal/match"
	"github.com/Gulliby-development/selwhy/internal/selector"
)

// runExplain explains whether each selector matches the requested element.
func runExplain(args []string, io base.IO) (int, error) {
	parsed, err := base.Parse(args, nil)
	if err != nil {
		return 0, err
	}

	switch len(parsed.Args) {
	case 0:
		return 0, errors.New("missing selector")
	case 1:
		return 0, errors.New("missing document")
	case 2:
		return 0, errors.New("missing element number")
	case 3:
	default:
		return 0, errors.New("too many arguments")
	}
	stdinCount := 0
	for _, arg := range parsed.Args {
		if arg == "-" {
			stdinCount++
		}
	}
	if stdinCount > 1 {
		return 0, errors.New("only one argument can be -")
	}

	selectorText, err := base.TextArg(io, parsed.Args[0])
	if err != nil {
		return 0, err
	}
	selectors, err := selector.Parse(selectorText)
	if err != nil {
		return 0, err
	}

	documentText, err := base.TextArg(io, parsed.Args[1])
	if err != nil {
		return 0, err
	}
	document, err := markup.Parse(documentText)
	if err != nil {
		return 0, err
	}

	elementNumber, ok := parseElementNumber(parsed.Args[2])
	if !ok {
		return 0, errors.New("invalid element number " + base.Show(parsed.Args[2]) +
			" (expected a whole number from 1 to 999999)")
	}
	if elementNumber > len(document.Elements) {
		return 0, errors.New("element " + strconv.Itoa(elementNumber) +
			" does not exist (the document has " + strconv.Itoa(len(document.Elements)) + " elements)")
	}
	element := document.Elements[elementNumber-1]

	var output strings.Builder
	output.WriteString(base.Line("element", strconv.Itoa(elementNumber)+" "+markup.Path(element)))
	output.WriteByte('\n')
	output.WriteString(base.Line("tag", base.Cell(markup.TagText(element))))
	output.WriteByte('\n')

	anyMatch := false
	for _, complex := range selectors {
		matched := match.ComplexMatches(complex, element)
		if matched {
			anyMatch = true
		}

		output.WriteString(base.Line("selector", base.Cell(selector.ComplexText(complex))))
		output.WriteByte('\n')
		result := "no match"
		if matched {
			result = "match"
		}
		output.WriteString(base.Line("result", result))
		output.WriteByte('\n')
		if !matched {
			output.WriteString(base.Line("reason", match.Reason(complex, element)))
			output.WriteByte('\n')
		}
	}
	io.Out(output.String())

	if !anyMatch {
		return 1, nil
	}
	return 0, nil
}

// parseElementNumber accepts a positive, bounded ASCII decimal number.
func parseElementNumber(text string) (int, bool) {
	if len(text) < 1 || len(text) > 6 {
		return 0, false
	}
	for i := 0; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return 0, false
		}
	}
	number, err := strconv.Atoi(text)
	if err != nil || number < 1 || number > 999999 {
		return 0, false
	}
	return number, true
}
