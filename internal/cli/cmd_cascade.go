package cli

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Gulliby-development/selwhy/internal/base"
	"github.com/Gulliby-development/selwhy/internal/markup"
	"github.com/Gulliby-development/selwhy/internal/sheet"
)

func runCascade(args []string, io base.IO) (int, error) {
	parsed, err := base.Parse(args, nil)
	if err != nil {
		return 0, err
	}

	if len(parsed.Args) == 0 {
		return 0, errors.New("missing style sheet")
	}
	if len(parsed.Args) == 1 {
		return 0, errors.New("missing document")
	}
	if len(parsed.Args) > 2 {
		return 0, errors.New("too many arguments")
	}
	if parsed.Args[0] == "-" && parsed.Args[1] == "-" {
		return 0, errors.New("only one argument can be -")
	}

	sheetText, err := base.TextArg(io, parsed.Args[0])
	if err != nil {
		return 0, err
	}
	rules, err := sheet.Parse(sheetText)
	if err != nil {
		return 0, err
	}

	documentText, err := base.TextArg(io, parsed.Args[1])
	if err != nil {
		return 0, err
	}
	doc, err := markup.Parse(documentText)
	if err != nil {
		return 0, err
	}

	rows := make([][]string, 0)
	styled := 0
	for _, element := range doc.Elements {
		winners := sheet.Cascade(rules, element)
		if len(winners) > 0 {
			styled++
		}
		for _, winner := range winners {
			from := "rule " + strconv.Itoa(winner.Rule) + " (" + winner.Weight.String() + ")"
			if winner.Important {
				from += " !important"
			}
			rows = append(rows, []string{
				strconv.Itoa(element.Num),
				markup.Path(element),
				winner.Prop,
				base.Cell(winner.Value),
				from,
			})
		}
	}

	lines := base.Table(
		[]string{"n", "path", "property", "value", "from"},
		rows,
	)
	lines = append(lines, base.Line("rules", strconv.Itoa(len(rules))))
	lines = append(lines, base.Line("styled", strconv.Itoa(styled)))
	io.Out(strings.Join(lines, "\n") + "\n")
	return 0, nil
}
