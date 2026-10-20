package cli

import (
	"errors"
	"strings"

	"github.com/Gulliby-development/selwhy/internal/base"
	"github.com/Gulliby-development/selwhy/internal/selector"
)

// runSpecificity prints the canonical specificity table for a selector list.
func runSpecificity(args []string, io base.IO) (int, error) {
	parsed, err := base.Parse(args, nil)
	if err != nil {
		return 0, err
	}
	if len(parsed.Args) == 0 {
		return 0, errors.New("missing selector")
	}
	if len(parsed.Args) > 1 {
		return 0, errors.New("too many arguments")
	}

	text, err := base.TextArg(io, parsed.Args[0])
	if err != nil {
		return 0, err
	}
	list, err := selector.Parse(text)
	if err != nil {
		return 0, err
	}

	rows := make([][]string, 0, len(list))
	for _, complex := range list {
		rows = append(rows, []string{
			base.Cell(selector.ComplexText(complex)),
			selector.Weigh(complex).String(),
		})
	}
	lines := base.Table([]string{"selector", "specificity"}, rows)
	io.Out(strings.Join(lines, "\n") + "\n")
	return 0, nil
}
