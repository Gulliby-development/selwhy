package cli

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Gulliby-development/selwhy/internal/base"
	"github.com/Gulliby-development/selwhy/internal/markup"
)

func runTree(args []string, io base.IO) (int, error) {
	parsed, err := base.Parse(args, nil)
	if err != nil {
		return 0, err
	}
	if len(parsed.Args) == 0 {
		return 0, errors.New("missing document")
	}
	if len(parsed.Args) > 1 {
		return 0, errors.New("too many arguments")
	}

	text, err := base.TextArg(io, parsed.Args[0])
	if err != nil {
		return 0, err
	}
	doc, err := markup.Parse(text)
	if err != nil {
		return 0, err
	}

	rows := make([][]string, 0, len(doc.Elements))
	for _, element := range doc.Elements {
		rows = append(rows, []string{
			strconv.Itoa(element.Num),
			markup.Path(element),
			base.Cell(markup.TagText(element)),
		})
	}

	lines := base.Table([]string{"n", "path", "tag"}, rows)
	lines = append(lines, base.Line("elements", strconv.Itoa(len(doc.Elements))))
	io.Out(strings.Join(lines, "\n") + "\n")
	return 0, nil
}
