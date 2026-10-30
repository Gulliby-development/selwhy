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

// runSelect parses a selector and document, then prints the selected elements.
func runSelect(args []string, io base.IO) (int, error) {
	parsed, err := base.Parse(args, []base.Spec{{Name: "--limit"}})
	if err != nil {
		return 0, err
	}

	limit := 0
	if values := parsed.Values["--limit"]; len(values) > 0 {
		limit, err = parseSelectLimit(values[len(values)-1])
		if err != nil {
			return 0, err
		}
	}

	switch len(parsed.Args) {
	case 0:
		return 0, errors.New("missing selector")
	case 1:
		return 0, errors.New("missing document")
	case 2:
	default:
		return 0, errors.New("too many arguments")
	}
	if parsed.Args[0] == "-" && parsed.Args[1] == "-" {
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

	selected := match.Select(selectors, document)
	listed := selected
	if limit > 0 && len(listed) > limit {
		listed = listed[:limit]
	}
	rows := make([][]string, 0, len(listed))
	for _, element := range listed {
		rows = append(rows, []string{
			strconv.Itoa(element.Num),
			markup.Path(element),
			base.Cell(markup.TagText(element)),
		})
	}

	var output strings.Builder
	for _, line := range base.Table([]string{"n", "path", "tag"}, rows) {
		output.WriteString(line)
		output.WriteByte('\n')
	}
	output.WriteString(base.Line("matches", strconv.Itoa(len(selected))))
	output.WriteByte('\n')
	if len(listed) < len(selected) {
		output.WriteString(base.Line("listed", strconv.Itoa(len(listed))))
		output.WriteByte('\n')
	}
	io.Out(output.String())

	if len(selected) == 0 {
		return 1, nil
	}
	return 0, nil
}

// parseSelectLimit validates the bounded ASCII limit syntax before conversion.
func parseSelectLimit(text string) (int, error) {
	valid := len(text) >= 1 && len(text) <= 4 && text[0] != '0'
	for i := 0; valid && i < len(text); i++ {
		valid = text[i] >= '0' && text[i] <= '9'
	}
	if valid {
		value, err := strconv.Atoi(text)
		if err == nil && value >= 1 && value <= 1000 {
			return value, nil
		}
	}
	return 0, errors.New("invalid --limit " + base.Show(text) + " (expected a whole number from 1 to 1000)")
}
